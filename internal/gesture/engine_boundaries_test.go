package gesture

import (
	"math"
	"testing"
	"time"
)

func TestMotionContract(t *testing.T) {
	for _, tt := range []struct {
		name   string
		dx, dy float64
	}{
		{"fractional right", .125, 0}, {"left", -.25, 0}, {"up", 0, -.125}, {"down", 0, .25}, {"diagonal", .25, -.125},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := setup(t)
			h.start()
			c := three(1 + tt.dx)
			for i := range c {
				c[i].YMM = tt.dy
			}
			fx := h.frame(4*time.Millisecond, c...)
			wantKinds(t, fx, MovePointer)
			if math.Abs(fx[0].Delta.XMM-tt.dx) > 1e-12 || math.Abs(fx[0].Delta.YMM-tt.dy) > 1e-12 || fx[0].Elapsed != 2*time.Millisecond || fx[0].At != 4*time.Millisecond {
				t.Fatal(fx)
			}
			wantKinds(t, h.frame(5*time.Millisecond, c...))
		})
	}
}

func TestStartPolicyAndGroupMovement(t *testing.T) {
	e, err := New(Policy{StartThresholdMM: 2, GracePeriod: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, e: e}
	e.SetAvailability(0, accessible)
	h.frame(0)
	h.frame(time.Millisecond, Contact{ID: 1, XMM: 100})
	h.frame(time.Second, Contact{ID: 1}, Contact{ID: 2})
	h.frame(2*time.Second, three(0)...)
	wantKinds(t, h.frame(3*time.Second, three(1)...), SuppressOrdinary)
	wantKinds(t, h.frame(4*time.Second, three(0)...), SuppressOrdinary)
	c := three(0)
	c[0].XMM = 6
	wantKinds(t, h.frame(5*time.Second, c...), ResetOrdinaryContacts, HoldPrimary)
}

func TestEveryContactDeparturePauses(t *testing.T) {
	for _, n := range []int{0, 1, 2, 4, 5} {
		t.Run(string(rune('0'+n)), func(t *testing.T) {
			h := setup(t)
			h.start()
			var c []Contact
			for i := 0; i < n; i++ {
				c = append(c, Contact{ID: int64(i + 1), XMM: 100})
			}
			fx := h.frame(10*time.Millisecond, c...)
			wantKinds(t, fx, SuppressOrdinary, ScheduleGrace)
			if fx[1].Deadline != 810*time.Millisecond {
				t.Fatal(fx)
			}
		})
	}
}

func TestPendingCleanupCannotBeBypassed(t *testing.T) {
	h := setup(t)
	h.start()
	h.e.Cancel(3 * time.Millisecond)
	first := h.e.Output().CleanupGeneration
	wantKinds(t, h.frame(4*time.Millisecond, three(10)...), SuppressOrdinary)
	fx, err := h.e.ConfirmCleanup(5*time.Millisecond, first+1)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx)
	if !h.e.Output().ReleasePending {
		t.Fatal("incorrect cleanup token accepted")
	}
	h.frame(6 * time.Millisecond)
	h.e.ConfirmCleanup(7*time.Millisecond, first)
	h.frame(8*time.Millisecond, three(0)...)
	h.frame(9*time.Millisecond, three(1)...)
	h.e.Cancel(10 * time.Millisecond)
	fx, err = h.e.ConfirmCleanup(11*time.Millisecond, first)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx)
	if !h.e.Output().ReleasePending {
		t.Fatal("stale acknowledgment cleared new obligation")
	}
}

func TestSuppressedCandidateRestoration(t *testing.T) {
	h := setup(t)
	h.frame(time.Millisecond, three(0)...)
	fx := h.frame(2*time.Millisecond, three(0)[:1]...)
	wantKinds(t, fx, RestoreOrdinaryContacts)
	if len(fx[0].Contacts) != 1 {
		t.Fatal(fx)
	}
	wantKinds(t, h.frame(3*time.Millisecond, three(0)[:1]...), ForwardOrdinary)
}

func TestBatchTerminationPrecedence(t *testing.T) {
	for _, name := range []string{"cancel", "failure", "disable", "lock", "physical"} {
		t.Run(name, func(t *testing.T) {
			h := setup(t)
			h.frame(time.Millisecond, three(0)...)
			frame := Frame{At: 2 * time.Millisecond, Reliable: true, Contacts: three(100)}
			input := Input{At: frame.At, Frame: &frame}
			a := accessible
			switch name {
			case "cancel":
				input.Cancel = true
			case "failure":
				input.Failure = true
			case "disable":
				a.Enabled = false
				input.Availability = &a
			case "lock":
				a.LifecycleAvailable = false
				input.Availability = &a
			case "physical":
				frame.ButtonsDown = true
			}
			fx, err := h.e.Process(input)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range fx {
				if f.Kind == HoldPrimary || f.Kind == MovePointer {
					t.Fatal("termination lost", fx)
				}
			}
		})
	}
}

func TestBatchExpiryPrecedence(t *testing.T) {
	h := setup(t)
	h.start()
	h.frame(10 * time.Millisecond)
	generation := h.e.State().TimerGeneration
	frame := Frame{At: 810 * time.Millisecond, Reliable: true, Contacts: three(100)}
	fx, err := h.e.Process(Input{At: frame.At, Frame: &frame, TimerGeneration: &generation})
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, CancelGrace, ReleasePrimary, SuppressOrdinary)
}

