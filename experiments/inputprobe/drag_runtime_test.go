//go:build linux && cgo

package main

import (
	"errors"
	"testing"
)

func TestStrokeEndReason(t *testing.T) {
	for _, tt := range []struct {
		name              string
		phase             dragPhase
		contacts, buttons int
		valid             bool
	}{
		{"all lifted", dragGrace, 0, 0, true},
		{"two remain", dragGrace, 2, 0, true},
		{"one remains", dragGrace, 1, 0, true},
		{"fourth added", dragGrace, 4, 0, false},
		{"physical press", dragWaitUp, 3, 1, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := strokeEndReason(tt.phase, tt.contacts, tt.buttons); (got == "") != tt.valid {
				t.Fatalf("reason=%q, valid=%v", got, tt.valid)
			}
		})
	}
}

func TestDragOutputFailedPressCleanup(t *testing.T) {
	var events []event
	fail := true
	o := &dragOutput{button: 272, pointerWrite: func(e event) error {
		events = append(events, e)
		if fail {
			fail = false
			return errors.New("injected write failure")
		}
		return nil
	}}
	if err := o.apply(dragEffect{press: true}, nil); err == nil || !o.held {
		t.Fatal("failed press lost cleanup obligation")
	}
	if err := o.release(); err != nil {
		t.Fatal(err)
	}
	if o.held || len(events) != 3 || events[1] != (event{1, 272, 0}) || events[2] != (event{0, 0, 0}) {
		t.Fatal(events)
	}
	if err := o.release(); err != nil || len(events) != 3 {
		t.Fatal("duplicate release")
	}
}
