# Technical Plan: Three-Finger Drag

## Status and scope

Intended repository location: `specs/001-three-finger-drag/plan.md`.

This draft covers RF-1 through RF-9 and the project constitution.
It contains design descriptions, illustrative JSON, and pseudocode. It contains no implementation code.

The specification remains the authority for observable behavior.
Technical proposals cannot resolve its open questions without an approved specification update.
The input backend remains provisional until its feasibility checks pass.

## 1. Specification decisions required before implementation

| Decision | Required clarification | Coverage |
| --- | --- | --- |
| Gesture recognition | Define valid contacts, movement thresholds, staggered contacts, and ending taps. | RF-2, RF-3, RF-7, OQ-2, OQ-3 |
| Timing | Define grace duration, restart rules, stationary contacts, and events at expiry. | RF-5, RF-6, RF-7, OQ-1, OQ-4 |
| Pointer behavior | Define motion scaling, acceleration, direction, and measurable response limits. | RF-3, RF-6, OQ-9 |
| Drag action | Confirm the logical primary-button action, including left-handed settings. | RF-2, RF-4, RF-7 |
| Failure state | Distinguish a terminated gesture from a button release that remains pending. | RF-8, RF-9 |
| Recovery | Define recovery, fresh-contact requirements, enablement, and disablement. | RF-2, RF-8, RF-9, OQ-5, OQ-6 |
| External actions | Define physical buttons, other pointing devices, cancellation, and lifecycle interruptions. | RF-3, RF-7, RF-8, RF-9, OQ-7 |
| Compatibility | Select touchpads, Hyprland versions, applications, window states, and the macOS reference. | RF-1, RF-4, OQ-8 |
| Configuration | Define public settings, defaults, validation limits, and invalid-configuration behavior. | RF-1, RF-5, RF-7, RF-9 |

RF-5 currently prescribes a timer start during situations where OQ-4 leaves timer restarts open.
RF-3 and RF-7 also require an explicit precedence rule at grace-period expiry.
RF-9 cannot define failure completion solely through the existing held-button definition of an active drag.

The specification must also address stationary active drags, repeated interruptions, contact noise, and contacts remaining after expiry.
It must define outcomes for process exit, session lock, suspend, device changes, and target-window disappearance.

The completion criteria must preserve the constitution's macOS-equivalence requirement.
User acceptance of unspecified differences does not establish equivalence.

## 2. Module structure

Use one Go module with Go 1.27.1 or later, as required by the repository.
Use small internal packages with one responsibility each.

| Proposed location | Responsibility | Coverage |
| --- | --- | --- |
| `cmd/three-finger-drag/` | Parse approved commands, start the application, and report English errors. | RF-1, RF-8, RF-9 |
| `internal/gesture/` | Recognize contacts and taps. Calculate motion. Apply drag transitions using supplied time and policy. | RF-2 through RF-7 |
| `internal/runtime/` | Order events, schedule timers, route effects, and coordinate cleanup. | RF-1 through RF-9 |
| `internal/input/` | Read complete device frames, track contacts, and detect device failures. | RF-1, RF-2, RF-3, RF-5, RF-7, RF-9 |
| `internal/output/` | Forward ordinary touchpad input and submit drag-button and pointer events. | RF-2 through RF-9 |
| `internal/platform/hyprland/` | Check the session, virtual-device routing, settings, and desktop availability. | RF-1, RF-4, RF-8, RF-9 |
| `internal/config/` | Read and validate explicit configuration. Save configuration atomically when requested. | RF-1, RF-5, RF-7, RF-9 |
| `internal/platform/linuxinput/` | Isolate native device operations and the proposed libevdev binding. | RF-1, RF-3, RF-8, RF-9 |
| `tests/integration/` | Exercise device adapters and Hyprland behavior with synthetic input. | RF-1 through RF-9 |

Place unit tests beside the packages they verify.
Store only authored synthetic fixtures. Never store captured user input.

### Dependency boundaries

Coverage: RF-1 through RF-9. Constitution principles 1, 3, and 6.

- The gesture package imports no device, display-server, configuration, or command-interface package.
- The gesture package accepts values and returns state changes plus abstract effects.
- The runtime supplies approved policy values and monotonic time.
- Platform adapters translate between operating-system events and domain values.
- The command interface assembles the application. It contains no gesture rules.
- Native types remain inside the Linux input adapter.
- All identifiers, comments, messages, documentation, and errors use English.

## 3. Input and output design

Coverage: RF-1 through RF-9.

### Proposed event path

