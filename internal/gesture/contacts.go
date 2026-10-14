package gesture

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

// ErrUnreliableFrame requires failure handling rather than gesture continuation.
var ErrUnreliableFrame = errors.New("contact state is unreliable")

// Validate checks already-converted policy values. Configuration decoding must
// also reject overflow or underflow before conversion to these types.
func (p Policy) Validate() error {
	if !finite(p.StartThresholdMM) || p.StartThresholdMM <= 0 {
		return errors.New("validate start threshold: must be finite and greater than zero")
	}
	if p.GracePeriod <= 0 {
		return errors.New("validate grace period: must be greater than zero")
	}
	return nil
}

// ValidContacts returns an owned, ID-sorted snapshot with palms excluded.
// Malformed frames are rejected as a whole. No contact-count debounce is used.
func ValidContacts(frame Frame) ([]Contact, error) {
	if !frame.Reliable {
		return nil, ErrUnreliableFrame
	}
	if frame.At < 0 {
		return nil, fmt.Errorf("validate frame time: %w", ErrUnreliableFrame)
	}
	seen := make(map[int64]struct{}, len(frame.Contacts))
	valid := make([]Contact, 0, len(frame.Contacts))
	for _, c := range frame.Contacts {
		if c.ID < 0 || !finite(c.XMM) || !finite(c.YMM) {
			return nil, fmt.Errorf("validate contact identity or position: %w", ErrUnreliableFrame)
		}
		if _, exists := seen[c.ID]; exists {
			return nil, fmt.Errorf("validate duplicate contact identity: %w", ErrUnreliableFrame)
		}
		seen[c.ID] = struct{}{}
		if !c.Palm {
			valid = append(valid, c)
		}
	}
	slices.SortFunc(valid, func(a, b Contact) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return valid, nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
