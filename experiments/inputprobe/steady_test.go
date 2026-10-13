//go:build linux && cgo

package main

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

func fingers(x float64) []contact {
	return []contact{{id: 10, x: x}, {id: 11, x: x}, {id: 12, x: x}}
}

func TestSteadyWindow(t *testing.T) {
	t.Run("exclude placement and end boundary", func(t *testing.T) {
		w := newSteadyWindow(5*time.Second, 15*time.Second)
		w.observe(time.Second, fingers(100))
		w.observe(5*time.Second, fingers(0))
		w.observe(6*time.Second, fingers(0.5))
		w.observe(15*time.Second, nil)
		if w.result.Frames != 2 || w.result.CountChanges != 0 || w.result.MinNonPalm != 3 || w.result.MaxDisplacementMM != 0.5 {
			t.Fatalf("unexpected window: %+v", w.result)
		}
	})
	t.Run("fixed baseline does not accumulate travel", func(t *testing.T) {
		w := newSteadyWindow(0, 10*time.Second)
		for i, x := range []float64{0, 0.6, 0, 0.6, 0} {
			w.observe(time.Duration(i)*time.Second, fingers(x))
		}
		if math.Abs(w.result.MaxDisplacementMM-0.6) > 1e-9 {
			t.Fatal(w.result)
		}
	})
	t.Run("track replacements without replacement jump", func(t *testing.T) {
		w := newSteadyWindow(0, 10*time.Second)
		w.observe(0, fingers(0))
		changed := fingers(100)
		changed[0].id = 20
		w.observe(time.Second, changed)
		if w.result.IdentityChanges != 1 || w.result.Segments != 2 || w.result.MaxDisplacementMM != 0 {
			t.Fatal(w.result)
		}
	})
	t.Run("palm exclusion and slot reorder", func(t *testing.T) {
		w := newSteadyWindow(0, 10*time.Second)
		w.observe(0, fingers(0))
		c := fingers(0)
		c[0], c[2] = c[2], c[0]
		c = append(c, contact{id: 99, x: 1000, palm: true})
		w.observe(time.Second, c)
		if w.result.CountChanges != 0 || w.result.IdentityChanges != 0 || w.result.MixedFrames != 1 || w.result.MaxDisplacementMM != 0 {
			t.Fatal(w.result)
		}
	})
	t.Run("count loss resets segment", func(t *testing.T) {
		w := newSteadyWindow(0, 10*time.Second)
		w.observe(0, fingers(0))
		w.observe(time.Second, fingers(0)[:2])
		w.observe(2*time.Second, fingers(100))
		if w.result.CountChanges != 2 || w.result.MinNonPalm != 2 || w.result.Segments != 2 || w.result.MaxDisplacementMM != 0 {
			t.Fatal(w.result)
		}
	})
	t.Run("serialize aggregates only", func(t *testing.T) {
		w := newSteadyWindow(0, time.Second)
		w.observe(0, fingers(123.456))
		b, err := json.Marshal(w.result)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "123.456") || strings.Contains(string(b), "previousIDs") {
			t.Fatalf("input state leaked: %s", b)
		}
	})
}
