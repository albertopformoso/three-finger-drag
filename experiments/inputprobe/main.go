//go:build linux && cgo

// Command inputprobe runs bounded input-backend experiments, not the drag service.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type options struct {
	mode         string
	device       string
	kind         string
	duration     time.Duration
	settle       time.Duration
	motion       bool
	hold         bool
	gain         float64
	button       string
	cursorSocket string
}

func parse(args []string, stderr io.Writer) (options, error) {
	var o options
	f := flag.NewFlagSet("inputprobe", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&o.mode, "mode", "inspect", "inspect, create, grab, proxy, transfer, drag, classify, steady, or travel")
	f.StringVar(&o.device, "device", "", "explicit physical touchpad path")
	f.StringVar(&o.kind, "kind", "touchpad", "synthetic device: touchpad or pointer")
	f.DurationVar(&o.duration, "duration", 10*time.Second, "probe lifetime (1s to 60s)")
	f.DurationVar(&o.settle, "settle", 3*time.Second, "virtual recognition wait, or placement interval in steady mode")
	f.BoolVar(&o.motion, "motion", false, "emit a small synthetic pointer movement after settling")
	f.BoolVar(&o.hold, "hold", false, "hold physical BTN_LEFT after settling until cleanup (pointer only)")
	f.Float64Var(&o.gain, "gain", 12, "experimental drag relative units per millimeter (0 to 100, exclusive of zero)")
	f.StringVar(&o.button, "button", "left", "drag output button: left or right; match the desktop's primary mapping")
	f.StringVar(&o.cursorSocket, "cursor-socket", "", "Hyprland command socket for travel mode or optional drag stroke measurement")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	if f.NArg() != 0 {
		return o, errors.New("arguments: unexpected positional arguments")
	}
	if o.duration < time.Second || o.duration > time.Minute {
		return o, errors.New("duration: must be between 1s and 60s")
	}
	if o.settle < 0 || o.settle >= o.duration {
		return o, errors.New("settle: must be nonnegative and less than duration")
	}
	switch o.mode {
	case "inspect", "grab", "proxy", "transfer", "drag", "classify", "steady", "travel":
		if o.device == "" {
			return o, errors.New("device: an explicit touchpad path is required")
		}
	case "create":
		if o.device != "" {
			return o, errors.New("device: create mode must not open physical input")
		}
	default:
		return o, errors.New("mode: expected inspect, create, grab, proxy, transfer, drag, classify, or steady")
	}
	if o.kind != "touchpad" && o.kind != "pointer" {
		return o, errors.New("kind: expected touchpad or pointer")
	}
	if o.mode != "create" && o.kind != "touchpad" {
		return o, errors.New("kind: physical touchpad modes require kind touchpad")
	}
	if (o.motion || o.hold) && (o.mode != "create" || o.kind != "pointer") {
		return o, errors.New("motion/hold: only valid for create mode with kind pointer")
	}
	if math.IsNaN(o.gain) || math.IsInf(o.gain, 0) || o.gain <= 0 || o.gain > 100 {
		return o, errors.New("gain: must be finite, greater than zero, and at most 100")
	}
	if o.button != "left" && o.button != "right" {
		return o, errors.New("button: expected left or right")
	}
	if o.mode == "travel" && o.cursorSocket == "" {
		return o, errors.New("travel: cursor-socket is required")
	}
	if o.cursorSocket != "" && o.mode != "travel" && o.mode != "drag" {
		return o, errors.New("cursor-socket: only valid in travel or drag mode")
	}
	return o, nil
}