1. Acquire the selected physical touchpad through an exclusive evdev grab.
2. Read complete contact frames through a narrow libevdev binding.
3. Supply normalized contact frames to the gesture logic.
4. Forward ordinary input through a virtual touchpad that Hyprland handles with libinput.
5. Suppress touchpad sequences owned by the drag gesture.
6. Submit drag-button actions and relative motion through a virtual pointer on the same seat.

The physical touchpad must have exactly one owner while the feature operates.
Otherwise, Hyprland could process the physical gesture and the generated drag simultaneously.
Do not capture unrelated keyboards or pointing devices.
Exclude the utility's virtual devices from device discovery to prevent feedback loops.

The virtual touchpad retains the physical device's relevant capabilities, axis ranges, and resolution.
The adapter must preserve ordinary one-finger pointing and two-finger scrolling.
The virtual pointer must retain button ownership throughout a drag and its grace period.

### Backend feasibility conditions

Coverage: RF-1, RF-2, RF-3, RF-6, RF-7, RF-8, RF-9.

This design requires more than dropping frames with three contacts.
Hyprland may already have received one-finger or two-finger input before the third finger arrives.
Synthetic contact termination could cause an unwanted tap, scroll, or stale contact.

Before adopting this backend, verify these conditions:

1. Hyprland recognizes both virtual devices and assigns them to the intended seat.
2. Ordinary virtual-touchpad behavior matches the selected physical-touchpad settings.
3. Gesture ownership can change without extra clicks, pointer jumps, stale contacts, or competing gestures.
4. The ending tap reaches the gesture logic but generates no additional desktop click.
5. Repositioned contacts can return to ordinary input without stale gesture state.
6. Relative drag motion avoids duplicate acceleration and preserves fractional movement.
7. Device teardown and process failure do not leave a held drag action in Hyprland.
8. Session loss and restoration cannot deliver stale movement or a delayed release into another session.

If these checks fail, revise the backend decision before implementation proceeds.
Do not claim that the proxy already preserves every libinput behavior.
The touchpad acceptance baseline must also settle palm rejection and partial-multitouch hardware support.

## 4. Data model

Coverage: RF-1 through RF-9. Constitution principles 3 and 5.

### In-memory records

| Record | Main fields | Rules and coverage |
| --- | --- | --- |
| Contact frame | Device ID, sequence, monotonic timestamp, contact IDs, positions, classification | Complete frames only. RF-2, RF-3, RF-5, RF-7. |
| Gesture policy | Start threshold, tap rules, grace duration, timer precedence, motion mapping | Values require specification approval. RF-2 through RF-7. |
| Gesture state | Phase, contact baseline, grace deadline, timer generation | No disk serialization. RF-2 through RF-7. |
| Output state | Desired button state, submitted button state, release pending, output-session identity | Keep failed release separate from gesture completion. RF-2, RF-7, RF-9. |
| Effect | Hold button, move pointer, release button, schedule timer, cancel timer, route input | Abstract commands without native handles. RF-2 through RF-9. |
| Availability | Enabled state, device access, output access, desktop availability | Determines readiness and failure handling. RF-1, RF-8, RF-9. |
| Explicit configuration | Schema version and approved user-selected settings | Only this record may persist. RF-1, RF-5, RF-7, RF-9. |

Use monotonic timestamps from one mapped clock domain.
Device timestamps and timer deadlines must be comparable before the runtime orders events.
Assign each timer a generation so canceled timers cannot affect later drags.

Proposed phases are `ready`, `dragging`, `grace`, and `unavailable`.
Keep pending output cleanup in a separate record.
This separation is a technical proposal pending the RF-9 specification correction.

### JSON example

Coverage: RF-1, RF-2, RF-8, RF-9.

This authored example describes an idle in-memory snapshot.
It is documentation data, not a configuration file, logging format, or recording feature.

```json
{
  "device_id": "synthetic-touchpad-1",
  "phase": "ready",
  "contact_count": 0,
  "contacts": [],
  "motion_baseline": null,
  "grace_deadline_ns": null,
  "timer_generation": 0,
  "desired_button_state": "released",
  "submitted_button_state": "released",
  "release_pending": false,
  "output_session_id": "synthetic-session-1",
  "availability": {
    "enabled": true,
    "device_access": true,
    "output_access": true,
    "desktop_available": true
  }
}
```

Kernel acceptance of an output write does not prove application receipt.
The model therefore records submitted output rather than claiming confirmed application state.
Integration tests must observe the resulting desktop behavior separately.

### Configuration persistence

Coverage: RF-1, RF-5, RF-7, RF-9. NFR-5 and NFR-6.

