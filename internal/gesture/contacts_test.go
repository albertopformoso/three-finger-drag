package gesture

import (
	"errors"
	"math"
	"testing"
	"time"
)

// fixtureClock has no sleeps, timers, goroutines, or platform dependencies.
type fixtureClock struct{ at time.Duration }

func (c *fixtureClock) advance(d time.Duration) { c.at += d }
func (c *fixtureClock) frame(contacts ...Contact) Frame {
	return Frame{At: c.at, Reliable: true, Contacts: contacts}
}

func TestPolicyValidate(t *testing.T) {
	for _, tt := range []struct {
		name      string
		threshold float64
		grace     time.Duration
		valid     bool
	}{
		{"defaults", 1, 500 * time.Millisecond, true},
		{"fractional threshold", 0.125, time.Nanosecond, true},
		{"zero threshold", 0, time.Second, false},
		{"negative threshold", -1, time.Second, false},
		{"infinite threshold", math.Inf(1), time.Second, false},
		{"nan threshold", math.NaN(), time.Second, false},
		{"zero grace", 1, 0, false},
		{"negative grace", 1, -time.Second, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := (Policy{tt.threshold, tt.grace}).Validate()
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v, error=%v", tt.valid, err)
			}
		})
	}
}

func TestValidContacts(t *testing.T) {
	t.Run("exclude palm and retain thumb without aliasing", func(t *testing.T) {
		// The thumb uses the same non-palm representation as another finger.
		input := []Contact{{ID: 3, XMM: 3}, {ID: 1, XMM: 1}, {ID: 4, Palm: true}, {ID: 2, XMM: 2}}
		got, err := ValidContacts(Frame{Reliable: true, Contacts: input})
		if err != nil || len(got) != 3 {
			t.Fatal(got, err)
		}
		for i, c := range got {
			if c.ID != int64(i+1) {
				t.Fatal("unsorted contacts", got)
			}
		}
		got[0].XMM = 999
		if input[1].XMM != 1 || input[0].ID != 3 {
			t.Fatal("input modified")
		}
	})
	t.Run("invalid frame is never partially accepted", func(t *testing.T) {
		for _, tt := range []struct {
			name  string
			frame Frame
		}{
			{"unreliable", Frame{}},
			{"negative time", Frame{Reliable: true, At: -1}},
			{"duplicate identity", Frame{Reliable: true, Contacts: []Contact{{ID: 1}, {ID: 1, Palm: true}}}},
			{"negative identity", Frame{Reliable: true, Contacts: []Contact{{ID: -1}}}},
			{"nonfinite coordinate", Frame{Reliable: true, Contacts: []Contact{{ID: 1, XMM: math.Inf(1)}}}},
		} {
			t.Run(tt.name, func(t *testing.T) {
				got, err := ValidContacts(tt.frame)
				if got != nil || !errors.Is(err, ErrUnreliableFrame) {
					t.Fatal(got, err)
				}
			})
		}
	})
	t.Run("brief fourth contact takes effect immediately", func(t *testing.T) {
		clock := fixtureClock{}
		contacts := []Contact{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}
		for _, count := range []int{3, 4, 3, 0} {
			clock.advance(time.Microsecond)
			frame := clock.frame(contacts[:count]...)
			got, err := ValidContacts(frame)
			if err != nil || len(got) != count {
				t.Fatal(got, err)
			}
		}
		if clock.at != 4*time.Microsecond {
			t.Fatal("incorrect synthetic time")
		}
	})
}
