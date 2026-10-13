//go:build linux && cgo

package main

import (
	"testing"
	"time"
)

func started(t *testing.T) *dragMachine {
	t.Helper()
	d := newDragMachine(12)
	if !d.step(0, fingers(0), 0).clearTouch {
		t.Fatal("no contact handoff")
	}
	if !d.step(time.Millisecond, fingers(1), 0).press {
		t.Fatal("no threshold hold")
	}
	return d
}

func TestDragMachine(t *testing.T) {
	t.Run("discard threshold motion and preserve fractions", func(t *testing.T) {
		d := newDragMachine(12)
		d.step(0, fingers(0), 0)
		e := d.step(time.Millisecond, fingers(1.5), 0)
		if !e.press || e.dx != 0 {
			t.Fatal(e)
		}
		e = d.step(2*time.Millisecond, fingers(2), 0)
		if e.dx != 6 || e.press {
			t.Fatal(e)
		}
	})
	t.Run("stationary hold never expires", func(t *testing.T) {
		d := started(t)
		if e := d.step(time.Hour, fingers(1), 0); e.release || e.dx != 0 || !d.held() {
			t.Fatal(e)
		}
	})
	t.Run("fixed grace and continuation", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, nil, 0)
		e := d.step(100*time.Millisecond, fingers(100), 0)
		if e.press || e.release || e.dx != 0 || d.stats.Continuations != 1 {
			t.Fatal(e, d.stats)
		}
		if d.tick(time.Second).release {
			t.Fatal("stale expiry")
		}
	})
	t.Run("expiry wins exact continuation boundary", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, nil, 0)
		if !d.step(510*time.Millisecond, fingers(101), 0).release {
			t.Fatal("expiry lost")
		}
		if d.step(520*time.Millisecond, fingers(102), 0).press {
			t.Fatal("rearmed with contacts down")
		}
	})
	t.Run("staggered return through one contact ends old drag", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, fingers(1)[:2], 0)
		d.step(30*time.Millisecond, nil, 0)
		d.step(100*time.Millisecond, fingers(100)[:1], 0)
		d.step(120*time.Millisecond, fingers(100)[:2], 0)
		d.step(140*time.Millisecond, fingers(100), 0)
		e := d.step(200*time.Millisecond, fingers(101), 0)
		if e.release || e.press || !e.forward || d.stats.Continuations != 0 || d.stats.SingleFingerEnds != 1 {
			t.Fatal(e, d.stats)
		}
	})
	t.Run("stationary return resumes and subsequent small motion moves", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, nil, 0)
		e := d.step(499*time.Millisecond, fingers(100), 0)
		if e.press || e.release || e.dx != 0 || d.stats.Continuations != 1 {
			t.Fatal(e, d.stats)
		}
		if d.tick(time.Second).release {
			t.Fatal("stationary return expired")
		}
		e = d.step(2*time.Second, fingers(100.5), 0)
		if e.dx != 6 || e.press || e.release {
			t.Fatal(e)
		}
	})
	t.Run("fresh one finger releases on contact and resumes pointing", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, nil, 0)
		e := d.step(20*time.Millisecond, []contact{{id: 5, x: 20}}, 0)
		if !e.release || !e.resumeTouch || e.press || d.stats.SingleFingerEnds != 1 {
			t.Fatal(e, d.stats)
		}
		if !d.step(30*time.Millisecond, fingers(100), 0).forward || d.held() {
			t.Fatal("old drag resumed")
		}
	})
	t.Run("remaining finger does not release without zero contacts", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, fingers(1)[:1], 0)
		if d.step(200*time.Millisecond, fingers(1)[:1], 0).release {
			t.Fatal("remaining finger released")
		}
		if !d.tick(510 * time.Millisecond).release {
			t.Fatal("no expiry")
		}
		e := d.step(520*time.Millisecond, nil, 0)
		if e.forward || e.release || e.press {
			t.Fatal(e)
		}
	})
	t.Run("two-contact return retains grace", func(t *testing.T) {
		d := started(t)
		d.step(10*time.Millisecond, nil, 0)
		if d.step(20*time.Millisecond, fingers(100)[:2], 0).release {
			t.Fatal("two contacts released")
		}
		d.step(30*time.Millisecond, fingers(100), 0)
		if d.stats.Continuations != 1 {
			t.Fatal("no continuation")
		}
	})
	t.Run("physical press releases and consumes", func(t *testing.T) {
		d := started(t)
		e := d.step(10*time.Millisecond, fingers(2), 1)
		if !e.release || e.forward || e.dx != 0 {
			t.Fatal(e)
		}
		if d.step(20*time.Millisecond, nil, 0).forward {
			t.Fatal("physical release leaked")
		}
	})
	t.Run("identity replacement and palm do not jump", func(t *testing.T) {
		d := started(t)
		c := fingers(100)
		c[0].id = 99
		c = append(c, contact{id: 50, x: 1000, palm: true})
		e := d.step(10*time.Millisecond, c, 0)
		if e.dx != 0 || e.release {
			t.Fatal(e)
		}
	})
}
