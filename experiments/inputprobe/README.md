# Input Backend Probe

This Linux/cgo experiment uses Go's standard library and the system libevdev library.
It is a bounded feasibility probe for T15 onward, not the three-finger drag service.
Native code stays behind the Go wrapper. No third-party Go module is required.

## Build and checks

Require Go 1.27.1 or later, a C compiler, pkg-config, and libevdev 1.14.0 or later.

```bash
make fmt
go test -race ./...
make probe
```

The executable is `/tmp/opencode/three-finger-inputprobe`.
Default duration is 10 seconds. Accepted durations range from 1 through 60 seconds.
SIGINT and SIGTERM request cleanup. Native reads poll in bounded intervals.
SIGKILL bypasses process cleanup. Kernel and compositor cleanup after SIGKILL still require external observation.

## T15: Synthetic lifecycle

Run in a host terminal with permission to create uinput devices:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=create -kind=pointer -duration=15s
```

This command opens no physical input device and emits no motion or button events.
It prints the synthetic device identity, waits, and destroys the device.
While it runs, inspect only the probe devices from another terminal:

```bash
hyprctl devices -j | jq '[.mice[] | select(.name | startswith("tfd-probe-")) | {name}]'
```

Repeat that query after exit. Record whether the probe device appeared and disappeared.
The creation result alone does not prove desktop recognition or teardown.

The opt-in native lifecycle test creates and destroys a real synthetic pointer:

```bash
go test -c -tags=integration -o /tmp/opencode/inputprobe.test ./experiments/inputprobe
sudo env TFD_UINPUT_TEST=1 /tmp/opencode/inputprobe.test -test.run '^TestSyntheticLifecycle$' -test.v
```

Without that environment variable, integration tests skip this device test explicitly.
The test does not verify compositor button cleanup.

## Probe modes

| Mode | Purpose | Physical input | Output |
| --- | --- | --- | --- |
| `inspect` | Copy static capabilities | Explicit touchpad only, no grab | Identity, slots, axis limits and resolution |
| `create` | Synthetic device lifecycle | None | Identity metadata and optional synthetic pointer action |
| `grab` | Bounded exclusive ownership | Explicit touchpad, no forwarding | Aggregate contact statistics |
| `proxy` | Ordinary-input forwarding comparison | Explicit touchpad, forwarded through a virtual touchpad | Identity and aggregate statistics |
| `classify` | Observe firmware contact classification | Explicit touchpad, no grab | Aggregate statistics only |

Physical modes require `-device` explicitly. The probe rejects virtual probe devices and non-multitouch devices.
`grab` temporarily suppresses the selected touchpad. `proxy` temporarily changes its route through Hyprland.
Both require zero contacts and released physical buttons at acquisition.
Use the keyboard to interrupt a probe. The duration also bounds normal operation.
Do not run physical probes concurrently.

No mode prints coordinates, tracking IDs, event timestamps, or individual input events.
Only static metadata and aggregate counts leave the process.
Classification reports raw firmware palm tool values. It does not implement a palm detector or recognize thumbs.
`frames_with_three_non_palm_and_palm` counts complete frames with exactly three non-palm contacts and at least one reported palm.
This counter establishes simultaneous classification without printing individual frames or contact identities.
No reported palm values does not prove that palms never occurred.

## T16 and T17: Recognition

Create a synthetic touchpad:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=create -kind=touchpad -duration=20s
```

Check the returned event node with metadata diagnostics and confirm Hyprland device recognition.
Record capabilities, axis ranges, resolution, seat, and effective settings.
The synthetic description uses five slots and the recorded physical device's axis ranges and resolution.
It does not claim to copy hardware quirks or firmware identity.

For pointer-motion delivery:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=create -kind=pointer -duration=10s -settle=5s -motion
```

The probe submits one 20-unit horizontal relative movement after the settle interval.
Observe pointer movement independently. No button is pressed by this command.
The settle interval allows inspection but does not prove desktop readiness.

The separate `-hold` flag presses physical `BTN_LEFT` after settling and releases it during normal cleanup.
This is an output-path experiment, not the logical-primary mapping required by the final utility.
Test it only with the pointer over a disposable test window.
Do not use submission success as evidence of a desktop button transition.

## T18 through T20: Ownership and ordinary input

Inspect the selected touchpad first:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=inspect -device=/dev/input/event10
```

