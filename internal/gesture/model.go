// Package gesture defines device-independent gesture state and effects.
// Callers supply complete contact frames and monotonic time in one clock domain.
package gesture

import "time"

// Contact is an active touch expressed in physical millimeters.
// Thumbs are valid contacts. Only contacts classified as palms are excluded.
type Contact struct {
	ID   int64
	XMM  float64
	YMM  float64
	Palm bool
}

// Frame contains only the selected touchpad's contacts and button state.
// At is elapsed monotonic time from the runtime's chosen origin.
type Frame struct {
	At          time.Duration
	Contacts    []Contact
	ButtonsDown bool
	Reliable    bool
}

// Point is a physical position or relative displacement in millimeters.
type Point struct{ XMM, YMM float64 }

// Policy contains recognition settings, independent of file configuration.
// Pointer acceleration belongs to the output mapping boundary.
type Policy struct {
	StartThresholdMM float64
	GracePeriod      time.Duration
}

// DefaultPolicy returns the approved start and repositioning defaults.
func DefaultPolicy() Policy {
	return Policy{StartThresholdMM: 1, GracePeriod: 800 * time.Millisecond}
}

// Phase describes gesture recognition, not confirmed desktop button state.
type Phase uint8

const (
	Unavailable Phase = iota
	AwaitingNeutral
	Ready
	Dragging
	Grace
	PointingUntilNeutral
)

// Availability separates requested enablement from readiness prerequisites.
type Availability struct {
	Enabled            bool
	InputAvailable     bool
	OutputAvailable    bool
	LifecycleAvailable bool
}

// Baseline records a fixed group position and its contact identities.
type Baseline struct {
	IDs      [3]int64
	Position Point
	Valid    bool
}

// State stores transient recognition state. It must never be persisted.
type State struct {
	Phase              Phase
	Armed              bool
	StartBaseline      Baseline
	MotionBaseline     Baseline
	GraceDeadline      time.Duration
	TimerGeneration    uint64
	SawZeroDuringGrace bool
	// MotionAt is the timestamp of the current motion baseline.
	MotionAt time.Duration
}

// OutputOwner binds a cleanup obligation to one device in one desktop session.
type OutputOwner struct{ Device, Session string }

// ButtonState distinguishes known release from an uncertain output write.
type ButtonState uint8

const (
	ButtonReleased ButtonState = iota
	ButtonHeld
	ButtonUnknown
)

// OutputState is separate from State: a terminated gesture can need cleanup.
// A successful submission alone is not confirmation of desktop receipt.
type OutputState struct {
	Owner          OutputOwner
	Desired        ButtonState
	Submitted      ButtonState
	ReleasePending bool
	// CleanupGeneration prevents a delayed acknowledgment from clearing a later hold.
	CleanupGeneration uint64
}

// EffectKind identifies a runtime action without native device handles.
type EffectKind uint8

const (
	HoldPrimary EffectKind = iota
	MovePointer
	ReleasePrimary
	ScheduleGrace
	CancelGrace
	ForwardOrdinary
	SuppressOrdinary
	ResetOrdinaryContacts
	RestoreOrdinaryContacts
	ConsumePhysicalButtons
)

// Effect carries relative motion or a generation-tagged timer action.
// Effects must be applied in order by the runtime.
type Effect struct {
	Kind       EffectKind
	Delta      Point
	Deadline   time.Duration
	Generation uint64
	// At and Elapsed let the output mapper preserve physical speed and fractions.
	At      time.Duration
	Elapsed time.Duration
	// Contacts is an owned snapshot for ordinary-input restoration after cleanup.
	Contacts []Contact
}

// Input groups events at one monotonic timestamp. Process applies termination
// before movement and expiry before continuation, regardless of callback order.
// A runtime must collect simultaneous events before calling Process.
type Input struct {
	At              time.Duration
	Availability    *Availability
	Failure         bool
	Cancel          bool
	Frame           *Frame
	TimerGeneration *uint64
}
