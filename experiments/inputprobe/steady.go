//go:build linux && cgo

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"time"
)

// contact is temporary input state. It is never serialized or logged.
type contact struct {
	id   int
	x, y float64
	palm bool
}

type steadyResult struct {
	Complete          bool    `json:"measurement_complete"`
	WindowMS          float64 `json:"measurement_window_ms"`
	Frames            uint64  `json:"measurement_frames"`
	MinNonPalm        int     `json:"min_non_palm_contacts"`
	MaxNonPalm        int     `json:"max_non_palm_contacts"`
	CountChanges      uint64  `json:"non_palm_count_changes"`
	IdentityChanges   uint64  `json:"identity_changes_at_unchanged_nonzero_count"`
	PalmFrames        uint64  `json:"frames_with_palm_tool"`
	MixedFrames       uint64  `json:"frames_with_three_non_palm_and_palm"`
	ThreeFrames       uint64  `json:"frames_with_three_non_palm_contacts"`
	Segments          uint64  `json:"three_contact_segments"`
	MaxDisplacementMM float64 `json:"max_group_displacement_from_segment_start_mm"`
	FirstFrameDelayMS float64 `json:"first_frame_delay_ms"`
	LastFrameOffsetMS float64 `json:"last_frame_offset_ms"`
	MaxFrameGapMS     float64 `json:"max_observed_frame_gap_ms"`
}

type steadyWindow struct {
	start, end           time.Duration
	result               steadyResult
	previousIDs          []int
	baselineX, baselineY float64
	last                 time.Duration
}

func newSteadyWindow(start, end time.Duration) *steadyWindow {
	return &steadyWindow{start: start, end: end, result: steadyResult{WindowMS: float64(end-start) / float64(time.Millisecond)}}
}

func (w *steadyWindow) observe(at time.Duration, contacts []contact) {
	if at < w.start || at >= w.end {
		return
	}
	ids := make([]int, 0, len(contacts))
	var x, y float64
	palms := 0
	for _, c := range contacts {
		if c.palm {
			palms++
			continue
		}
		ids = append(ids, c.id)
		x += c.x
		y += c.y
	}
	slices.Sort(ids)
	n := len(ids)
	r := &w.result
	first := r.Frames == 0
	same := !first && slices.Equal(ids, w.previousIDs)
	if first {
		r.MinNonPalm = n
		r.FirstFrameDelayMS = float64(at-w.start) / float64(time.Millisecond)
	} else {
		if n != len(w.previousIDs) {
			r.CountChanges++
		}
		if n > 0 && n == len(w.previousIDs) && !same {
			r.IdentityChanges++
		}
		r.MaxFrameGapMS = max(r.MaxFrameGapMS, float64(at-w.last)/float64(time.Millisecond))
	}
	r.MinNonPalm = min(r.MinNonPalm, n)
	r.MaxNonPalm = max(r.MaxNonPalm, n)
	if palms > 0 {
		r.PalmFrames++
	}
	if n == 3 {
		r.ThreeFrames++
		if palms > 0 {
			r.MixedFrames++
		}
		x /= 3
		y /= 3
		if !same {
			w.baselineX, w.baselineY = x, y
			r.Segments++
		} else {
			r.MaxDisplacementMM = max(r.MaxDisplacementMM, math.Hypot(x-w.baselineX, y-w.baselineY))
		}
	}
	r.Frames++
	r.LastFrameOffsetMS = float64(at-w.start) / float64(time.Millisecond)
	w.previousIDs = ids
	w.last = at
}

func runSteady(ctx context.Context, p *physical, o options, stdout, stderr io.Writer) (result error) {
	// Validate resolution before asking the user to hold contacts.
	if _, err := p.contacts(); err != nil {
		return err
	}
	now, err := monotonicNow()
	if err != nil {
		return err
	}
	// The outer deadline also bounds device initialization. Never extend it here.
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("steady measurement: bounded context required")
	}
	remaining := time.Until(deadline)
	if remaining <= o.settle {
		return errors.New("steady measurement: insufficient time after device initialization")
	}
	w := newSteadyWindow(now+o.settle, now+remaining)
	if _, err := fmt.Fprintf(stderr, "PLACE CONTACTS: %.1f seconds, then hold until DONE.\n", o.settle.Seconds()); err != nil {
		return err
	}
	defer func() { result = errors.Join(result, json.NewEncoder(stdout).Encode(w.result)) }()
	var frames frameBuffer
	announced := false
	for {
		now, err = monotonicNow()
		if err != nil {
			return err
		}
		if now >= w.end {
			w.result.Complete = true
			break
		}
		if ctx.Err() != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				w.result.Complete = true
				break
			}
			return fmt.Errorf("steady measurement interrupted: %w", ctx.Err())
		}
		if !announced && now >= w.start {
			if _, err := fmt.Fprintln(stderr, "MEASURING: keep contacts still until DONE."); err != nil {
				return err
			}
			announced = true
		}
		e, present, err := p.next()
		if err != nil {
			return err
		}
		if !present {
			continue
		}
		frame, err := frames.add(e)
		if err != nil {
			return err
		}
		if frame == nil {
			continue
		}
		contacts, err := p.contacts()
		if err != nil {
			return err
		}
		w.observe(p.eventTime, contacts)
	}
	if _, err := fmt.Fprintln(stderr, "DONE: lift your contacts."); err != nil {
		return err
	}
	if w.result.Frames == 0 {
		return errors.New("steady measurement: no complete frames in measurement window; result inconclusive")
	}
	return nil
}
