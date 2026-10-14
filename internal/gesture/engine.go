package gesture

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// Engine owns recognition state. One runtime owner must serialize its calls.
// It never reads a clock, file, device, or desktop connection.
type Engine struct {
	policy             Policy
	state              State
	availability       Availability
	output             OutputState
	lastAt             time.Duration
	last               Frame
	valid              []Contact
	haveFrame          bool
	consumeButtons     bool
	ordinarySuppressed bool
}

// New creates an initially unavailable engine. Policy is copied and immutable.
func New(policy Policy) (*Engine, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &Engine{policy: policy}, nil
}

// State returns a value snapshot without native or mutable contact references.
func (e *Engine) State() State { return e.state }

// Output returns desired output and cleanup state, not desktop confirmation.
func (e *Engine) Output() OutputState { return e.output }

func (e *Engine) available() bool {
	a := e.availability
	return a.Enabled && a.InputAvailable && a.OutputAvailable && a.LifecycleAvailable
}

func (e *Engine) time(at time.Duration) error {
	if at < 0 || at < e.lastAt {
		return errors.New("gesture time: negative or out-of-order monotonic timestamp")
	}
	e.lastAt = at
	return nil
}

// Process is the ordered entry point for simultaneous events. Cleanup evidence
// is handled separately by ConfirmCleanup after the runtime verifies its owner.
func (e *Engine) Process(input Input) ([]Effect, error) {
	if input.Frame != nil && input.Frame.At != input.At {
		return e.fail(), errors.New("process frame: timestamp differs from event batch")
	}
	if input.Failure {
		physical := e.active() && input.Frame != nil && input.Frame.ButtonsDown
		fx, err := e.Fail(input.At)
		if physical {
			e.consumeButtons = true
			fx = append(fx, Effect{Kind: ConsumePhysicalButtons})
		}
		return fx, err
	}
	var effects []Effect
	physical := e.active() && input.Frame != nil && input.Frame.ButtonsDown
	if input.Availability != nil {
		fx, err := e.SetAvailability(input.At, *input.Availability)
		effects = append(effects, fx...)
		if err != nil {
			return effects, err
		}
	}
	if input.Cancel {
		fx, err := e.Cancel(input.At)
		effects = append(effects, fx...)
		if err != nil {
			return effects, err
		}
		if physical {
			e.consumeButtons = true
			effects = append(effects, Effect{Kind: ConsumePhysicalButtons})
		}
		// Cancellation also prevents a not-yet-started gesture at this timestamp.
		return effects, nil
	}
	if physical && !e.active() {
		e.consumeButtons = true
	}
	if input.Frame != nil {
		fx, err := e.Frame(*input.Frame)
		effects = append(effects, fx...)
		if err != nil {
			return effects, err
		}
	}
	if input.TimerGeneration != nil {
		fx, err := e.Tick(input.At, *input.TimerGeneration)
		effects = append(effects, fx...)
		return effects, err
	}
	if input.Frame == nil && input.Availability == nil {
		if err := e.time(input.At); err != nil {
			return e.fail(), err
		}
	}
	return effects, nil
}

// SetAvailability handles enablement and lifecycle changes. Repeated values are
// harmless. Termination calls must precede movement at a shared timestamp.
func (e *Engine) SetAvailability(at time.Duration, a Availability) ([]Effect, error) {
	if err := e.time(at); err != nil {
		return e.fail(), err
	}
	if a == e.availability {
		return nil, nil
	}
	wasAvailable := e.available()
	e.availability = a
	if !e.available() {
		fx := e.terminate(Unavailable)
		if !a.InputAvailable {
			e.haveFrame = false
		}
		if !e.output.ReleasePending && !e.consumeButtons && e.haveFrame {
			fx = append(fx, e.ordinary()...)
		}
		return fx, nil
	}
	if !wasAvailable {
		e.haveFrame = false
		e.state.Phase = AwaitingNeutral
		e.state.Armed = false
	}
	return nil, nil
}