Run a bounded ownership check with fingers lifted:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=grab -device=/dev/input/event10 -duration=10s
```

During the grab, confirm the selected touchpad stops moving the desktop pointer.
After exit, confirm ordinary pointing returns. Other devices must retain their ordinary behavior.
For ordinary-input comparison, run:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=proxy -device=/dev/input/event10 -duration=20s
```

Compare one-finger pointing and two-finger scrolling with the unproxied baseline, in separate runs.
Record effective virtual-device settings. Differences in settings or hardware quirks can change the result.
The proxy forwards full frames unchanged. It does not suppress or transform three-finger gestures.
On synchronization loss or write failure, this experiment exits and destroys its resources.
It does not implement the production pending-release confirmation mechanism.

## T21 through T25: Transfer and drag motion

### Actual drag-transfer experiment

October 13 update: after all valid contacts lift, a fresh one-finger return immediately releases the drag and restores pointing.
Three fingers returning together within 500 ms continue the hold. A return first reported as one finger ends it instead.
There is no ending-tap duration or displacement test. The result counter is now `single_finger_ends`.

`drag` mode adds a synthetic pointer hold to the observed contact-termination strategy:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=drag -device=/dev/input/event9 -duration=30s -settle=5s
```

Lift all contacts until READY. Test only on disposable Mousepad text first.
Place three fingers, move beyond 1 mm, and continue moving to select text without pressing.
Lift all fingers and wait longer than 500 ms. Confirm selection stops extending and ordinary pointing returns.
Lift contacts before the bounded probe ends. Ctrl+C also ends the experiment and attempts release.

This mode uses fixed 1 mm start recognition, 500 ms grace, and ending-tap limits of 200 ms and 1 mm.
Returning exactly three valid contacts before expiry immediately resumes the hold, even while stationary.
The new contact position becomes the motion baseline. Subsequent movement requires no continuation threshold.
Start threshold-crossing movement and continuation return displacement are discarded.
Physical button presses terminate the experimental gesture, with input suppressed until a neutral return.
It emits aggregate start, continuation, release-reason, and movement-frame counters only.

Output scaling uses experimental `-gain=12` relative units/mm followed by the virtual pointer's compositor acceleration.
It is not verified to match ordinary touchpad motion. Do not use subjective success as T25 numerical evidence.
The default output button is physical left. Use `-button=right` only when testing a verified swapped desktop primary mapping.
The harness does not automatically discover logical-primary mapping.

The timer loop uses nonblocking reads with bounded native polling. It is not a latency acceptance result.
The probe attempts release on error, signal, and deadline, then tears down virtual devices.
It does not implement compositor-confirmed pending cleanup, automatic session recovery, or lock/suspend notification handling.
T26 and T27 need independent teardown observations in an isolated test session before those behaviors can be claimed.

### First isolate contact termination

The `transfer` mode forwards ordinary input until at least three non-palm contacts appear.
It then sends an all-contacts-up frame to the virtual touchpad and suppresses input until all contacts and buttons lift.
It generates no synthetic drag button or pointer movement. This isolates whether terminating ordinary input causes an unwanted click or scroll.

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=transfer -device=/dev/input/event9 -duration=30s
```

For T21, move one finger over disposable Mousepad text, then add two fingers without clicking.
Observe any caret movement or selection caused by the handoff, and any pointer jump.
Lift everything and confirm ordinary pointing returns. Repeat with a short and a longer initial one-finger contact.
For T22, start two-finger scrolling in a long disposable document, then add a third finger.
Observe whether scrolling stops cleanly or leaves motion, a click, or a competing desktop gesture.

Keep existing tap and scroll settings enabled. Report a failed handoff rather than masking it by changing ordinary-input settings.
If either handoff produces an unwanted action, stop before introducing a synthetic button hold and revise the routing strategy.
This probe checks an intermediate transition, not a complete drag gesture or T21/T22 completion by itself.

These experiments require the successful recognition and ordinary-input evidence above first.
The contact-only mode isolates contact termination. The separate drag mode adds a bounded experimental gesture.
Do not mark these tasks complete using proxy-only, contact-termination-only, or unit-test results.

Required next probes:

