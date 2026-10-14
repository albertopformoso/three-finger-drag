package gesture

import (
	"errors"
	"math"
	"slices"
	"testing"
	"time"
)

var accessible = Availability{Enabled: true, InputAvailable: true, OutputAvailable: true, LifecycleAvailable: true}

type harness struct {
	t     *testing.T
	e     *Engine
	clock fixtureClock
}

func setup(t *testing.T) *harness {
	t.Helper()
	e, err := New(DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, e: e}
	if _, err := e.SetAvailability(0, accessible); err != nil {
		t.Fatal(err)
	}
	h.frame(0)
	return h
}
func three(x float64) []Contact { return []Contact{{ID: 1, XMM: x}, {ID: 2, XMM: x}, {ID: 3, XMM: x}} }
func (h *harness) frame(at time.Duration, contacts ...Contact) []Effect {
	h.t.Helper()
	h.clock.at = at
	fx, err := h.e.Frame(h.clock.frame(contacts...))
	if err != nil {
		h.t.Fatal(err)
	}
	return fx
}
func kinds(fx []Effect) []EffectKind {
	out := make([]EffectKind, 0, len(fx))
	for _, f := range fx {
		out = append(out, f.Kind)
	}
	return out
}
func wantKinds(t *testing.T, fx []Effect, want ...EffectKind) {
	t.Helper()
	if !slices.Equal(kinds(fx), want) {
		t.Fatalf("effects %v, want %v", kinds(fx), want)
	}
}
func (h *harness) start() {
	h.t.Helper()
	h.frame(time.Millisecond, three(0)...)
	wantKinds(h.t, h.frame(2*time.Millisecond, three(1)...), ResetOrdinaryContacts, HoldPrimary)
}

func TestDefaultPolicy(t *testing.T) {
	if DefaultPolicy().GracePeriod != 800*time.Millisecond {
		t.Fatal(DefaultPolicy())
	}
}

func TestStartAndMotion(t *testing.T) {
	h := setup(t)
	wantKinds(t, h.frame(time.Millisecond, three(0)...), SuppressOrdinary)
	wantKinds(t, h.frame(2*time.Millisecond, three(.99)...), SuppressOrdinary)
	wantKinds(t, h.frame(3*time.Millisecond, three(1)...), ResetOrdinaryContacts, HoldPrimary)
	fx := h.frame(4*time.Millisecond, three(1.125)...)
	wantKinds(t, fx, MovePointer)
	if fx[0].Delta.XMM != .125 {
		t.Fatal(fx)
	}
	wantKinds(t, h.frame(time.Hour, three(1.125)...))
	if h.e.State().Phase != Dragging {
		t.Fatal(h.e.State())
	}
}

func TestBaselineReplacement(t *testing.T) {
	h := setup(t)
	h.frame(time.Millisecond, three(0)...)
	c := three(100)
	c[0].ID = 4
	wantKinds(t, h.frame(2*time.Millisecond, c...), SuppressOrdinary)
	c = three(101)
	c[0].ID = 4
	wantKinds(t, h.frame(3*time.Millisecond, c...), ResetOrdinaryContacts, HoldPrimary)
	c[0].ID = 5
	c[0].XMM = 1000
	wantKinds(t, h.frame(4*time.Millisecond, c...))
}

func TestGraceContinuationAndExpiry(t *testing.T) {
	for _, tt := range []struct {
		name   string
		at     time.Duration
		resume bool
	}{
		{"before", 809 * time.Millisecond, true}, {"exact", 810 * time.Millisecond, false}, {"after", 811 * time.Millisecond, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t)
			h.start()
			fx := h.frame(10 * time.Millisecond)
			wantKinds(t, fx, SuppressOrdinary, ScheduleGrace)
			if fx[1].Deadline != 810*time.Millisecond {
				t.Fatal(fx)
			}
			generation := fx[1].Generation
			fx = h.frame(tt.at, three(100)...)
			if tt.resume {
				wantKinds(t, fx, CancelGrace)
				fx, err := h.e.Tick(time.Second, generation)
				if err != nil {
					t.Fatal(err)
				}
				wantKinds(t, fx)
				wantKinds(t, h.frame(2*time.Second, three(100)...))
				wantKinds(t, h.frame(3*time.Second, three(100.125)...), MovePointer)
			} else {
				wantKinds(t, fx, CancelGrace, ReleasePrimary, SuppressOrdinary)
				if !h.e.Output().ReleasePending {
					t.Fatal("lost cleanup")
				}
			}
		})
	}
}

func TestFixedDeadlineAndNewInterruption(t *testing.T) {
	h := setup(t)
	h.start()
	h.frame(10*time.Millisecond, three(1)[:2]...)
	h.frame(100 * time.Millisecond)
	if h.e.State().GraceDeadline != 810*time.Millisecond {
		t.Fatal("deadline extended")
	}
	h.frame(200*time.Millisecond, three(100)...)
	fx := h.frame(250 * time.Millisecond)
	if fx[1].Deadline != 1050*time.Millisecond {
		t.Fatal(fx)
	}
}