func TestBatchCancelPreservesPhysicalConsumption(t *testing.T) {
	h := setup(t)
	h.start()
	frame := Frame{At: 3 * time.Millisecond, Reliable: true, Contacts: three(1), ButtonsDown: true}
	fx, err := h.e.Process(Input{At: frame.At, Cancel: true, Frame: &frame})
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, ReleasePrimary, SuppressOrdinary, ConsumePhysicalButtons)
	h.e.ConfirmCleanup(4*time.Millisecond, h.e.Output().CleanupGeneration)
	wantKinds(t, h.frame(5*time.Millisecond), RestoreOrdinaryContacts, ConsumePhysicalButtons)
}

func TestAvailabilityRecoveryAndPhysicalRelease(t *testing.T) {
	h := setup(t)
	h.start()
	a := accessible
	a.Enabled = false
	fx, err := h.e.SetAvailability(3*time.Millisecond, a)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, ReleasePrimary, SuppressOrdinary)
	fx, err = h.e.SetAvailability(4*time.Millisecond, a)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx)
	fx, err = h.e.ConfirmCleanup(5*time.Millisecond, h.e.Output().CleanupGeneration)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, RestoreOrdinaryContacts)
	wantKinds(t, h.frame(6*time.Millisecond, three(10)...), ForwardOrdinary)
	h.e.SetAvailability(7*time.Millisecond, accessible)
	h.frame(8*time.Millisecond, three(0)...)
	if h.e.State().Armed {
		t.Fatal("armed existing contacts")
	}
	h.frame(9 * time.Millisecond)
	if !h.e.State().Armed {
		t.Fatal("not armed")
	}
}

func TestFailureCancelsGraceAndRejectsStaleTimer(t *testing.T) {
	h := setup(t)
	h.start()
	h.frame(10 * time.Millisecond)
	generation := h.e.State().TimerGeneration
	fx, err := h.e.Fail(20 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx, CancelGrace, ReleasePrimary, SuppressOrdinary)
	fx, err = h.e.Tick(time.Second, generation)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds(t, fx)
	if !h.e.Output().ReleasePending {
		t.Fatal("failure lost release")
	}
}

func TestConfiguredGraceAndFreshSingleBoundary(t *testing.T) {
	for _, tt := range []struct {
		name   string
		at     time.Duration
		single bool
	}{
		{"before", 209 * time.Millisecond, true}, {"at", 210 * time.Millisecond, false}, {"after", 211 * time.Millisecond, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, err := New(Policy{StartThresholdMM: 1, GracePeriod: 200 * time.Millisecond})
			if err != nil {
				t.Fatal(err)
			}
			h := &harness{t: t, e: e}
			e.SetAvailability(0, accessible)
			h.frame(0)
			h.start()
			h.frame(10 * time.Millisecond)
			generation := e.State().TimerGeneration
			fx, err := e.Tick(100*time.Millisecond, generation)
			if err != nil {
				t.Fatal(err)
			}
			wantKinds(t, fx)
			fx = h.frame(tt.at, Contact{ID: 8})
			wantKinds(t, fx, CancelGrace, ReleasePrimary, SuppressOrdinary)
			if (e.State().Phase == PointingUntilNeutral) != tt.single {
				t.Fatal(e.State())
			}
		})
	}
}

func FuzzEngineEffects(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f.Add([]byte{3, 3, 0, 3, 0, 1, 2, 7, 4, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2048 {
			return
		}
		h := setup(t)
		held := false
		for i, b := range data {
			at := time.Duration(i+1) * 100 * time.Millisecond
			var fx []Effect
			var err error
			switch b % 8 {
			case 5:
				fx, err = h.e.Cancel(at)
			case 6:
				fx, err = h.e.Tick(at, h.e.State().TimerGeneration)
			case 7:
				fx, err = h.e.Fail(at)
			default:
				var c []Contact
				for n := 0; n < int(b%5); n++ {
					c = append(c, Contact{ID: int64(n), XMM: float64(b), Palm: b == 255})
				}
				fx, err = h.e.Frame(Frame{At: at, Reliable: true, Contacts: c})
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, effect := range fx {
				switch effect.Kind {
				case HoldPrimary:
					if held {
						t.Fatal("duplicate hold")
					}
					held = true
				case ReleasePrimary:
					if !held {
						t.Fatal("unowned release")
					}
					held = false
				case MovePointer:
					if !held || !finite(effect.Delta.XMM) || !finite(effect.Delta.YMM) {
						t.Fatal("invalid movement")
					}
				}
			}
			if h.e.Output().ReleasePending && b%2 == 0 {
				h.e.ConfirmCleanup(at, h.e.Output().CleanupGeneration)
			}
			if !h.e.availability.InputAvailable && b%3 == 0 {
				h.e.SetAvailability(at, accessible)
			}
		}
		fx, err := h.e.Cancel(time.Duration(len(data)+1) * 100 * time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
		for _, effect := range fx {
			if effect.Kind == ReleasePrimary {
				held = false
			}
		}
		if held {
			t.Fatal("hold survives final cancellation")
		}
	})
}
