//go:build linux && cgo

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestTravelMeter(t *testing.T) {
	t.Run("distance speed and observer endpoint", func(t *testing.T) {
		var output bytes.Buffer
		point := cursorPoint{500, 300}
		m := travelMeter{mode: "test", readCursor: func() (cursorPoint, error) { return point, nil }, output: &output}
		if err := m.begin(0, []contact{{id: 1, x: 10, y: 10}}); err != nil {
			t.Fatal(err)
		}
		m.observe(time.Second, []contact{{id: 1, x: 13, y: 14}})
		point = cursorPoint{506, 308}
		if err := m.finish(""); err != nil {
			t.Fatal(err)
		}
		if !m.r.Valid || m.r.InputDistance != 5 || m.r.AverageSpeed != 5 || m.r.PointerDistance != 10 || m.r.Gain != 2 {
			t.Fatal(m.r)
		}
		var record map[string]any
		if err := json.Unmarshal(output.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		if _, ok := record["startCursor"]; ok {
			t.Fatal("raw state leaked")
		}
	})
	t.Run("identity change rejected", func(t *testing.T) {
		m := travelMeter{readCursor: func() (cursorPoint, error) { return cursorPoint{}, nil }, output: &bytes.Buffer{}}
		m.begin(0, []contact{{id: 1}})
		if m.observe(time.Second, []contact{{id: 2, x: 100}}) {
			t.Fatal("replacement accepted")
		}
	})
	t.Run("stationary stroke inconclusive", func(t *testing.T) {
		m := travelMeter{readCursor: func() (cursorPoint, error) { return cursorPoint{}, nil }, output: &bytes.Buffer{}}
		m.begin(0, []contact{{id: 1}})
		m.observe(time.Second, []contact{{id: 1}})
		m.finish("")
		if m.r.Valid {
			t.Fatal("stationary measurement passed")
		}
	})
	t.Run("observer failure reported", func(t *testing.T) {
		m := travelMeter{readCursor: func() (cursorPoint, error) { return cursorPoint{}, errors.New("injected") }, output: &bytes.Buffer{}}
		if err := m.begin(0, []contact{{id: 1}}); err == nil || m.active {
			t.Fatal("observer failure ignored")
		}
	})
}
