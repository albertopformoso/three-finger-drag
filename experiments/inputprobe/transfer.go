//go:build linux && cgo

package main

// transferCut isolates the provisional all-contacts-up handoff strategy.
// It intentionally emits no drag-button actions or relative pointer motion.
type transferCut struct {
	slots          int
	keys           []uint16
	slot           int
	blocked        bool
	physicalAction bool
	stats          transferStats
}

type transferStats struct {
	Handoffs         uint64 `json:"contact_handoffs"`
	SuppressedFrames uint64 `json:"suppressed_frames"`
	NeutralReturns   uint64 `json:"neutral_returns"`
}

func newTransferCut(slots, slot int, keys []uint16) *transferCut {
	return &transferCut{slots: slots, slot: slot, keys: keys}
}

func (t *transferCut) route(frame []event, validContacts, allContacts, buttons int) []event {
	initialSlot := t.slot
	for _, e := range frame {
		if e.kind == 3 && e.code == 47 {
			t.slot = int(e.value)
		}
	}
	if t.blocked {
		t.stats.SuppressedFrames++
		if allContacts == 0 && buttons == 0 {
			t.blocked = false
			t.stats.NeutralReturns++
		}
		return nil
	}
	if buttons != 0 {
		t.physicalAction = true
	}
	if validContacts >= 3 && !t.physicalAction {
		t.blocked = true
		t.stats.Handoffs++
		t.stats.SuppressedFrames++
		return t.contactsUp()
	}
	if allContacts == 0 && buttons == 0 {
		t.physicalAction = false
	}
	// Suppressed frames can change the physical current slot without informing
	// the virtual device. Explicit selection prevents stale slot interpretation.
	out := make([]event, 0, len(frame)+1)
	out = append(out, event{3, 47, int32(initialSlot)})
	return append(out, frame...)
}

func (t *transferCut) contactsUp() []event {
	out := make([]event, 0, 2*t.slots+len(t.keys)+2)
	for slot := 0; slot < t.slots; slot++ {
		out = append(out, event{3, 47, int32(slot)}, event{3, 57, -1})
	}
	for _, key := range t.keys {
		out = append(out, event{1, key, 0})
	}
	out = append(out, event{3, 47, 0}, event{0, 0, 0})
	return out
}