Use JSON through the Go standard library for approved explicit settings.
Use `$XDG_CONFIG_HOME/three-finger-drag/config.json` when `XDG_CONFIG_HOME` specifies an absolute directory.
Otherwise, use `$HOME/.config/three-finger-drag/config.json`.
Do not create a configuration file merely because the utility starts.

Reject unknown fields and unsupported schema versions with English errors.
The specification must define whether invalid configuration stops startup or selects an approved fallback.
Public timing settings remain pending OQ-1. This plan assigns no default grace duration.

Atomic save procedure:

1. Validate the complete explicit configuration.
2. Create a restricted temporary file in the destination directory.
3. Write the complete JSON document.
4. Flush the file and check the result.
5. Close the file and check the result.
6. Rename the temporary file over the destination.
7. Flush the containing directory and check the result.
8. Remove an unused temporary file after a failed save.

Before replacement, failures preserve the previous valid file.
After replacement, report durability failures without claiming that the previous version remains installed.
Never persist contacts, device frames, gesture state, pointer paths, or runtime snapshots.

## 5. Main functionality pseudocode

These procedures describe behavior, not executable code.
References to approved policy remain unresolved until the specification supplies that policy.

### Startup and readiness

Coverage: RF-1, RF-8, RF-9.

1. Read explicit configuration and validate approved settings.
2. Check the supported session, selected device, required access, and output capabilities.
3. Prepare virtual devices and verify desktop recognition before routing input.
4. Acquire exclusive input ownership through the verified backend procedure.
5. Apply the approved startup state.
6. If any step fails, release acquired resources and report the operation, cause, and available guidance.

### Event ordering

Coverage: RF-2 through RF-9.

1. Receive a complete contact frame, timer event, output result, or availability change.
2. Map its time into the runtime's monotonic clock domain.
3. Ignore stale timer generations and obsolete session events.
4. Apply the specification's approved precedence rules.
5. Pass domain values to the gesture logic.
6. Apply returned effects in order.
7. Record output submission results without assuming application receipt.
8. If an effect fails, enter the failure procedure.

### Start a drag

Coverage: RF-2, RF-3, RF-4, RF-8.

1. If the feature is not ready, apply the approved unavailable-state routing policy.
2. If the contact count differs from three, preserve permitted ordinary input.
3. If three contacts remain below the approved movement threshold, generate no drag or click.
4. If exactly three contacts satisfy the approved movement rule, claim the gesture through the verified routing procedure.
5. Submit the approved primary-button hold at the current pointer position.
6. If submission fails, enter the failure procedure.
7. Establish the contact baseline and enter the dragging phase.
8. Apply the approved initial-motion rule without an absolute pointer warp.

### Move and pause

Coverage: RF-3, RF-4, RF-5.

1. While three valid contacts move, calculate relative motion from the current baseline.
2. Submit relative motion without another button press.
3. If the contact count changes from three, stop gesture-driven pointer movement.
4. Enter the grace phase and schedule its approved deadline.
5. Keep the drag action held during the grace phase.
6. Apply approved stationary-contact and repeated-contact-change rules.

File managers, text widgets, and draggable title bars receive the same ordinary drag action.
The utility does not move files directly, edit selections, or command arbitrary window movement.

### Continue during the grace period

Coverage: RF-5, RF-6.

1. Track fresh contact baselines without moving the pointer.
2. If exactly three contacts satisfy the continuation rule before expiry, cancel the current timer generation.
3. Enter the dragging phase without releasing or pressing the button again.
4. Calculate future movement from the new baseline.
5. Exclude absolute contact displacement caused by finger repositioning.

### End a drag

Coverage: RF-7, RF-8.

1. Detect an approved ending tap or a valid grace-period expiry.
2. Consume the ending tap without forwarding a click.
3. Cancel the current timer generation.
4. Submit one button release for the held drag action.
5. If submission fails, enter the failure procedure.
6. Clear gesture baselines and apply the approved contact-rearming rule.
7. Restore ordinary input through the verified routing procedure.

### Handle failures

Coverage: RF-9, with RF-8 interaction.

1. Mark the current gesture terminated and reject its future movement effects.
2. Cancel its timer generation and discard queued movement.
3. Mark any held or uncertain button state for release.
4. If the current output session permits release, submit the release immediately.
5. Otherwise, retain only the in-memory cleanup obligation.
6. When release becomes possible, apply the approved session-specific cleanup rule.
7. Report the failed operation, cause, and available guidance without input-event details.
8. Apply the approved recovery policy without resuming the interrupted drag.

The specification must decide ordinary-input routing while cleanup remains pending.
No process can execute cleanup after an uncatchable termination.
Kernel device teardown and compositor behavior therefore require direct failure tests.