1. Transfer one-contact pointing into three-contact drag without an extra click or jump.
2. Transfer two-contact scrolling into drag without leaving scrolling active.
3. Consume a fresh ending tap, including its lift after grace expiry.
4. Continue a drag after repositioning and return to ordinary input without stale contacts.
5. Compare group-motion output with ordinary pointing using T10's distance and timing criteria.

If contact termination produces a tap, record a failed experiment and revise the routing design.
Do not hide that result by disabling unrelated ordinary pointing features.

### T25 measured motion comparison

The `travel` mode measures ordinary physical one-finger pointing without grabbing input.
It reads contact displacement in millimeters and queries the Hyprland command socket for cursor endpoints.
Only derived displacement, speed, query duration, and per-stroke aggregates are printed. Absolute cursor positions and contact paths remain in memory.

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=travel -device=/dev/input/event9 -duration=30s -settle=5s -cursor-socket="$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock"
```

Keep fingers lifted until READY. Make separate straight one-finger strokes, lifting and pausing at least one second between strokes.
Use short strokes near the display center to avoid pointer clipping. Include slow, medium, and fast motions with similar touchpad distances.
Use no other pointing device during a measurement. Reposition the pointer only between measured runs.

For the matching drag measurement, use the same socket flag with `-mode=drag`.
The drag measurement begins after the start threshold, excluding discarded recognition movement.
Lift the contacts and wait at least one second before the next stroke.
The moving segment ends when valid contacts first drop below three. Staggered lifting is accepted because drag motion stops then.
Adding a fourth valid contact or pressing a physical button marks the stroke interrupted.
Record the active gain and compositor settings with results.

Each stroke waits 200 ms after lift before its final cursor query to allow output to settle.
This wait and the synchronous cursor queries perturb execution. These measurements are for motion comparison, not latency acceptance.
The first query can overlap finger movement. Query time is reported, but is not a complete uncertainty bound.
Unmatched physical travel, inconsistent speed, edge clipping, other pointer activity, and observer overlap make comparison inconclusive.
Inspect input path versus displacement to identify curved or reversed strokes.
Compare matched travel and speed samples against T10's distance tolerance only after accounting for measurement uncertainty.
No pass/fail motion-equivalence conclusion is automatically generated from these manual strokes.

## T26 and T27: Cleanup and session loss

Use the synthetic pointer hold in a disposable test window to observe normal release and device removal.
An external observer must verify button state after forced termination.
Physical ownership restoration needs its own observation after process termination.
These are separate results from the synthetic lifecycle unit/integration checks.

Coordinate session-loss testing with the user after teardown evidence exists.
Use an isolated test session. Do not terminate the user's working compositor as an unattended test.
Record whether old output reaches another session and whether the original button state clears.
The harness does not reconnect or prove T05 cleanup confirmation.

## T28: Contact classification

### Controlled stationary window

Use the `steady` mode for placement-excluding measurements:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=steady -device=/dev/input/event9 -duration=15s -settle=5s
```

Place contacts during `PLACE CONTACTS`. Hold them still through `MEASURING` until `DONE`, then lift.
Run three fingertips, thumb plus two fingertips, and three fingertips plus palm as separate labeled scenarios.
Repeat each scenario three times. Report accidental movement separately.

The probe reads without grabbing or generating input.
Measurement uses complete-frame kernel monotonic timestamps within the window, excluding placement and the ending boundary.
Output contains only aggregate counts, timing coverage, and maximum group displacement in millimeters.
The first measured frame supplies the baseline. Placement transitions do not contribute to measured count changes.
Three-contact identity changes or count departures start a fresh displacement segment and are counted separately.
The maximum displacement is relative to each segment's fixed starting average position, not accumulated path length.
Compare it with the 1 mm recognition threshold, alongside count and identity changes.
Check measurement completion, frame count, first/last offsets, and gaps before drawing conclusions.
No frames means inconclusive. This experiment does not prove behavior on untested palm placements or devices.

### Whole-run contact summaries

Run separate bounded scenarios for three fingers, four fingers, a thumb, and a palm:

```bash
sudo /tmp/opencode/three-finger-inputprobe -mode=classify -device=/dev/input/event10 -duration=10s
```

Record only the scenario label and aggregate result.
Compare intended contact counts with maximum observed contacts and reported palm-tool frames.
Repeated frames with contact-count changes can identify a scenario needing further noise analysis, but do not establish its cause.
Do not infer thumb classification or stable identities from aggregate counts alone.
