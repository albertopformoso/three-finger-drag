//go:build linux && cgo

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// dragOutput tracks uncertain button submissions so cleanup still attempts release.
type dragOutput struct {
	touchWrite   func(event) error
	pointerWrite func(event) error
	clearFrame   []event
	button       uint16
	held         bool
}

func writeEvents(write func(event) error, events []event) error {
	for _, e := range events {
		if err := write(e); err != nil {
			return err
		}
	}
	return nil
}

func (o *dragOutput) release() error {
	if !o.held {
		return nil
	}
	// Try SYN_REPORT even if the release write fails. Preserve uncertain state.
	err := errors.Join(o.pointerWrite(event{1, o.button, 0}), o.pointerWrite(event{0, 0, 0}))
	if err == nil {
		o.held = false
	}
	return err
}

func (o *dragOutput) apply(e dragEffect, frame []event) error {
	if e.clearTouch {
		if err := writeEvents(o.touchWrite, o.clearFrame); err != nil {
			return err
		}
	}
	if e.release {
		if err := o.release(); err != nil {
			return err
		}
	}
	if e.press {
		o.held = true
		if err := writeEvents(o.pointerWrite, []event{{1, o.button, 1}, {0, 0, 0}}); err != nil {
			return err
		}
	}
	if e.dx != 0 || e.dy != 0 {
		if err := writeEvents(o.pointerWrite, []event{{2, 0, e.dx}, {2, 1, e.dy}, {0, 0, 0}}); err != nil {
			return err
		}
	}
	if e.forward || e.resumeTouch {
		return writeEvents(o.touchWrite, frame)
	}
	return nil
}

func neutralGrab(ctx context.Context, p *physical) error {
	for pass := 0; pass < 2; pass++ {
		for {
			if ctx.Err() != nil {
				return fmt.Errorf("acquire touchpad: %w", ctx.Err())
			}
			_, ok, err := p.next()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
		}
		contacts, _, buttons := p.summary()
		if contacts != 0 || buttons != 0 {
			return errors.New("acquire touchpad: lift all contacts and release buttons before READY")
		}
		if pass == 0 {
			if err := p.grab(); err != nil {
				return err
			}
		}
	}
	return nil
}

func runDrag(ctx context.Context, p *physical, options options, stdout, stderr io.Writer) (result error) {
	if _, err := p.contacts(); err != nil {
		return err
	}
	var meter *travelMeter
	if options.cursorSocket != "" {
		meter = &travelMeter{mode: "three_finger_drag", readCursor: func() (cursorPoint, error) { return cursorPosition(options.cursorSocket) }, output: stdout}
		if _, err := meter.query(); err != nil {
			return err
		}
		defer func() { result = errors.Join(result, meter.finish("probe ended before complete stroke")) }()
	}
	touch, err := createVirtual(p, true, fmt.Sprintf("tfd-probe-touchpad-%d", os.Getpid()))
	if err != nil {
		return err
	}
	defer touch.close()
	pointer, err := createVirtual(nil, false, fmt.Sprintf("tfd-probe-drag-pointer-%d", os.Getpid()))
	if err != nil {
		return err
	}
	defer pointer.close()
	button := uint16(272)
	if options.button == "right" {
		button = 273
	}
	cut := newTransferCut(p.info().Slots, p.currentSlot(), p.countKeys())
	output := &dragOutput{touchWrite: touch.write, pointerWrite: pointer.write, clearFrame: cut.contactsUp(), button: button}
	defer func() {
		result = errors.Join(result, output.release())
		// Kernel-device teardown is not compositor-confirmed production cleanup.
		if result != nil {
			fmt.Fprintln(stderr, "Probe stopped. Output teardown does not prove desktop button cleanup; verify normal pointing.")
		}
	}()
	if err := json.NewEncoder(stdout).Encode(map[string]any{
		"mode": "drag", "pid": os.Getpid(), "touchpad_node": touch.node(), "pointer_node": pointer.node(),
		"gain_units_per_mm": options.gain, "output_button": options.button, "motion_equivalence": "unverified",
	}); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stderr, "PREPARE: lift all fingers and release buttons until READY."); err != nil {
		return err
	}
	if !wait(ctx, options.settle) {
		return nil
	}
	if err := neutralGrab(ctx, p); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stderr, "READY: three-finger dragging enabled for this bounded probe. Ctrl+C stops it."); err != nil {
		return err
	}
	d := newDragMachine(options.gain)
	defer func() { result = errors.Join(result, json.NewEncoder(stdout).Encode(d.stats)) }()
	var buffer frameBuffer
	slot := p.currentSlot()
	for ctx.Err() == nil {
		e, ok, err := p.next()
		if err != nil {
			return err
		}
		if !ok {
			now, err := monotonicNow()
			if err != nil {
				return err
			}
			if err := output.apply(d.tick(now), nil); err != nil {
				return err
			}
			continue
		}
		frame, err := buffer.add(e)
		if err != nil {
			return err
		}
		if frame == nil {
			continue
		}
		initialSlot := slot
		for _, item := range frame {
			if item.kind == 3 && item.code == 47 {
				slot = int(item.value)
			}
		}
		contacts, err := p.contacts()
		if err != nil {
			return err
		}
		_, _, buttons := p.summary()
		before := d.phase
		effect := d.step(p.eventTime, contacts, buttons)
		if ctx.Err() != nil {
			break
		}
		// Re-select the physical slot after periods of suppression.
		forward := append([]event{{3, 47, int32(initialSlot)}}, frame...)
		if effect.resumeTouch {
			forward = p.snapshot()
		}
		if err := output.apply(effect, forward); err != nil {
			return err
		}
		if meter != nil {
			if effect.press || (before == dragGrace && d.phase == dragMoving) {
				if meter.active {
					if err := meter.finish("continuation interrupted stroke"); err != nil {
						return err
					}
				}
				if err := meter.begin(p.eventTime, nonPalm(contacts)); err != nil {
					return err
				}
			} else if meter.active && before == dragMoving {
				if d.phase != dragMoving {
					reason := strokeEndReason(d.phase, len(nonPalm(contacts)), buttons)
					if !wait(ctx, 200*time.Millisecond) {
						return nil
					}
					if err := meter.finish(reason); err != nil {
						return err
					}
				} else if !meter.observe(p.eventTime, nonPalm(contacts)) {
					if err := meter.finish("contact identities changed during stroke"); err != nil {
						return err
					}
				}
			}
		}
	}
	if _, err := fmt.Fprintln(stderr, "DONE: probe stopped; verify the button released and ordinary input returned."); err != nil {
		return err
	}
	return nil
}

// Motion stops at the first departure from three contacts. Remaining fingers
// need not lift in the same report to measure the completed moving segment.
func strokeEndReason(phase dragPhase, validContacts, buttons int) string {
	if phase == dragGrace && validContacts < 3 && buttons == 0 {
		return ""
	}
	return "stroke interrupted by added contacts or a physical button"
}