// Frame processes one complete snapshot from the selected touchpad. Invalid
// input terminates recognition; returned failure effects must still be applied.
func (e *Engine) Frame(frame Frame) ([]Effect, error) {
	if err := e.time(frame.At); err != nil {
		return e.fail(), err
	}
	valid, err := ValidContacts(frame)
	if err != nil {
		return e.fail(), err
	}
	previousCount := len(e.valid)
	e.last = frame
	e.last.Contacts = append([]Contact(nil), frame.Contacts...)
	e.valid = valid
	e.haveFrame = true
	n := len(valid)
	neutral := n == 0 && !frame.ButtonsDown
	if e.consumeButtons {
		if !frame.ButtonsDown {
			e.consumeButtons = false
		}
		if neutral && !e.output.ReleasePending && e.available() {
			e.arm()
		}
		if e.output.ReleasePending {
			return []Effect{{Kind: SuppressOrdinary}, {Kind: ConsumePhysicalButtons}}, nil
		}
		return append(e.ordinary(), Effect{Kind: ConsumePhysicalButtons}), nil
	}
	if e.output.ReleasePending {
		return []Effect{{Kind: SuppressOrdinary}}, nil
	}
	if !e.available() {
		if !e.availability.InputAvailable {
			return []Effect{{Kind: SuppressOrdinary}}, nil
		}
		return e.ordinary(), nil
	}
	// A physical action wins over movement and starts no new synthetic hold.
	if frame.ButtonsDown {
		if e.active() {
			e.consumeButtons = true
			fx := e.terminate(AwaitingNeutral)
			return append(fx, Effect{Kind: ConsumePhysicalButtons}), nil
		}
		e.state.Armed = false
		e.state.Phase = AwaitingNeutral
		e.state.StartBaseline = Baseline{}
		return e.ordinary(), nil
	}
	if e.state.Phase == Grace && frame.At >= e.state.GraceDeadline {
		// Count this frame for later rearming, but never use its movement to resume.
		return e.terminate(AwaitingNeutral), nil
	}
	switch e.state.Phase {
	case Unavailable, AwaitingNeutral:
		if neutral {
			e.arm()
		}
		return e.ordinary(), nil
	case PointingUntilNeutral:
		if neutral {
			e.arm()
		}
		return e.ordinary(), nil
	case Ready:
		if n != 3 {
			e.state.StartBaseline = Baseline{}
			if n > 3 {
				return e.suppress(), nil
			}
			return e.ordinary(), nil
		}
		b := baseline(valid)
		if !e.state.StartBaseline.Valid || b.IDs != e.state.StartBaseline.IDs {
			e.state.StartBaseline = b
			return e.suppress(), nil
		}
		p := e.state.StartBaseline.Position
		if math.Hypot(b.Position.XMM-p.XMM, b.Position.YMM-p.YMM) < e.policy.StartThresholdMM {
			return e.suppress(), nil
		}
		e.state.Phase, e.state.Armed = Dragging, false
		e.state.MotionBaseline = b
		e.state.MotionAt = frame.At
		e.output.Desired = ButtonHeld
		e.output.Submitted = ButtonUnknown
		e.ordinarySuppressed = true
		return []Effect{{Kind: ResetOrdinaryContacts}, {Kind: HoldPrimary}}, nil
	case Dragging:
		if n != 3 {
			if frame.At > time.Duration(math.MaxInt64)-e.policy.GracePeriod {
				return e.fail(), errors.New("schedule grace: monotonic deadline overflows")
			}
			e.state.Phase = Grace
			e.state.GraceDeadline = frame.At + e.policy.GracePeriod
			e.state.TimerGeneration++
			e.state.SawZeroDuringGrace = n == 0
			return []Effect{{Kind: SuppressOrdinary}, {Kind: ScheduleGrace, Deadline: e.state.GraceDeadline, Generation: e.state.TimerGeneration}}, nil
		}
		b := baseline(valid)
		old := e.state.MotionBaseline
		elapsed := frame.At - e.state.MotionAt
		e.state.MotionBaseline = b
		e.state.MotionAt = frame.At
		if old.IDs != b.IDs {
			return nil, nil
		}
		delta := Point{b.Position.XMM - old.Position.XMM, b.Position.YMM - old.Position.YMM}
		if !finite(delta.XMM) || !finite(delta.YMM) {
			return e.fail(), ErrUnreliableFrame
		}
		if delta == (Point{}) {
			return nil, nil
		}
		return []Effect{{Kind: MovePointer, Delta: delta, At: frame.At, Elapsed: elapsed}}, nil
	case Grace:
		if n == 0 {
			e.state.SawZeroDuringGrace = true
		}
		if n == 1 && previousCount == 0 && e.state.SawZeroDuringGrace {
			// Release confirmation must precede restoring the fresh contact.
			return e.terminate(PointingUntilNeutral), nil
		}
		if n == 3 {
			fx := e.cancelTimer()
			e.state.Phase = Dragging
			e.state.MotionBaseline = baseline(valid)
			e.state.MotionAt = frame.At
			return fx, nil
		}
		return []Effect{{Kind: SuppressOrdinary}}, nil
	}
	return nil, fmt.Errorf("gesture phase: invalid value %d", e.state.Phase)
}