func main() {
	o, err := parse(os.Args[1:], os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err = run(ctx, o, os.Stdout)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func wait(ctx context.Context, duration time.Duration) bool {
	t := time.NewTimer(duration)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

// frameBuffer rejects synchronization loss and excessive frames before forwarding.
type frameBuffer struct{ events []event }

func (b *frameBuffer) add(e event) ([]event, error) {
	if e.kind == 0 && e.code == 3 {
		b.events = nil
		return nil, errors.New("frame: synchronization lost")
	}
	if len(b.events) >= 512 {
		b.events = nil
		return nil, errors.New("frame: exceeds 512 events")
	}
	b.events = append(b.events, e)
	if e.kind == 0 && e.code == 0 {
		result := b.events
		b.events = nil
		return result, nil
	}
	return nil, nil
}

type aggregate struct {
	Frames                        uint64 `json:"frames"`
	MaxContacts                   int    `json:"max_contacts"`
	FramesWithPalm                uint64 `json:"frames_with_palm_tool"`
	FramesWithThreeNonPalmAndPalm uint64 `json:"frames_with_three_non_palm_and_palm"`
	ContactCountChanges           uint64 `json:"contact_count_changes"`
	FramesWithButton              uint64 `json:"frames_with_physical_button"`
	previous                      int
}

func (a *aggregate) observe(contacts, palms, buttons int) {
	a.Frames++
	if contacts > a.MaxContacts {
		a.MaxContacts = contacts
	}
	if palms > 0 {
		a.FramesWithPalm++
		if contacts-palms == 3 {
			a.FramesWithThreeNonPalmAndPalm++
		}
	}
	if buttons > 0 {
		a.FramesWithButton++
	}
	if contacts != a.previous {
		a.ContactCountChanges++
	}
	a.previous = contacts
}

func run(parent context.Context, o options, stdout io.Writer) (result error) {
	ctx, cancel := context.WithTimeout(parent, o.duration)
	defer cancel()
	var p *physical
	if o.mode != "create" {
		var err error
		p, err = openPhysical(o.device)
		if err != nil {
			return err
		}
		defer p.close()
		if o.mode == "inspect" {
			return json.NewEncoder(stdout).Encode(p.info())
		}
		if o.mode == "steady" {
			return runSteady(ctx, p, o, stdout, os.Stderr)
		}
		if o.mode == "travel" {
			return runTravel(ctx, p, o, stdout, os.Stderr)
		}
		if o.mode == "drag" {
			return runDrag(ctx, p, o, stdout, os.Stderr)
		}
	}
	var v *virtual
	if o.mode == "create" || o.mode == "proxy" || o.mode == "transfer" {
		var err error
		name := fmt.Sprintf("tfd-probe-%s-%d", o.kind, os.Getpid())
		if o.mode == "proxy" || o.mode == "transfer" {
			name = fmt.Sprintf("tfd-probe-touchpad-%d", os.Getpid())
		}
		v, err = createVirtual(p, o.kind == "touchpad", name)
		if err != nil {
			return err
		}
		defer v.close()
		// This is identity metadata, not an input recording.
		if err := json.NewEncoder(stdout).Encode(map[string]string{"device": name, "node": v.node(), "result": "created; desktop recognition unverified"}); err != nil {
			return err
		}
		if !wait(ctx, o.settle) {
			return nil
		}
	}
	if o.mode == "create" {
		if o.hold {
			// Always attempt release even if the press or its SYN_REPORT fails.
			defer func() { result = errors.Join(result, v.write(event{1, 272, 0}), v.write(event{0, 0, 0})) }()
			if err := v.write(event{1, 272, 1}); err != nil {
				return err
			}
			if err := v.write(event{0, 0, 0}); err != nil {
				return err
			}
		}
		if o.motion {
			if err := v.write(event{2, 0, 20}); err != nil {
				return err
			}
			if err := v.write(event{0, 0, 0}); err != nil {
				return err
			}
		}
		<-ctx.Done()
		return nil
	}
	if o.mode == "grab" || o.mode == "proxy" || o.mode == "transfer" {
		// Drain pre-grab history and require a neutral state before transfer.
		for {
			if ctx.Err() != nil {
				return errors.New("acquire touchpad: deadline reached before neutral state")
			}
			_, ok, err := p.next()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
		}
		contacts, _, buttons := p.summary()
		if contacts != 0 || buttons != 0 {
			return errors.New("acquire touchpad: lift all fingers and release physical buttons first")
		}
		if err := p.grab(); err != nil {
			return err
		}
		// Recheck after acquisition to reject contacts arriving during the handoff.
		for {
			if ctx.Err() != nil {
				return errors.New("acquire touchpad: deadline reached during handoff")
			}
			_, ok, err := p.next()
			if err != nil {
				return err
			}
			if !ok {
				break
			}
		}
		contacts, _, buttons = p.summary()
		if contacts != 0 || buttons != 0 {
			return errors.New("acquire touchpad: contacts changed during handoff; probe stopped")
		}
		if _, err := fmt.Fprintln(stdout, `{"result":"grabbed; restoration and duplicate suppression unverified"}`); err != nil {
			return err
		}
	}
	var frames frameBuffer
	var counts aggregate
	defer func() { result = errors.Join(result, json.NewEncoder(stdout).Encode(counts)) }()
	var transfer *transferCut
	if o.mode == "transfer" {
		transfer = newTransferCut(p.info().Slots, p.currentSlot(), p.countKeys())
		defer func() { result = errors.Join(result, json.NewEncoder(stdout).Encode(transfer.stats)) }()
		fmt.Fprintln(os.Stderr, "TRANSFER PROBE: ordinary input stops at three contacts until all contacts and buttons lift. No drag button is generated.")
	}
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
		contacts, palms, buttons := p.summary()
		counts.observe(contacts, palms, buttons)
		if transfer != nil {
			// Reserve only valid non-palm contacts, as required by the specification.
			frame = transfer.route(frame, contacts-palms, contacts, buttons)
		}
		if v != nil {
			for _, item := range frame {
				if err := v.write(item); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
