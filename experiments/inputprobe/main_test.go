//go:build linux && cgo

package main

import (
	"io"
	"testing"
)

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		name  string
		args  []string
		valid bool
	}{
		{"physical requires selection", nil, false},
		{"synthetic lifecycle", []string{"-mode=create"}, true},
		{"reject accidental physical create", []string{"-mode=create", "-device=/dev/input/event10"}, false},
		{"bounded lifetime", []string{"-mode=create", "-duration=61s"}, false},
		{"negative lifetime", []string{"-mode=create", "-duration=-1s"}, false},
		{"settle shorter than lifetime", []string{"-mode=create", "-duration=2s"}, false},
		{"inspect explicit device", []string{"-device=/dev/input/event10"}, true},
		{"no pointer hold on touchpad", []string{"-mode=create", "-hold"}, false},
		{"explicit pointer hold", []string{"-mode=create", "-kind=pointer", "-hold"}, true},
		{"invalid mode", []string{"-mode=drag"}, false},
		{"steady explicit device", []string{"-mode=steady", "-device=/dev/input/event9", "-duration=15s", "-settle=5s"}, true},
		{"steady without device", []string{"-mode=steady"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse(tt.args, io.Discard)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v, error=%v", tt.valid, err)
			}
		})
	}
}

func TestFrameBuffer(t *testing.T) {
	t.Run("never expose partial frames", func(t *testing.T) {
		var b frameBuffer
		movement := event{3, 53, 100}
		if frame, err := b.add(movement); frame != nil || err != nil {
			t.Fatalf("partial frame=%v error=%v", frame, err)
		}
		frame, err := b.add(event{0, 0, 0})
		if err != nil || len(frame) != 2 || frame[0] != movement {
			t.Fatalf("complete frame=%v error=%v", frame, err)
		}
	})
	t.Run("reject synchronization loss", func(t *testing.T) {
		var b frameBuffer
		b.add(event{3, 53, 100})
		if frame, err := b.add(event{0, 3, 0}); frame != nil || err == nil {
			t.Fatal("lost synchronization was accepted")
		}
		frame, err := b.add(event{0, 0, 0})
		if err != nil || len(frame) != 1 {
			t.Fatal("stale events survived synchronization failure")
		}
	})
	t.Run("bound frame memory", func(t *testing.T) {
		var b frameBuffer
		for i := 0; i < 512; i++ {
			if _, err := b.add(event{3, 53, 100}); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := b.add(event{3, 53, 100}); err == nil {
			t.Fatal("unbounded frame accepted")
		}
	})
}

func TestAggregate(t *testing.T) {
	var a aggregate
	a.observe(3, 0, 0)
	a.observe(4, 1, 1)
	a.observe(4, 0, 0)
	a.observe(0, 0, 0)
	if a.Frames != 4 || a.MaxContacts != 4 || a.FramesWithPalm != 1 || a.ContactCountChanges != 3 || a.FramesWithButton != 1 {
		t.Fatalf("unexpected aggregate: %+v", a)
	}
}

func TestAggregateMixedContacts(t *testing.T) {
	for _, tt := range []struct {
		name            string
		contacts, palms int
		want            uint64
	}{
		{"three fingers and palm", 4, 1, 1},
		{"three fingers and two palms", 5, 2, 1},
		{"three fingers only", 3, 0, 0},
		{"two fingers and palm", 3, 1, 0},
		{"palm only", 1, 1, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var a aggregate
			a.observe(tt.contacts, tt.palms, 0)
			if a.FramesWithThreeNonPalmAndPalm != tt.want {
				t.Fatalf("mixed-contact frames = %d, want %d", a.FramesWithThreeNonPalmAndPalm, tt.want)
			}
		})
	}
}
