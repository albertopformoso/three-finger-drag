//go:build linux && cgo

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"slices"
	"time"
)

// cursorPoint is temporary observer state, never printed or persisted.
type cursorPoint struct{ X, Y float64 }

func cursorPosition(socket string) (cursorPoint, error) {
	c, err := net.DialTimeout("unix", socket, time.Second)
	if err != nil {
		return cursorPoint{}, fmt.Errorf("connect Hyprland observer: %w", err)
	}
	defer c.Close()
	if err := c.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return cursorPoint{}, err
	}
	if _, err := io.WriteString(c, "j/cursorpos"); err != nil {
		return cursorPoint{}, err
	}
	var p cursorPoint
	if err := json.NewDecoder(io.LimitReader(c, 4096)).Decode(&p); err != nil {
		return p, fmt.Errorf("read Hyprland cursor result: %w", err)
	}
	return p, nil
}

type travelResult struct {
	Mode            string  `json:"measurement"`
	Stroke          uint64  `json:"stroke"`
	Valid           bool    `json:"complete_stroke"`
	Reason          string  `json:"reason,omitempty"`
	Frames          uint64  `json:"contact_frames"`
	InputDistance   float64 `json:"touchpad_displacement_mm"`
	InputPath       float64 `json:"touchpad_path_length_mm"`
	InputDX         float64 `json:"touchpad_dx_mm"`
	InputDY         float64 `json:"touchpad_dy_mm"`
	MotionMS        float64 `json:"motion_span_ms"`
	AverageSpeed    float64 `json:"mean_path_speed_mm_per_s"`
	PointerDistance float64 `json:"pointer_displacement_logical_px"`
	PointerDX       float64 `json:"pointer_dx_logical_px"`
	PointerDY       float64 `json:"pointer_dy_logical_px"`
	Gain            float64 `json:"logical_px_per_mm"`
	ObserverMS      float64 `json:"max_cursor_query_ms"`
}

type travelMeter struct {
	mode                string
	readCursor          func() (cursorPoint, error)
	output              io.Writer
	ids                 []int
	x, y, lastX, lastY  float64
	firstMove, lastMove time.Duration
	lastFrame           time.Duration
	moved               bool
	startCursor         cursorPoint
	active              bool
	serial              uint64
	r                   travelResult
}

func (m *travelMeter) query() (cursorPoint, error) {
	start := time.Now()
	p, err := m.readCursor()
	m.r.ObserverMS = max(m.r.ObserverMS, float64(time.Since(start))/float64(time.Millisecond))
	return p, err
}

func (m *travelMeter) begin(at time.Duration, contacts []contact) error {
	if len(contacts) == 0 {
		return errors.New("measure stroke: no contacts")
	}
	m.serial++
	m.r = travelResult{Mode: m.mode, Stroke: m.serial, Frames: 1}
	m.ids, m.x, m.y = group(contacts)
	m.lastX, m.lastY = m.x, m.y
	m.moved = false
	m.lastFrame = at
	p, err := m.query()
	if err != nil {
		return err
	}
	m.startCursor = p
	m.active = true
	return nil
}

func (m *travelMeter) observe(at time.Duration, contacts []contact) bool {
	if !m.active {
		return false
	}
	if len(contacts) == 0 {
		return false
	}
	ids, x, y := group(contacts)
	if !slices.Equal(ids, m.ids) {
		return false
	}
	delta := math.Hypot(x-m.lastX, y-m.lastY)
	if delta > 0 {
		if !m.moved {
			m.firstMove = m.lastFrame
			m.moved = true
		}
		m.lastMove = at
		m.r.InputPath += delta
	}
	m.lastX, m.lastY = x, y
	m.lastFrame = at
	m.r.InputDX, m.r.InputDY = x-m.x, y-m.y
	m.r.InputDistance = math.Hypot(m.r.InputDX, m.r.InputDY)
	m.r.Frames++
	return true
}

func (m *travelMeter) finish(reason string) error {
	if !m.active {
		return nil
	}
	m.active = false
	p, err := m.query()
	if err != nil {
		reason = "cursor observation failed"
	}
	m.r.PointerDX, m.r.PointerDY = p.X-m.startCursor.X, p.Y-m.startCursor.Y
	m.r.PointerDistance = math.Hypot(m.r.PointerDX, m.r.PointerDY)
	m.r.MotionMS = float64(m.lastMove-m.firstMove) / float64(time.Millisecond)
	if m.moved && m.r.MotionMS > 0 {
		m.r.AverageSpeed = m.r.InputPath / (m.r.MotionMS / 1000)
	}
	if m.r.InputDistance > 0 {
		m.r.Gain = m.r.PointerDistance / m.r.InputDistance
	}
	if reason == "" && (!m.moved || m.r.InputDistance == 0 || m.r.MotionMS <= 0) {
		reason = "no measurable stroke"
	}
	m.r.Valid = reason == ""
	m.r.Reason = reason
	return errors.Join(err, json.NewEncoder(m.output).Encode(m.r))
}

func nonPalm(all []contact) []contact {
	var out []contact
	for _, c := range all {
		if !c.palm {
			out = append(out, c)
		}
	}
	return out
}

func runTravel(ctx context.Context, p *physical, o options, stdout, stderr io.Writer) (result error) {
	m := &travelMeter{mode: "ordinary_one_finger", readCursor: func() (cursorPoint, error) { return cursorPosition(o.cursorSocket) }, output: stdout}
	if _, err := p.contacts(); err != nil {
		return err
	}
	if _, err := m.query(); err != nil {
		return err
	}
	fmt.Fprintln(stderr, "REFERENCE: after READY make separate one-finger straight strokes, then lift and pause for one second. Keep away from screen edges.")
	if !wait(ctx, o.settle) {
		return nil
	}
	fmt.Fprintln(stderr, "READY: measure ordinary pointing. This mode does not grab the touchpad.")
	defer func() { result = errors.Join(result, m.finish("probe ended before finger lift")) }()
	var frames frameBuffer
	armed := false
	for ctx.Err() == nil {
		e, ok, err := p.next()
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		frame, err := frames.add(e)
		if err != nil {
			return err
		}
		if frame == nil {
			continue
		}
		all, err := p.contacts()
		if err != nil {
			return err
		}
		contacts := nonPalm(all)
		_, _, buttons := p.summary()
		if m.active {
			if len(all) == 0 && buttons == 0 {
				if !wait(ctx, 200*time.Millisecond) {
					return nil
				}
				if err := m.finish(""); err != nil {
					return err
				}
				armed = true
			} else if buttons != 0 || !m.observe(p.eventTime, contacts) {
				if err := m.finish("contact identity/count or physical button changed"); err != nil {
					return err
				}
				armed = false
			}
		} else if len(all) == 0 && buttons == 0 {
			armed = true
		} else if armed && len(contacts) == 1 && buttons == 0 {
			if err := m.begin(p.eventTime, contacts); err != nil {
				return err
			}
			armed = false
		}
	}
	return nil
}