// Tick processes a generation-tagged deadline. Stale timers have no effects.
func (e *Engine) Tick(at time.Duration, generation uint64) ([]Effect, error) {
	if e.state.Phase != Grace || generation != e.state.TimerGeneration {
		return nil, nil
	}
	if err := e.time(at); err != nil {
		return e.fail(), err
	}
	if at < e.state.GraceDeadline {
		return nil, nil
	}
	return e.terminate(AwaitingNeutral), nil
}

// Cancel ends an active gesture without changing requested enablement.
func (e *Engine) Cancel(at time.Duration) ([]Effect, error) {
	if err := e.time(at); err != nil {
		return e.fail(), err
	}
	if !e.active() {
		return nil, nil
	}
	return e.terminate(AwaitingNeutral), nil
}

// Fail terminates input processing and invalidates cached input reliability.
// The runtime must deliver fresh availability and neutral input before recovery.
func (e *Engine) Fail(at time.Duration) ([]Effect, error) {
	err := e.time(at)
	return e.fail(), err
}

// ConfirmCleanup is called only after the runtime verifies release or owner
// disappearance. The runtime validates output-session identity before this call.
func (e *Engine) ConfirmCleanup(at time.Duration, generation uint64) ([]Effect, error) {
	if !e.output.ReleasePending || generation != e.output.CleanupGeneration {
		return nil, nil
	}
	if err := e.time(at); err != nil {
		return e.fail(), err
	}
	e.output.ReleasePending = false
	e.output.Submitted = ButtonReleased
	if !e.haveFrame || !e.last.Reliable || !e.availability.InputAvailable {
		return nil, nil
	}
	if e.available() && len(e.valid) == 0 && !e.last.ButtonsDown && !e.consumeButtons {
		e.arm()
	}
	if e.consumeButtons {
		return nil, nil
	}
	e.ordinarySuppressed = true
	return e.ordinary(), nil
}

// suppress records routing ownership even before a drag crosses its threshold.
func (e *Engine) suppress() []Effect {
	e.ordinarySuppressed = true
	return []Effect{{Kind: SuppressOrdinary}}
}

func (e *Engine) ordinary() []Effect {
	if e.ordinarySuppressed {
		e.ordinarySuppressed = false
		return []Effect{{Kind: RestoreOrdinaryContacts, Contacts: append([]Contact(nil), e.last.Contacts...)}}
	}
	return []Effect{{Kind: ForwardOrdinary}}
}

func (e *Engine) active() bool { return e.state.Phase == Dragging || e.state.Phase == Grace }

func (e *Engine) arm() {
	e.state.Phase = Ready
	e.state.Armed = true
	e.state.StartBaseline = Baseline{}
}

func (e *Engine) cancelTimer() []Effect {
	if e.state.Phase != Grace {
		return nil
	}
	fx := []Effect{{Kind: CancelGrace, Generation: e.state.TimerGeneration}}
	e.state.TimerGeneration++
	e.state.GraceDeadline = 0
	e.state.SawZeroDuringGrace = false
	return fx
}

func (e *Engine) terminate(phase Phase) []Effect {
	held := e.active()
	fx := e.cancelTimer()
	e.state.Phase = phase
	e.state.Armed = false
	e.state.StartBaseline = Baseline{}
	e.state.MotionBaseline = Baseline{}
	if held {
		e.output.Desired = ButtonReleased
		e.output.Submitted = ButtonUnknown
		e.output.ReleasePending = true
		e.output.CleanupGeneration++
		fx = append(fx, Effect{Kind: ReleasePrimary, Generation: e.output.CleanupGeneration})
	}
	if held || e.output.ReleasePending {
		fx = append(fx, Effect{Kind: SuppressOrdinary})
	}
	return fx
}

func (e *Engine) fail() []Effect {
	e.haveFrame = false
	e.availability.InputAvailable = false
	return e.terminate(Unavailable)
}

func baseline(valid []Contact) Baseline {
	b := Baseline{Valid: true}
	for i, c := range valid {
		b.IDs[i] = c.ID
		// Divide first to avoid overflowing a sum of finite coordinates.
		b.Position.XMM += c.XMM / 3
		b.Position.YMM += c.YMM / 3
	}
	return b
}
