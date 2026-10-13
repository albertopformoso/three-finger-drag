//go:build linux && cgo

package main

import "testing"

func TestTransferCut(t *testing.T) {
	t.Run("one and two contacts transfer without button events", func(t *testing.T) {
		for _, n := range []int{1, 2} {
			cut := newTransferCut(5, 0, []uint16{330, 325, 333, 334, 335, 328})
			if got := cut.route([]event{{0, 0, 0}}, n, n, 0); len(got) == 0 {
				t.Fatal("ordinary input suppressed")
			}
			got := cut.route([]event{{3, 47, 2}, {3, 57, 99}, {0, 0, 0}}, 3, 3, 0)
			ups := 0
			for _, e := range got {
				if e.kind == 1 && e.code == 272 {
					t.Fatal("handoff generated a button")
				}
				if e.kind == 3 && e.code == 57 && e.value == -1 {
					ups++
				}
			}
			if ups != 5 || got[len(got)-1] != (event{0, 0, 0}) {
				t.Fatal("incomplete contact termination")
			}
		}
	})
	t.Run("remain suppressed until neutral and restore slot", func(t *testing.T) {
		cut := newTransferCut(5, 0, nil)
		cut.route(nil, 3, 3, 0)
		if got := cut.route([]event{{3, 47, 4}}, 1, 1, 0); len(got) != 0 {
			t.Fatal("premature return")
		}
		cut.route(nil, 0, 0, 1)
		if !cut.blocked {
			t.Fatal("physical button bypassed neutral check")
		}
		cut.route(nil, 0, 0, 0)
		got := cut.route([]event{{3, 57, 100}, {0, 0, 0}}, 1, 1, 0)
		if got[0] != (event{3, 47, 4}) || cut.stats.Handoffs != 1 || cut.stats.NeutralReturns != 1 {
			t.Fatal("incorrect slot or lifecycle counts")
		}
	})
	t.Run("palm does not initiate handoff", func(t *testing.T) {
		cut := newTransferCut(5, 0, nil)
		if got := cut.route([]event{{0, 0, 0}}, 2, 3, 0); len(got) == 0 || cut.blocked {
			t.Fatal("palm counted as valid finger")
		}
	})
	t.Run("preserve physical button sequence", func(t *testing.T) {
		cut := newTransferCut(5, 0, nil)
		cut.route([]event{{1, 272, 1}, {0, 0, 0}}, 1, 1, 1)
		cut.route(nil, 3, 3, 0)
		if cut.blocked {
			t.Fatal("existing physical action interrupted")
		}
		cut.route(nil, 0, 0, 0)
		cut.route(nil, 3, 3, 0)
		if !cut.blocked {
			t.Fatal("did not rearm after neutral")
		}
	})
}
