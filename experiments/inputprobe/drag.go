//go:build linux && cgo

package main

import (
	"math"
	"slices"
	"time"
)

type dragPhase int

const (
	dragIdle dragPhase = iota
	dragCandidate
	dragMoving
	dragGrace
	dragWaitUp
	dragPointing
)

type dragEffect struct {
	forward, clearTouch, press, release bool
	resumeTouch                         bool
	dx, dy                              int32
}

type dragStats struct {
	Starts               uint64   `json:"drag_starts"`
	Continuations        uint64   `json:"drag_continuations"`
	SingleFingerEnds     uint64   `json:"single_finger_ends"`
	Expiries             uint64   `json:"grace_expiries"`
	PhysicalEnds         uint64   `json:"physical_button_ends"`
	MotionFrames         uint64   `json:"drag_motion_frames"`
	GraceEntries         uint64   `json:"grace_entries"`
	GraceThreeReturns    uint64   `json:"grace_periods_with_three_contacts_before_deadline"`
	GraceBaselineResets  uint64   `json:"grace_baseline_resets"`
	ExpiriesWithThree    uint64   `json:"expiries_with_three_contacts"`
	MinReturnRemainingMS *float64 `json:"min_three_contact_return_time_remaining_ms"`
}

// dragMachine is a feasibility model with fixed approved recognition defaults.
// The gain is experimental output scaling, not a verified libinput equivalent.
type dragMachine struct {
	phase                      dragPhase
	ids                        []int
	x, y, fractionX, fractionY float64
	deadline                   time.Duration
	previousCount              int
	physicalAction             bool
	gain                       float64
	stats                      dragStats
	graceReturned              bool
}

func newDragMachine(gain float64) *dragMachine { return &dragMachine{gain: gain} }

func (d *dragMachine) held() bool { return d.phase == dragMoving || d.phase == dragGrace }

func (d *dragMachine) finish() dragEffect {
	e := dragEffect{release: d.held()}
	d.phase = dragWaitUp
	d.ids = nil
	d.fractionX, d.fractionY = 0, 0
	return e
}

func (d *dragMachine) tick(now time.Duration) dragEffect {
	if d.phase == dragGrace && now >= d.deadline {
		d.stats.Expiries++
		if d.previousCount == 3 {
			d.stats.ExpiriesWithThree++
		}
		return d.finish()
	}
	return dragEffect{}
}

func (d *dragMachine) step(now time.Duration, all []contact, buttons int) dragEffect {
	var valid []contact
	for _, c := range all {
		if !c.palm {
			valid = append(valid, c)
		}
	}
	n := len(valid)
	previous := d.previousCount
	d.previousCount = n
	// Expiry wins over continuation at the same timestamp.
	if d.phase == dragGrace && now >= d.deadline {
		e := d.tick(now)
		if len(all) == 0 && buttons == 0 {
			d.phase = dragIdle
		}
		return e
	}
	if d.phase == dragWaitUp {
		if len(all) == 0 && buttons == 0 {
			d.phase = dragIdle
			d.physicalAction = false
		}
		return dragEffect{}
	}
	if d.phase == dragPointing {
		if len(all) == 0 && buttons == 0 {
			d.phase = dragIdle
		}
		return dragEffect{forward: true}
	}
	if d.phase == dragIdle {
		if buttons != 0 {
			d.physicalAction = true
		}
		if d.physicalAction || n < 3 {
			if len(all) == 0 && buttons == 0 {
				d.physicalAction = false
			}
			return dragEffect{forward: true}
		}
		d.phase = dragCandidate
		d.ids = nil
		if n == 3 {
			d.baseline(valid)
		}
		return dragEffect{clearTouch: true}
	}
	if buttons != 0 {
		if d.held() {
			d.stats.PhysicalEnds++
		}
		return d.finish()
	}
	switch d.phase {
	case dragCandidate:
		if n != 3 {
			d.ids = nil
			if len(all) == 0 {
				d.phase = dragIdle
			}
			return dragEffect{}
		}
		if d.qualifies(valid) {
			d.phase = dragMoving
			d.baseline(valid)
			d.stats.Starts++
			return dragEffect{press: true}
		}
	case dragMoving:
		if n != 3 {
			d.phase = dragGrace
			d.deadline = now + 500*time.Millisecond
			d.ids = nil
			d.graceReturned = false
			d.stats.GraceEntries++
			return dragEffect{}
		}
		ids, x, y := group(valid)
		if !slices.Equal(ids, d.ids) {
			d.baseline(valid)
			return dragEffect{}
		}
		fx := (x-d.x)*d.gain + d.fractionX
		fy := (y-d.y)*d.gain + d.fractionY
		d.x, d.y = x, y
		// Real device coordinates are bounded; clamp before integer conversion.
		dx, dy := int32(math.Trunc(max(-32767, min(32767, fx)))), int32(math.Trunc(max(-32767, min(32767, fy))))
		d.fractionX, d.fractionY = fx-float64(dx), fy-float64(dy)
		if dx != 0 || dy != 0 {
			d.stats.MotionFrames++
		}
		return dragEffect{dx: dx, dy: dy}
	case dragGrace:
		if n == 3 {
			if !d.graceReturned {
				d.graceReturned = true
				d.stats.GraceThreeReturns++
				remaining := float64(d.deadline-now) / float64(time.Millisecond)
				if d.stats.MinReturnRemainingMS == nil || remaining < *d.stats.MinReturnRemainingMS {
					d.stats.MinReturnRemainingMS = &remaining
				}
			}
			d.stats.GraceBaselineResets++
			d.phase = dragMoving
			d.baseline(valid)
			d.stats.Continuations++
			return dragEffect{}
		}
		d.ids = nil
		if n == 1 && previous == 0 {
			d.stats.SingleFingerEnds++
			d.phase = dragPointing
			return dragEffect{release: true, resumeTouch: true}
		}
	}
	return dragEffect{}
}

func group(contacts []contact) ([]int, float64, float64) {
	ids := make([]int, 0, len(contacts))
	var x, y float64
	for _, c := range contacts {
		ids = append(ids, c.id)
		x += c.x
		y += c.y
	}
	slices.Sort(ids)
	return ids, x / float64(len(contacts)), y / float64(len(contacts))
}

func (d *dragMachine) baseline(contacts []contact) {
	d.ids, d.x, d.y = group(contacts)
	d.fractionX, d.fractionY = 0, 0
}

func (d *dragMachine) qualifies(contacts []contact) bool {
	ids, x, y := group(contacts)
	if !slices.Equal(ids, d.ids) {
		d.baseline(contacts)
		return false
	}
	return math.Hypot(x-d.x, y-d.y) >= 1
}