## 6. Technical decisions and discarded alternatives

| Decision | Justification | Discarded alternative | Coverage |
| --- | --- | --- | --- |
| One Go application with standard-library configuration, timing, and errors | Matches the minimal-stack principle. | General application frameworks add dependencies without required behavior. | RF-1 through RF-9 |
| Pure gesture logic with injected time | Makes transitions deterministic and independent of Linux devices. | Gesture rules inside device callbacks couple behavior to hardware timing. | RF-2 through RF-7 |
| One ordered runtime event loop | Keeps button, motion, and timer effects ordered. | Independent gesture handlers introduce avoidable races. | RF-2 through RF-9 |
| Provisional exclusive input proxy | Can reserve three-finger input and suppress the ending tap before desktop delivery. | A passive listener cannot stop duplicate desktop processing. | RF-2, RF-7, RF-8 |
| Narrow libevdev binding | Reuses Linux device handling, synchronization recovery, and virtual-device support. | Handwritten ioctl structures duplicate platform-sensitive work. | RF-1, RF-8, RF-9 |
| Virtual pointer for drag output | Uses ordinary desktop drag behavior for all three user operations. | Application-specific automation does not cover arbitrary applications. | RF-2, RF-3, RF-4, RF-7 |
| Relative pointer motion | Preserves the current pointer position and supports contact repositioning. | Absolute touchpad mapping causes unwanted pointer relocation. | RF-2, RF-3, RF-6 |
| Keep Hyprland's libinput path for ordinary gestures | Avoids reimplementing ordinary pointing and scrolling. | A complete custom touchpad interpreter expands scope and risks RF-8. | RF-1, RF-8 |
| JSON for explicit configuration | The Go standard library supports parsing and serialization. | YAML requires another parser. A database has no required role. | RF-1, RF-5, RF-7, RF-9 |
| Atomic XDG configuration writes | Satisfies the persistence principle and interrupted-save requirement. | In-place writes can leave partial configuration. | RF-1, RF-9, NFR-6 |
| In-memory gesture and cleanup state | Satisfies the prohibition on input-event persistence. | Event logs and replay recordings persist user input. | RF-2 through RF-9, NFR-5 |

The libevdev binding is the proposed native dependency. Its necessity is limited to Linux device access and event handling.
Keep binding code small and outside the gesture package.
Review its exact API surface and version before implementation.
The compiler and native build tools support that binding only.

### Why the existing libinput drag feature is insufficient

Coverage: RF-5, RF-6, RF-7, RF-8.

Current libinput documentation describes different finger-count rules:

- One remaining finger ends its drag and restores pointer movement.
- Two remaining fingers retain its drag.
- Returning one or two fingers after a full lift ends its drag.
- A subsequent one-finger or two-finger tap can produce a separate click.

RF-5 instead requires a paused grace period after every departure from three contacts.
RF-7 consumes the ending tap without an additional click.
Directly enabling the existing feature therefore does not satisfy this specification.

Public libinput gesture events also do not expose reliable physical contact counts in every situation.
Its documentation specifically warns against implementing tap recognition from hold gestures.
A separate libinput listener alone cannot establish every required contact transition or suppress compositor processing.

## 7. Testing strategy

### Deterministic unit tests

Coverage: RF-2 through RF-9. Constitution principle 4.

Use authored synthetic contact frames and an injected monotonic clock.
Advance time explicitly. Do not use real sleeps for gesture tests.
Assert observable effect sequences and state invariants rather than private implementation details.

| Coverage | Required checks |
| --- | --- |
| RF-2 | Start only with qualifying three-contact movement. Stationary contacts produce no generated click or drag. |
| RF-3 | Relative movement maintains one held button and does not issue repeated presses. |
| RF-4 | Start effects hold the button before drag motion and preserve the current pointer origin. |
| RF-5 | Zero, one, two, and additional contacts pause movement and apply the approved grace rules. |
| RF-6 | Continuation cancels stale deadlines, preserves the hold, and rejects repositioning displacement. |
| RF-7 | Expiry and ending taps release once. Ending taps produce no additional click. |
| RF-8 | Ordinary input routes only under the approved ownership rules. Competing three-finger output remains suppressed. |
| RF-9 | Failures cancel motion and timers. Pending releases survive recoverable output failure without resuming the drag. |

Test before, at, and after every approved timing boundary.
Test repeated contact changes, stationary contacts, delayed events, and contacts remaining after release.
Use generated synthetic sequences to check balanced button ownership and stale-timer rejection.
Keep generated events in memory.

### Adapter integration tests