func TestFreshSingleReturnOrdering(t *testing.T) {
	h := setup(t)
	h.start()
	h.frame(10 * time.Millisecond)
	fx := h.frame(20*time.Millisecond, Contact{ID: 9, XMM: 20})
	wantKinds(t, fx, CancelGrace, ReleasePrimary, SuppressOrdinary)
	wantKinds(t, h.frame(21*time.Millisecond, Contact{ID: 9, XMM: 21}), SuppressOrdinary)
	fx, err := h.e.ConfirmCleanup(22*time.Millisecond, h.e.Output().CleanupGeneration)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, RestoreOrdinaryContacts)
	if len(fx[0].Contacts) != 1 {
		t.Fatal(fx)
	}
	fx[0].Contacts[0].XMM = 999
	if h.e.last.Contacts[0].XMM != 21 {
		t.Fatal("restoration aliases input")
	}
	wantKinds(t, h.frame(23*time.Millisecond, three(100)...), ForwardOrdinary)
	if h.e.State().Armed {
		t.Fatal("rearmed before all lifted")
	}
	h.frame(24 * time.Millisecond)
	if !h.e.State().Armed {
		t.Fatal("not rearmed")
	}
}

func TestRemainingFingerAndPalm(t *testing.T) {
	h := setup(t)
	h.start()
	h.frame(10*time.Millisecond, Contact{ID: 1})
	wantKinds(t, h.frame(20*time.Millisecond, Contact{ID: 1}), SuppressOrdinary)
	h.frame(30*time.Millisecond, Contact{ID: 8, Palm: true})
	wantKinds(t, h.frame(40*time.Millisecond, Contact{ID: 9}, Contact{ID: 8, Palm: true}), CancelGrace, ReleasePrimary, SuppressOrdinary)
}

func TestFailureAndRepeatedCommands(t *testing.T) {
	h := setup(t)
	h.start()
	fx, err := h.e.SetAvailability(3*time.Millisecond, accessible)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx)
	fx, err = h.e.Frame(Frame{At: 4 * time.Millisecond, Reliable: false})
	if !errors.Is(err, ErrUnreliableFrame) {
		t.Fatal(err)
	}
	wantKinds(t, fx, ReleasePrimary, SuppressOrdinary)
	fx, err = h.e.Cancel(5 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx)
	if !h.e.Output().ReleasePending {
		t.Fatal("cancel cleared cleanup")
	}
	h.e.ConfirmCleanup(6*time.Millisecond, h.e.Output().CleanupGeneration)
	h.e.SetAvailability(7*time.Millisecond, accessible)
	h.frame(8*time.Millisecond, three(10)...)
	if h.e.State().Armed {
		t.Fatal("resumed failed gesture")
	}
	h.frame(9 * time.Millisecond)
	if !h.e.State().Armed {
		t.Fatal("not recovered after neutral")
	}
}

func TestTimerReleaseRearmsWithoutNewEmptyFrame(t *testing.T) {
	h := setup(t)
	h.start()
	h.frame(10 * time.Millisecond)
	fx, err := h.e.Tick(810*time.Millisecond, h.e.State().TimerGeneration)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, CancelGrace, ReleasePrimary, SuppressOrdinary)
	h.e.ConfirmCleanup(811*time.Millisecond, h.e.Output().CleanupGeneration)
	if !h.e.State().Armed {
		t.Fatal("requires redundant neutral frame")
	}
}

func TestPhysicalPressPreventsMovement(t *testing.T) {
	h := setup(t)
	h.start()
	fx, err := h.e.Frame(Frame{At: 3 * time.Millisecond, Reliable: true, ButtonsDown: true, Contacts: three(100)})
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, ReleasePrimary, SuppressOrdinary, ConsumePhysicalButtons)
	h.e.ConfirmCleanup(4*time.Millisecond, h.e.Output().CleanupGeneration)
	wantKinds(t, h.frame(5*time.Millisecond), RestoreOrdinaryContacts, ConsumePhysicalButtons)
	if !h.e.State().Armed {
		t.Fatal("no rearm after button release")
	}
}

func TestInvalidTimeAndDeadlineOverflow(t *testing.T) {
	h := setup(t)
	h.start()
	fx, err := h.e.Frame(Frame{At: time.Millisecond, Reliable: true})
	if err == nil {
		t.Fatal("accepted stale frame")
	}
	wantKinds(t, fx, ReleasePrimary, SuppressOrdinary)
	h = setup(t)
	h.start()
	fx, err = h.e.Frame(Frame{At: time.Duration(math.MaxInt64), Reliable: true})
	if err == nil {
		t.Fatal("accepted overflowing deadline")
	}
	wantKinds(t, fx, ReleasePrimary, SuppressOrdinary)
}

func TestSyntheticSequenceButtonInvariant(t *testing.T) {
	for seed := uint64(1); seed <= 20; seed++ {
		h := setup(t)
		random := seed
		held := false
		for i := 1; i <= 500; i++ {
			random = random*6364136223846793005 + 1
			count := int(random % 5)
			var contacts []Contact
			for id := 0; id < count; id++ {
				contacts = append(contacts, Contact{ID: int64(id), XMM: float64((random >> 16) % 100)})
			}
			at := time.Duration(i) * 20 * time.Millisecond
			fx := h.frame(at, contacts...)
			for _, effect := range fx {
				switch effect.Kind {
				case HoldPrimary:
					if held {
						t.Fatal("duplicate hold")
					}
					held = true
				case ReleasePrimary:
					if !held {
						t.Fatal("release without hold")
					}
					held = false
				case MovePointer:
					if !held {
						t.Fatal("drag movement without hold")
					}
				}
			}
			if h.e.Output().ReleasePending {
				if _, err := h.e.ConfirmCleanup(at, h.e.Output().CleanupGeneration); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