Coverage: RF-1 through RF-9.

Run device tests in an isolated test session with synthetic virtual hardware.
Verify capabilities, exclusive ownership, complete frames, synchronization loss, hotplug, permission failure, and output-write failure.
Check that the utility ignores its own virtual devices.
Verify input-to-output ownership changes and delayed desktop device discovery.
Observe desktop results rather than treating successful kernel writes as proof of acceptance.

Verify preserved pointing and scrolling against the same session without the utility.
Check no extra click when the third finger arrives or an ending tap completes.
Check no competing desktop action while the feature owns three-finger gestures.
Verify device cleanup after graceful exit and forced process termination.

### Desktop acceptance matrix

Coverage: RF-1 through RF-9, NFR-1, and NFR-2.

Run the full acceptance suite on Arch Linux and CachyOS with the agreed Hyprland versions.
Use the touchpads and representative applications selected under OQ-8.
Include each approved window state and application type.

| Requirement | Desktop evidence |
| --- | --- |
| RF-1 | Recorded environment versions and results for both distributions. |
| RF-2 | Three-finger movement starts at the current pointer position without a physical press. |
| RF-3 | Pointer motion follows approved direction, scale, and response limits. |
| RF-4 | File drops, text selections, and title-bar window movement complete in representative applications. |
| RF-5 | Every specified contact-count change pauses movement while retaining the drag. |
| RF-6 | Repositioning continues the same drag without a pointer jump. |
| RF-7 | Expiry and taps end the drag without an additional click. |
| RF-8 | Ordinary pointing and scrolling remain available outside a drag. Competing actions do not occur. |
| RF-9 | Disconnection, permission loss, and desktop loss terminate the gesture and satisfy the approved release policy. |

Compare the same authored interaction scripts with the selected macOS reference.
Store scenario descriptions, aggregate timings, and pass/fail results only.
Do not capture or persist input-event streams.
Response-time acceptance remains blocked until OQ-9 defines its limits.

### Persistence and language tests

Coverage: RF-1, RF-9, NFR-5 through NFR-8. Constitution principles 5 and 6.

- Check XDG resolution, missing configuration, unknown fields, unsupported versions, and approved validation rules.
- Inject failures before writing, during writing, before replacement, and after replacement.
- Verify that readers never receive a partial configuration document.
- Verify that runtime activity creates no input recordings or gesture-state files.
- Check that error messages identify the operation and cause in English.
- Review identifiers, comments, and documentation for the English-language requirement.

### Required checks during implementation

Coverage: RF-1 through RF-9. Constitution principle 4.

1. Run `make fmt` after Go changes.
2. Run `go test ./...`.
3. Run `go test -race ./...` for runtime and adapter concurrency changes.
4. Run `go test -tags=integration ./...` in the prepared integration environment.
5. Run the required desktop acceptance scenarios on both distributions.
6. Add a regression test for every corrected defect.
7. Report unavailable integration prerequisites explicitly. Do not count skipped acceptance scenarios as passes.

## 8. Implementation order and completion

| Stage | Output | Coverage |
| --- | --- | --- |
| 1 | Approved specification corrections, behavior rules, and compatibility baseline. | RF-1 through RF-9 |
| 2 | Evidence that the provisional input backend satisfies its feasibility conditions. | RF-1, RF-2, RF-3, RF-6, RF-7, RF-8, RF-9 |
| 3 | Gesture model and deterministic behavioral tests. | RF-2 through RF-7 |
| 4 | Verified input routing, output handling, and failure cleanup. | RF-1 through RF-9 |
| 5 | Approved configuration, command interface, and atomic persistence. | RF-1, RF-8, RF-9, NFR-5 through NFR-8 |
| 6 | Distribution acceptance results, macOS comparison, and regression coverage. | RF-1 through RF-9 |

Implementation starts only after the required behavior decisions enter the active specification.
The final backend must pass RF-8 without weakening ordinary gesture preservation.
Every RF requires acceptance evidence. A planned test is not evidence of a passing requirement.

When repository files are added, update the project tree in `README.md`.
After code changes, run the required tests and `graphify update .`.

## 9. Technical references

Coverage: RF-1, RF-5, RF-7, RF-8, RF-9.

- [libinput three-finger drag behavior](https://wayland.freedesktop.org/libinput/doc/latest/drag-3fg.html)
- [libinput gesture and hold-event limitations](https://wayland.freedesktop.org/libinput/doc/latest/gestures.html)
- [Linux uinput and libevdev guidance](https://docs.kernel.org/input/uinput.html)

These references establish API capabilities and limitations. They do not prove this backend works with the selected Hyprland release.
