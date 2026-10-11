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
| Configuration | Define remaining settings and invalid-configuration behavior. T01 through T03 approve drag-start, ending-tap, and grace settings. | RF-1, RF-2, RF-5, RF-7, RF-9 |

The specification must also address contact noise.
Device-selection and switching details still require specification decisions.

The completion criteria must preserve the constitution's macOS-equivalence requirement.
User acceptance of unspecified differences does not establish equivalence.

### Approved T01 decisions

Coverage: RF-2 and RF-3. Specification reference: RF-2 and resolved OQ-3.

- Exclude detected palms and include thumbs in valid contacts.
- Use the arithmetic average of exactly three valid contact positions for group movement.
- Allow individual fingers to remain stationary.
- Allow staggered placement without a time limit.
- Establish the start baseline when the third valid contact arrives.
- Before drag start, discard the baseline when valid contact count or identities change.
- Establish a fresh baseline whenever exactly three valid contacts remain after that change.
- Measure Euclidean displacement from the fixed baseline in touchpad millimeters.
- Start when displacement reaches or exceeds the threshold. Do not accumulate traveled distance.
- Default to 1 mm. Accept configurable finite numeric thresholds greater than zero.
- Reject zero, negative, nonnumeric, and nonfinite thresholds.

The default still requires macOS comparison. T04 matches ordinary pointing and discards movement through the threshold-crossing event.
T11 defines invalid-configuration recovery and configuration changes during a drag.
T03 uses the same displacement threshold for grace-period continuation.

### Approved T02 decisions

Coverage: RF-7. Specification reference: ending-tap recognition and resolved OQ-2.

- Require a fresh single valid contact after all valid contacts lift. Include thumbs and ignore detected palms.
- Recognize the ending tap on that contact's lift, not its touchdown.
- Default to a maximum duration of 200 ms, including equality.
- Default to a maximum displacement of 1 mm from the initial contact position, including equality.
- Check the maximum displacement throughout contact. Returning after exceeding the limit does not restore the candidate.
- Cancel the candidate if its duration or displacement exceeds its limit, or a second valid contact appears.
- Require all valid contacts to lift before another attempt after cancellation.
- Configure duration and movement independently from the drag-start threshold.
- Accept only finite numeric limits greater than zero.
- Require touchdown and lift before grace expiry. Touchdown does not extend the grace deadline.
- Consume a recognized ending tap and release the held drag without an additional click.

Expiry takes priority over lift exactly at the grace deadline. Candidate cancellation alone does not release the drag.
T11 defines invalid-configuration recovery and configuration changes during a drag.
Both defaults require comparison with macOS.

### Approved T03 decisions

Coverage: RF-5, RF-6, and RF-7. Specification references: resolved OQ-1 and OQ-4.

- Default to a configurable grace duration of 300 ms. Accept only finite numeric durations greater than zero.
- Start the timer on the first departure from three valid contacts during dragging outside grace.
- Preserve the deadline through all further contact changes, including complete lifting and stationary three-contact returns.
- Establish a fresh continuation baseline when exactly three valid contacts return.
- Reset that baseline when valid contact count or identities change. Preserve the deadline.
- Use group displacement and the configured drag-start threshold for continuation.
- Resume only when that threshold is reached strictly before the deadline.
- Cancel the timer after successful continuation. A later interruption starts a new grace period.
- At the exact deadline, expiry takes priority over qualifying movement and candidate lift. Release once.

T04 discards movement through the continuation threshold-crossing event. T06 requires zero contacts before rearming after expiry.
T11 defines configuration changes during a drag and invalid-configuration recovery.
The 300 ms default requires comparison with macOS.

### Approved T04 decisions

Coverage: RF-2, RF-3, RF-4, and RF-6.

- Start or resume at the current pointer position. Discard movement through the threshold-crossing event, including any overshoot.
- Establish the motion baseline at that event's group position. Move only with subsequent group movement.
- Match ordinary one-finger pointing direction, sensitivity, and acceleration on the same touchpad.
- Preserve rightward and upward pointer directions. Ignore natural-scrolling direction for pointer movement.
- Hold the logical primary button, respecting the desktop's left-handed mapping.
- When contact identities change while the count stays three, keep the hold and reset the motion baseline.
- Discard contact-replacement displacement. Continue with subsequent movement without requiring another recognition threshold.
- While three contacts remain stationary outside grace, stop pointer motion and retain the hold indefinitely.
- Do not start a grace timer solely because active-drag contacts become stationary.

Stationary returns during grace still follow T03's fixed deadline and continuation threshold.
T10 sets a 5% or 2-logical-pixel distance tolerance and separate processing and desktop-response limits.
Backend feasibility must verify the ordinary-pointing match.

### Approved T05 decisions

Coverage: RF-8 and RF-9.

- Failure immediately terminates the gesture, stops its movement, cancels its timers, and prevents its resumption.
- Track pending button release separately from the terminated gesture.
- Suppress input from the affected touchpad while cleanup remains pending. Discard suppressed input without later replay.
- Leave other devices outside this restriction.
- Complete cleanup after button release or desktop confirmation that the owning virtual device or session no longer exists.
- Retain the obligation after unconfirmed loss. Never deliver an old release into a new session.
- Restore ordinary pointing and scrolling after cleanup when touchpad access is available.
- Restore ordinary input independently of new three-finger drag availability, following T06's readiness rules.
- Report the failed operation, cause, pending-release status, and available recovery guidance in English.
- Suppress repeated identical errors from unchanged cleanup failures.

The backend must establish reliable evidence for release or owner disappearance.
Kernel write acceptance alone does not prove desktop receipt. Missing confirmation keeps cleanup pending.
T10 defines measurable cleanup response limits. T08 defines process and session lifecycle behavior.

### Approved T06 decisions

Coverage: RF-2, RF-7, RF-8, and RF-9. Specification references: resolved OQ-5 and OQ-6.

- Request enabled operation by default at startup. Allow configuration to request disabled startup.
- Before arming, require available access, completed cleanup, and an observed zero-contact state.
- Recover automatically after failure when enabled and those prerequisites hold. Never resume a terminated gesture.
- Provide runtime enable and disable commands. Keep their overrides in memory until restart.
- Read startup enablement from configuration on restart. Runtime commands do not rewrite configuration.
- Disablement immediately terminates dragging or grace, cancels timers, and releases the held button.
- Apply T05 cleanup if release fails. Enablement and disablement cannot bypass pending cleanup.
- After normal completion, require zero valid contacts before another drag. Zero contacts at completion already satisfies this condition.
- Preserve ordinary pointing and scrolling while waiting for fresh contacts when access and cleanup permit.
- Once disabled and cleanup completes, return three-finger gestures to normal desktop handling.

Arming is a latch for the next fresh gesture, not a requirement that contacts remain absent during recognition.
Consume it when a drag starts. Clear it on failure or disablement.
Keep requested enablement separate from access, cleanup, armed state, and ordinary-input routing.

### Approved T07 decisions

Coverage: RF-3, RF-7, RF-8, and RF-9. RF-2's rearming criteria also require released physical buttons.

- A physical button press on the selected touchpad terminates dragging or grace and releases the utility's held button.
- Consume the terminating physical press and its matching release to avoid an additional click.
- Preserve ordinary physical-button actions that began outside a three-finger drag.
- Block recognition while a selected-touchpad physical button remains held.
- Rearm only with all selected-touchpad physical buttons released and zero valid contacts, plus T06's other readiness checks.
- Let the desktop process other pointing devices without interception or gesture termination.
- Recognize three-finger dragging on one selected touchpad. Never combine contacts across devices.
- Reject a second instance in the same desktop session with an error, without disturbing the existing instance.
- Direct enable, disable, and cancel commands to the existing instance.
- Cancellation stops movement, cancels timers, and releases the utility's hold while preserving requested enablement.
- Require fresh contacts after cancellation. Apply T05 if release fails.

Cancellation does not guarantee undo. Applications may treat release as a drop.
The backend must preserve consumed-press tracking until the matching release, including across cleanup and disablement.

### Approved T08 decisions

Coverage: RF-4, RF-8, and RF-9.

- Lock or suspend preparation terminates dragging or grace, stops movement, cancels timers, releases the held button, and blocks new drags.
- Apply T05 when cleanup remains pending. Unlock and resume require T06 readiness checks and fresh contacts.
- Normal exit stops movement and timers, releases the button, and restores normal touchpad handling before exit.
- Failed release during normal exit follows T05 cleanup. Do not report successful shutdown while cleanup remains unresolved.
- After crash or forced termination, the system must clear the utility's held button state and restore normal touchpad handling.
- A later restart follows startup configuration and fresh-contact checks. It never restores an old gesture.
- Session end or connection loss terminates the gesture and blocks new drags. Never send old input into another session.
- Recovery requires an available unlocked session, completed cleanup, and T06 readiness checks, including T07's released-button condition.
- Monitor changes preserve normal gesture rules while the session remains available.
- Accept desktop pointer and window repositioning. Continue relative movement without adding compensating movement.
- Target-window closure or application exit preserves normal gesture rules. Leave target handling to the desktop and application.

Target disappearance does not select a replacement target or guarantee undo.
System cleanup after process death requires observed integration evidence, not assumptions about finalizers or shutdown handlers.

### Approved T09 decisions

Coverage: RF-1, RF-4, and RF-8. Specification reference: resolved OQ-8 and RF-1's compatibility baseline.

- Support native Wayland and XWayland applications within Hyprland. Standalone X11 sessions remain outside the MVP.
- Require separate positions and stable identities for at least three contacts, plus detection of a fourth contact.
- Require physical axis resolution for millimeter measurements and sufficient contact information for palm rejection.
- Defer hardware that lacks any required capability. Do not infer compatibility from a model name alone.
- Use fully updated stable Arch Linux and CachyOS installations when acceptance testing begins.
- Record exact kernel, Hyprland, libinput, XWayland, and application versions. Keep them fixed for each acceptance run.
- Exclude development builds from the acceptance baseline.
- Use Thunar for file dragging and Mousepad for text selection and draggable title bars under both application backends.
- Verify the actual backend of each tested application instance.
- Test floating title-bar movement. Other window states follow ordinary Hyprland primary-button drag behavior without forced state changes.
- Use native macOS three-finger dragging on one compatible Mac, with fixed model, trackpad, macOS build, and settings.
- Use Finder for reference file dragging and TextEdit for reference text selection and title-bar movement.

T13 records actual hardware and software inventories before acceptance runs. T09 approves the selection policy, not test results.
Apple's guide establishes feature enablement. It does not supply numerical drag thresholds or prove compatibility by model year.
The macOS reference remains native behavior rather than another utility's approximation.

### Approved T10 decisions

Coverage: RF-2, RF-3, RF-6, RF-7, and RF-9. Specification reference: NFR-2 and resolved OQ-9.

- Measure utility processing delay separately from independently observed desktop response.
- Require processing p95 at most 5 ms, with no sample above 20 ms.
- Require desktop-response p95 at most 50 ms, with no sample above 100 ms.
- Start timing at the qualifying input event, termination command, or grace deadline specified in NFR-2.
- For recoverable cleanup, start timing when release becomes possible in the owning session.
- Measure continuation's state transition internally. Measure visible motion from the first subsequent movement event.
- Use at least 100 samples per operation, distribution, and application backend. Keep slow samples and calculate results separately.
- Use a warmed-up session without deliberate background workload. Fix power mode, refresh rate, and settings.
- Use monotonic timestamps and an independent desktop observer. Record precision and uncertainty.
- Mark results inconclusive if uncertainty could change pass to fail.
- Compare three-finger and ordinary one-finger motion on the same Linux touchpad using equivalent distances and speeds.
- Allow distance error up to the larger of 5% of the reference distance and 2 logical pixels. Require matching direction.
- Compare macOS gesture outcomes and review every mismatch before claiming equivalence. Report timing differences separately.
- Retain aggregates and measurement details only. Do not persist captured input-event streams or per-event runtime records.

These are approved targets, not benchmark results. The Linux latency limits remain independent of macOS timing measurements.

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
Select only one touchpad for gesture recognition. Leave other touchpads and pointing devices under normal desktop control.
Do not monitor other pointing devices to cancel this gesture. Their events do not alter its lifetime.

The virtual touchpad retains the physical device's relevant capabilities, axis ranges, and resolution.
The adapter must preserve ordinary one-finger pointing and two-finger scrolling.
The virtual pointer must retain button ownership throughout a drag and its grace period.
Map logical primary-button effects through the desktop's button mapping. Do not assume a physical left-button event always means primary.
Preserve the output button identity needed to release the held action.
Pass ordinary-touchpad motion settings as values across the platform boundary. Keep desktop queries outside the gesture package.
The adapter must reproduce ordinary-touchpad sensitivity and acceleration without applying acceleration twice.
Identical settings on a virtual mouse do not alone prove an equivalent touchpad response.

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
6. Relative drag motion matches ordinary-touchpad sensitivity and acceleration, avoids duplicate acceleration, and preserves fractional movement.
7. Device teardown and process failure do not leave a held drag action in Hyprland.
8. Session loss and restoration cannot deliver stale movement or a delayed release into another session.

Failure checks must verify suppression of the affected touchpad until release or confirmed owner disappearance.
Prove that suppression survives recoverable access failures without capturing unrelated devices.
Identify the evidence used to confirm cleanup. A disconnected control channel is not confirmation of owner disappearance.

If these checks fail, revise the backend decision before implementation proceeds.
Do not claim that the proxy already preserves every libinput behavior.
The touchpad acceptance baseline must verify palm rejection, thumb inclusion, and three independently tracked contacts with stable identities.
Require fourth-contact detection and physical axis resolution. Defer partial-multitouch devices that cannot satisfy these requirements.
It must also verify contact identities and axis resolution for millimeter measurements.
Verify logical primary-button behavior with both standard and left-handed mappings.

## 4. Data model

Coverage: RF-1 through RF-9. Constitution principles 3 and 5.

### In-memory records

| Record | Main fields | Rules and coverage |
| --- | --- | --- |
| Contact frame | Device ID, sequence, monotonic timestamp, contact IDs, positions, classification | Complete frames only. RF-2, RF-3, RF-5, RF-7. |
| Gesture policy | Shared start/continuation threshold, tap limits, grace duration, expiry precedence, ordinary-pointing motion values | T01 through T04 approve recognition, grace, and motion behavior. RF-2 through RF-7. |
| Gesture state | Phase, valid contact identities, start baseline, continuation baseline, motion baseline, grace deadline, timer generation | Keep recognition baselines fixed for each contact group. No disk serialization. RF-2 through RF-7. |
| Tap candidate | Contact identity, initial position, touchdown time, maximum displacement, cancellation state, all-contacts-lifted eligibility | Keep in memory. Reset after the required zero-contact state. RF-7. |
| Output state | Desired button state, submitted button state, held output button, release pending, owning device and session identities, cleanup evidence | Keep cleanup separate and session-bound. Submission alone is not desktop confirmation. RF-2, RF-4, RF-7, RF-9. |
| Effect | Hold button, move pointer, release button, schedule timer, cancel timer, route input | Abstract commands without native handles. RF-2 through RF-9. |
| Availability | Configured startup state, optional runtime override, requested enablement, armed state, access, desktop availability, ordinary-input routing | Keep requested enablement and readiness separate. RF-1, RF-2, RF-7, RF-8, RF-9. |
| Failure report state | Failed operation, cause, pending-release status, last reported error identity | Keep in memory. Suppress unchanged repeated errors without storing input events. RF-9. |
| Physical-button state | Selected device identity, held physical buttons, consumed press identities awaiting release | Preserve button pairing across gesture termination and cleanup. Keep in memory. RF-7, RF-8. |
| Instance ownership | Desktop-session identity and exclusive ownership handle | Reject duplicate startup before device acquisition. Runtime commands target the owner. RF-8. |
| Lifecycle state | Session identity, session availability, lock state, suspend state, shutdown requested | Gate drag readiness independently of requested enablement. Keep in memory. RF-4, RF-8, RF-9. |
| Explicit configuration | Schema version, startup enablement, drag-start threshold, ending-tap limits, grace duration, and other approved settings | Runtime overrides never persist. RF-1, RF-2, RF-5, RF-7, RF-9. |

Use monotonic timestamps from one mapped clock domain.
Device timestamps and timer deadlines must be comparable before the runtime orders events.
Assign each timer a generation so canceled timers cannot affect later drags.

Proposed phases are `disabled`, `unavailable`, `awaiting_contacts_up`, `ready`, `dragging`, and `grace`.
Keep pending output cleanup in a separate record.
T05 requires this separation. A terminated gesture may remain unavailable for new drags while ordinary input resumes after cleanup.

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
    "armed": true,
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

Coverage: RF-1, RF-2, RF-5, RF-7, RF-9. NFR-5 and NFR-6.

Use JSON through the Go standard library for approved explicit settings.
Use `$XDG_CONFIG_HOME/three-finger-drag/config.json` when `XDG_CONFIG_HOME` specifies an absolute directory.
Otherwise, use `$HOME/.config/three-finger-drag/config.json`.
Do not create a configuration file merely because the utility starts.

Use the Boolean setting `enabled_on_startup`, with an implicit default of `true`.
An explicit `false` requests disabled startup. Apply strict JSON validation to this setting.
Runtime enable and disable commands change only an in-memory override.
On restart, discard the override and read the configured startup state again.

Reject unknown fields and unsupported schema versions with English errors.
The specification must define whether invalid configuration stops startup or selects an approved fallback.
Use `grace_period_ms` for the grace duration, with a default of 300 ms when absent.
Accept only finite JSON numbers greater than zero. Reject zero, negative, nonnumeric, and nonfinite values.
Do not write the implicit default to disk.

Use `drag_start_threshold_mm` for the approved drag-start setting.
If the setting is absent, use 1 mm without writing an implicit default to disk.
Accept only a finite JSON number greater than zero. Reject strings, booleans, null, zero, negative values, and nonfinite values.
Do not convert numeric strings into numbers.
Pass the validated threshold to gesture logic as a policy value, without exposing configuration access to that package.
Use this same threshold for continuation. Do not introduce a separate continuation setting.

Use `ending_tap_max_duration_ms` for the ending-tap duration limit, with a default of 200 ms.
Use `ending_tap_max_displacement_mm` for the ending-tap displacement limit, with a default of 1 mm.
Resolve each absent setting independently. Do not write implicit defaults to disk.
Apply the same finite, positive JSON-number validation to each setting.
Keep these limits independent of `drag_start_threshold_mm`.
Supply validated limits to the gesture package through policy values.

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

For normal startup, acquire exclusive instance ownership for the desktop session before opening or grabbing touchpads.
If another instance owns the session, report an English error and exit without changing its resources.
Use atomic ownership acquisition so simultaneous starts cannot both succeed.
Control commands connect to the owner instead of acquiring ownership or starting another gesture-processing instance.

1. Read explicit configuration and validate approved settings.
2. Check the supported session, selected device, required access, and output capabilities.
3. Prepare virtual devices and verify desktop recognition before routing input.
4. Acquire exclusive input ownership through the verified backend procedure.
5. Read requested enablement from startup configuration. Clear any runtime override and initialize recognition as unarmed.
6. If any step fails, release acquired resources and report the operation, cause, and available guidance.

Use the readiness procedure after successful startup and each relevant availability change.
Disabled startup must preserve normal desktop gestures without reserving three-finger input after cleanup.

### Enablement, recovery, and rearming

Coverage: RF-2, RF-7, RF-8, and RF-9.

1. On an enable command, set the in-memory override to enabled and evaluate readiness.
2. On a disable command, set the override to disabled and clear armed recognition.
3. If disablement interrupts dragging or grace, terminate the gesture, cancel timers, and discard queued movement.
4. Release its held button immediately. If release fails, retain T05's pending cleanup and affected-touchpad suppression.
5. While cleanup remains pending, block new drags and discard affected-touchpad input regardless of requested enablement.
6. After cleanup, restore ordinary pointing and scrolling when touchpad access permits, independently of readiness.
7. If requested enablement is disabled, allow normal desktop three-finger handling and remain disabled.
8. If required access or lifecycle availability is absent, remain unarmed and unavailable for new drags.
9. Once enabled with access, lifecycle availability, and completed cleanup, wait for zero contacts and released physical touchpad buttons.
10. When both conditions hold, arm recognition and enter ready state.
11. Permit subsequent fresh contacts to form a gesture without clearing arming merely because contacts appear.
12. On drag start, consume arming. On failure, clear arming without changing requested enablement.
13. At normal completion, rearm only if zero contacts, released physical buttons, and the other readiness prerequisites hold.

Runtime commands must target the running instance. Their transport belongs to the command-interface implementation.
An ordinary physical press clears arming. Preserve its normal button action and wait for the full rearming conditions afterward.
Failed gestures never resume. Automatic recovery enables recognition only for a fresh gesture after readiness checks pass.

### Physical buttons and explicit cancellation

Coverage: RF-3, RF-7, RF-8, and RF-9.

1. Track physical-button state only for the selected touchpad.
2. If a release matches a consumed terminating press, consume it and clear that pairing record.
3. If a physical press occurs during dragging or grace, record it for consumption and terminate the gesture.
4. Stop gesture movement, discard queued movement, cancel timers, and clear armed recognition.
5. Release only the utility's held output button. Apply T05 cleanup if release fails.
6. Retain consumed-press records until their matching releases, even after ordinary input restoration or disablement.
7. For physical actions outside a drag, preserve ordinary behavior unless pending cleanup requires suppression.
8. Block recognition until all selected-touchpad physical buttons release and zero valid contacts satisfy T06's remaining readiness checks.
9. On an explicit cancel command during dragging or grace, perform the same termination and cleanup without consuming unrelated physical events.
10. Preserve requested enablement after cancellation. Do not resume the terminated gesture.

Events from other pointers stay outside this procedure. Their motion can move the desktop pointer during the utility's held action.
Do not claim that releasing the held action undoes an application operation.

### Lifecycle and desktop interruptions

Coverage: RF-4, RF-8, and RF-9.

1. Translate platform lock, suspend preparation, resume, unlock, session-loss, and normal-exit notifications into runtime lifecycle events.
2. On lock or suspend preparation, clear arming and mark lifecycle availability false.
3. Terminate any active gesture, discard queued movement, cancel timers, and release the held button.
4. Route unresolved release through T05 cleanup. Preserve the user's requested enablement state.
5. On unlock or resume, recheck session identity, access, lock state, and pending cleanup before evaluating fresh-contact readiness.
6. On session end or connection loss, block new drags and perform session-bound cleanup.
7. Reject stale movement and releases from the old session. Do not treat lost connection as confirmation of cleanup.
8. On normal exit, mark shutdown requested and block new drags before stopping gesture activity.
9. Release the held button and complete cleanup before restoring ordinary touchpad ownership and tearing down remaining resources.
10. If cleanup fails, retain the obligation and report the failure without claiming successful shutdown.
11. On display changes, preserve the gesture phase and existing grace deadline while lifecycle availability permits.
12. Use the desktop's resulting pointer position as the origin for subsequent relative output. Add no compensation or absolute warp.
13. On target disappearance, retain normal movement and ending rules without selecting or restoring an application target.

The Hyprland and Linux adapters own lifecycle notification access. Gesture logic receives plain lifecycle values through the runtime.
Normal exit restores ordinary touchpad handling before releasing session-instance ownership.
Do not persist lifecycle state or pending gesture state across process restarts.

### Event ordering

Coverage: RF-2 through RF-9.

1. Receive a complete contact frame, timer event, output result, or availability change.
2. Map its time into the runtime's monotonic clock domain.
3. Ignore stale timer generations and obsolete session events.
4. Order gesture events and deadlines by monotonic time. Process grace expiry before movement or tap lift at the same timestamp.
5. Pass domain values to the gesture logic.
6. Apply returned effects in order.
7. Record output submission results without assuming application receipt.
8. If an effect fails, enter the failure procedure.

### Start a drag

Coverage: RF-2, RF-3, RF-4, RF-8.

1. If release remains pending, discard affected-touchpad input without gesture recognition or ordinary forwarding.
2. Otherwise, require enabled, armed recognition with access available. Use RF-8 for ordinary routing while waiting.
3. Exclude detected palms and include thumbs when forming the valid contact set.
4. If the valid contact count or identities change, discard the previous start baseline.
5. If the count differs from three, preserve permitted ordinary input and wait for exactly three valid contacts.
6. If no start baseline exists, store the three contact identities and their average position in millimeters.
7. On subsequent frames, calculate the current average position for the same three contacts.
8. Calculate the straight-line distance from the fixed start baseline. Do not sum frame-to-frame distances.
9. If displacement is below the configured threshold, generate no drag.
10. If displacement reaches or exceeds the threshold, claim the gesture through the verified routing procedure.
11. Submit a logical primary-button hold through the desktop mapping at the current pointer position.
12. If submission fails, enter the failure procedure.
13. Consume armed recognition, establish the motion baseline at the crossing event, and enter dragging.
14. Discard all movement through that event. Emit movement only for subsequent group changes.

### Move and pause

Coverage: RF-3, RF-4, RF-5.

1. Outside grace, check valid contact count and identities before calculating motion.
2. If the count departs from three, stop gesture-driven pointer movement and enter grace with a new timer generation.
3. Set the grace deadline to interruption time plus the configured duration.
4. If identities change while the count stays three, reset the motion baseline to the new group position.
5. For that identity-change event, emit no motion, button transition, or grace timer.
6. Otherwise, calculate subsequent group displacement from the motion baseline.
7. Map that displacement and speed to the ordinary one-finger pointing response through the verified motion path.
8. Submit relative motion without another button press. Update the motion baseline.
9. When contacts stop moving, emit no further gesture-driven motion. Keep the hold without a stationary-contact timeout.
10. During grace, preserve the hold and deadline through later contact changes and stationary returns.

File managers, text widgets, and draggable title bars receive the same ordinary drag action.
The utility does not move files directly, edit selections, or command arbitrary window movement.

### Continue during the grace period

Coverage: RF-5, RF-6.

1. Process grace expiry before movement at or after its deadline.
2. Before expiry, discard the continuation baseline if valid contact count or identities change.
3. If exactly three valid contacts remain without a baseline, store their identities and average position in millimeters.
4. On subsequent frames, measure straight-line displacement from that fixed baseline. Do not accumulate traveled distance.
5. Below the shared drag-start threshold, preserve the pause and deadline.
6. At or above the threshold strictly before expiry, cancel the current timer generation.
7. Enter dragging without releasing or pressing the button again.
8. Establish the motion baseline at the threshold-crossing event's group position.
9. Discard all movement through that event, including repositioning displacement and threshold overshoot.
10. Emit motion only for subsequent group changes, using the ordinary-pointing mapping.

A later departure from three contacts creates a new grace deadline under the move-and-pause procedure.

### Recognize an ending tap

Coverage: RF-7.

1. During the grace period, track whether all valid contacts have lifted.
2. After zero valid contacts, start a candidate when exactly one valid contact touches down before expiry.
3. Store its identity, initial position, and monotonic touchdown time. Initialize maximum displacement to zero.
4. Preserve the grace deadline and defer recognition until finger lift.
5. Ignore detected palms. Count a thumb as a valid contact.
6. Update maximum displacement using Euclidean distance from the initial position in millimeters.
7. Cancel the candidate if displacement or elapsed duration exceeds its configured limit.
8. Cancel the candidate if a second valid contact appears.
9. After cancellation, wait for zero valid contacts before accepting a fresh candidate.
10. If grace expiry precedes or coincides with candidate lift, discard the candidate and end the drag through expiry.
11. If the candidate lifts before grace expiry, check duration and maximum displacement against their inclusive limits.
12. If both limits pass, emit one ending-tap result for consumption and drag release.

Expiry takes priority when candidate lift and grace expiry share a timestamp. Submit only one drag release.
Candidate cancellation alone generates no drag-release effect.
Clear candidate state when the grace period ends through continuation, expiry, recognized tap, or failure.

### End a drag

Coverage: RF-7, RF-8.

1. Detect an approved ending tap or a valid grace-period expiry.
2. Consume the ending tap without forwarding a click.
3. Cancel the current timer generation.
4. Submit one button release for the held drag action.
5. If submission fails, enter the failure procedure.
6. Clear gesture baselines and rearm only after zero contacts, released physical touchpad buttons, and the other readiness checks.
7. Restore ordinary input through the verified routing procedure.

### Handle failures

Coverage: RF-9, with RF-8 interaction.

1. Mark the current gesture terminated and reject its future movement effects.
2. Cancel its timers, clear its tap candidate, and discard queued movement.
3. Track any held or uncertain button state as an in-memory cleanup obligation with its owning device and session identities.
4. While release remains pending, suppress affected-touchpad input. Discard it without buffering for later replay.
5. Leave other devices outside the suppression rule.
6. If release is possible in the owning session, submit it immediately.
7. Complete cleanup only after release or desktop confirmation that the owning virtual device or session no longer exists.
8. If loss remains unconfirmed, retain the obligation and suppression. Do not infer success from connection loss.
9. When another release opportunity occurs, verify the owning session before retrying. Never redirect cleanup into a new session.
10. Report the operation, cause, pending-release status, and available guidance in English, without input-event details.
11. Suppress repeated copies of the same unchanged error during cleanup retries.
12. After cleanup, restore ordinary input when touchpad access is available, independently of new-drag readiness.
13. Preserve requested enablement and clear arming. Automatically evaluate readiness after recovery, without resuming the terminated gesture.

Retain only cleanup and error-deduplication metadata. Do not persist it or retain suppressed input events.
The adapter supplies cleanup evidence to the runtime. Gesture logic must not query the desktop.
No process can execute cleanup after an uncatchable termination.
Kernel device teardown and compositor behavior therefore require direct failure tests.
T08 requires these tests to observe cleared button state and normal touchpad handling after process death.
If those outcomes fail, the backend does not satisfy acceptance. Do not mark the crash scenario passed based on process exit alone.

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
| Ordinary-touchpad motion mapping | Matches the user's existing pointing direction, sensitivity, and acceleration. | Independent drag sensitivity introduces a different response. | RF-3, RF-6 |
| Logical primary-button effects | Preserves selection and dragging under left-handed settings. | A fixed physical left-button mapping can trigger a secondary action. | RF-2, RF-4 |
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
| RF-2 | Start at the configured group-displacement threshold. Verify palms, thumbs, staggered placement, identity changes, and stationary contacts. |
| RF-3 | Relative movement maintains one held button and does not issue repeated presses. |
| RF-4 | Start effects hold the button before drag motion and preserve the current pointer origin. |
| RF-5 | Zero, one, two, and additional contacts pause movement and apply the approved grace rules. |
| RF-6 | Continuation cancels stale deadlines, preserves the hold, and rejects repositioning displacement. |
| RF-7 | Verify fresh-contact eligibility, inclusive tap limits, cancellation, configuration, and recognition on lift. Expiry and ending taps release once. |
| RF-8 | Ordinary input routes only under the approved ownership rules. Competing three-finger output remains suppressed. |
| RF-9 | Failures terminate gestures. Cleanup requires release or confirmed owner disappearance. Reject cross-session releases and deduplicate repeated errors. |

Test before, at, and after every approved timing boundary.
Test repeated contact changes, stationary contacts, delayed events, and contacts remaining after release.
Use generated synthetic sequences to check balanced button ownership and stale-timer rejection.
Keep generated events in memory.

### T01 acceptance scenarios

Coverage: RF-2 and RF-3. These are authored future tests, not executed implementation tests.

| Scenario | Expected result |
| --- | --- |
| No configured threshold | Use 1 mm without creating configuration. |
| Three-contact group displaces 0.99 mm from its baseline | No drag starts. |
| The same group reaches 1 mm | Start one drag at the current pointer position. |
| Threshold is 2 mm | No start at 1.99 mm. Start at 2 mm. |
| One contact moves 3 mm while two remain stationary | The average position moves 1 mm and starts a drag with the default threshold. |
| Three contacts remain stationary | Generate neither a click nor a drag. |
| Contacts arrive separately with arbitrary delays | Establish the baseline only when the third valid contact arrives. |
| Contacts move before the third contact arrives | Exclude earlier movement from start displacement. |
| The group repeatedly moves out and back by 0.6 mm | No drag starts, even when total traveled distance exceeds 1 mm. |
| Contact count changes or one identity changes before start | Reset the baseline. Count only displacement after a new three-contact baseline. |
| A palm joins the same three valid contacts | Ignore the palm without changing the baseline. |
| The group includes a thumb and two other fingers | Apply the same three-contact displacement rule. |
| The configuration supplies zero, a negative number, a string, null, a boolean, or a nonfinite value | Reject the threshold. |

Also test equal displacement along each axis and diagonally to verify direction-independent millimeter measurements.

### T02 acceptance scenarios

Coverage: RF-7. These scenarios define future tests, not executed implementation tests.
Unless stated otherwise, use default limits and a grace deadline later than the candidate's lift.

| Scenario | Expected result |
| --- | --- |
| All valid contacts lift, then one finger touches and lifts within both limits | Recognize on lift. Release the drag once without an additional click. |
| A valid candidate touches down but remains down | Do not recognize an ending tap at touchdown. Preserve the grace deadline. |
| One drag finger remains and lifts without a fresh touchdown | No ending tap occurs. |
| The fresh contact is a thumb | Apply the same ending-tap rules. |
| A detected palm remains while all fingers lift | The valid contact count reaches zero. A later fresh finger can become a candidate. |
| Candidate durations are 199 ms, 200 ms, and 201 ms | Accept the first two durations. Cancel the third candidate. |
| Maximum displacements are 0.99 mm, 1 mm, and 1.01 mm | Accept the first two distances. Cancel the third candidate. |
| Displacement exceeds 1 mm and returns to zero | The candidate remains canceled. |
| Motion travels more than 1 mm but always stays within 1 mm of the initial position | Accept movement. Total traveled distance is not the displacement limit. |
| A second valid contact appears, then lifts | Cancel the candidate. The remaining contact cannot restore it. |
| All valid contacts lift after cancellation, then one touches down before grace expiry | Allow a fresh candidate under the same limits. |
| A detected palm appears beside a candidate | Keep the candidate valid if its other conditions still hold. |
| Grace expiry occurs before candidate lift | End through expiry. The later lift cannot end the same drag again. |
| Candidate lift coincides exactly with grace expiry | Expiry wins. Release once without recognizing an ending tap. |
| Duration is configured to 300 ms and displacement to 2 mm | Accept 300 ms and 2 mm. Reject values exceeding either configured limit. |
| Only one ending-tap setting is present | Use that value and the independent default for the other setting. |
| The drag-start threshold changes | Ending-tap limits remain unchanged. |
| Either tap setting is zero, negative, nonnumeric, or nonfinite | Reject the value. Configuration recovery remains pending T11. |

Use synthetic coordinates and injected monotonic time for boundary checks.
Verify that canceled candidates cannot emit ending-tap results when their original contact later lifts.

### T03 acceptance scenarios

Coverage: RF-5, RF-6, and RF-7. These scenarios define future deterministic tests.
Use a 300 ms grace duration and a 1 mm shared movement threshold unless stated otherwise.
Times below are relative to the first interruption.

| Scenario | Expected result |
| --- | --- |
| No grace setting is present | Use 300 ms without writing configuration. |
| Grace duration is configured to 500 ms | Set the deadline 500 ms after the interruption. |
| Grace duration is zero, negative, nonnumeric, or nonfinite | Reject the value. Configuration recovery remains pending T11. |
| Dragging changes from three contacts to zero, one, two, or four | Pause movement and set one deadline at 300 ms. |
| Two remaining contacts lift at 100 ms | Preserve the original 300 ms deadline. |
| Contacts repeatedly change during grace | Reset affected continuation baselines without extending the deadline. |
| Three contacts return at 100 ms and remain stationary | Remain paused and release at 300 ms. |
| Three contacts return at a different touchpad position | Establish a fresh baseline without moving the pointer. |
| Group displacement reaches 0.99 mm at 299 ms | Remain paused. Release at 300 ms unless another qualifying event occurs earlier. |
| Group displacement reaches 1 mm at 299 ms | Continue the same drag and cancel the timer. |
| Group displacement reaches 1 mm at 300 ms or later | Expiry wins. Release once without continuing the expired drag. |
| A valid ending tap lifts at 299 ms | End through the tap and cancel the timer. Generate no additional click. |
| A valid candidate lifts at exactly 300 ms | End through expiry once, without recognizing an ending tap. |
| Three contacts return, move 0.6 mm, then one contact changes identity | Discard displacement. Establish a fresh baseline while preserving the deadline. |
| Displacement repeatedly returns to zero without reaching 1 mm | Do not continue, regardless of total traveled distance. |
| The shared threshold is configured to 2 mm | Require 2 mm for both drag start and continuation. |
| Continuation succeeds at 200 ms, then another interruption occurs at 250 ms | Set the new deadline to 550 ms. |
| The canceled first timer fires at 300 ms after that continuation | Ignore its stale generation. Preserve the new deadline. |

Also test a detected palm beside the returning group. It must change neither baseline nor deadline.
Use injected monotonic time to verify exact-deadline precedence independently of callback ordering.

### T04 acceptance scenarios

Coverage: RF-2, RF-3, RF-4, and RF-6. These scenarios define future tests, not executed implementation tests.

| Scenario | Expected result |
| --- | --- |
| Start displacement crosses a 1 mm threshold at 1.4 mm | Hold primary at the existing pointer position. Emit no motion for the crossing event. |
| A subsequent event moves the same group by 0.2 mm | Map only that 0.2 mm movement through the ordinary-pointing response. |
| Continuation qualifies before expiry with threshold overshoot | Cancel grace and keep the hold. Discard all motion through the crossing event. |
| Fingers return at a distant touchpad position | Repositioning causes no pointer movement. Only movement after qualifying continuation moves the pointer. |
| Equivalent group and one-finger trajectories use equal displacement and speed | Direction matches. Distance error stays within 5% of the reference or 2 logical pixels, whichever is larger. |
| Group movement is rightward or upward | Pointer movement is rightward or upward, respectively. |
| Natural scrolling is enabled or disabled | The same group trajectory produces the same pointer direction. |
| Left-handed mode swaps physical button mappings | Start produces ordinary primary-button selection or dragging, not a secondary action. |
| A contact identity changes while three valid contacts remain | Keep the hold and reset the motion baseline. Emit no replacement movement or grace timer. |
| The new group then moves less than the start threshold | Move normally. Contact replacement during dragging requires no new recognition threshold. |
| Three active-drag contacts stop for longer than the grace duration | Stop pointer motion and retain the hold. No grace timer starts. |
| Three contacts return during grace and remain stationary | Preserve T03's deadline and release at expiry. |

Use deterministic effect tests for threshold, baseline, and stationary-contact rules.
Use desktop comparisons for sensitivity, acceleration, and button mapping. A successful output write alone cannot prove these behaviors.
Document comparison trajectories and aggregate results without saving input recordings.

### T05 acceptance scenarios

Coverage: RF-8 and RF-9. These scenarios define future tests, not executed implementation tests.

| Scenario | Expected result |
| --- | --- |
| Failure occurs during dragging or grace | Terminate the gesture immediately. Stop movement, cancel timers, and discard queued movement. |
| Release remains possible after failure | Release immediately in the owning session. Complete cleanup after release. |
| Release is temporarily unavailable | Keep a separate pending release and suppress affected-touchpad input. |
| Contacts move, scroll, or tap during cleanup | Discard their input. Do not replay it after cleanup. |
| Another input device produces events during cleanup | This touchpad suppression rule does not capture or suppress those events. |
| Contacts satisfy continuation after failure | Do not resume the terminated gesture. |
| A canceled gesture timer arrives after failure | Ignore it without producing movement or another gesture action. |
| The control connection disappears without owner confirmation | Keep cleanup pending. Do not infer button release. |
| The desktop confirms destruction of the owning virtual device | Complete cleanup without requiring release into another device. |
| The desktop confirms the owning session no longer exists | Complete cleanup without sending an old release into the new session. |
| A new session appears without confirmation of the old owner's disappearance | Retain pending cleanup. Never send the old release to the new session. |
| Release becomes possible again in the owning session | Attempt release at the earliest opportunity. Complete cleanup after release. |
| Cleanup completes while touchpad access is available | Restore ordinary pointing and scrolling independently of three-finger drag availability. |
| Cleanup completes before touchpad access returns | Restore ordinary input when access becomes available. |
| Cleanup retries encounter the same unchanged failure | Report operation, cause, pending status, and available guidance once rather than on every retry. |

Use deterministic runtime tests for routing, canceled effects, and error deduplication.
Use adapter integration tests to verify release evidence and confirmation of device or session disappearance.
Check that successful submission alone cannot clear an obligation without the verified release contract.

### T06 acceptance scenarios

Coverage: RF-2, RF-7, RF-8, and RF-9. These scenarios define future tests, not executed implementation tests.

| Scenario | Expected result |
| --- | --- |
| Startup setting is absent | Request enabled operation without writing implicit configuration. |
| Startup setting is false | Keep dragging disabled and allow normal desktop gesture behavior after cleanup. |
| Startup or enablement occurs with contacts down | Wait for available access, completed cleanup, and zero valid contacts before arming. |
| Readiness prerequisites hold and the contact count is zero | Arm recognition. Subsequent fresh contacts can start a drag under RF-2. |
| Contacts appear after arming | Keep recognition armed until drag start or loss of readiness. Do not require zero contacts continuously. |
| Grace expires with one, two, or three contacts remaining | End the drag and wait for all valid contacts to lift before another drag. |
| A drag ends with zero contacts | Treat the zero-contact requirement as satisfied. Rearm when the remaining checks pass. |
| A failed gesture loses access, then access and cleanup recover | Automatically wait for zero contacts and arm a new gesture if still enabled. |
| Failure recovery occurs while explicitly disabled | Keep dragging disabled. Restore ordinary input when access and cleanup permit. |
| Disablement occurs during dragging or grace | Immediately terminate, stop motion, cancel timers, and release without waiting for expiry. |
| Disablement cannot release the held button | Preserve T05 cleanup and suppression. Do not restore desktop three-finger handling prematurely. |
| Enablement occurs during pending cleanup | Record enabled intent but block recognition and affected-touchpad input until cleanup completes. |
| Cleanup completes while enabled contacts remain down | Restore ordinary pointing and scrolling while waiting for zero contacts to arm dragging. |
| Disablement completes with no pending cleanup | Stop reserving three-finger gestures. Allow normal desktop gestures. |
| Runtime disable overrides configured enabled startup | Remain disabled until another command or restart. Restart uses configuration. |
| Runtime enable overrides configured disabled startup | Enable subject to readiness checks. Restart returns to configured disabled startup. |
| Runtime enable or disable command executes | Leave startup configuration unchanged. |

Use deterministic state tests to distinguish requested enablement, readiness, and ordinary-input routing.
Use integration tests to verify commands reach the running instance and disabled operation restores normal desktop gestures.

### T07 acceptance scenarios

Coverage: RF-3, RF-7, RF-8, and RF-9. These scenarios define future tests, not executed implementation tests.
For T06 scenarios, arming now also requires every selected-touchpad physical button released.

| Scenario | Expected result |
| --- | --- |
| A physical touchpad button is pressed during dragging | Terminate, cancel timers, and release the utility's hold. Consume the physical press. |
| A physical touchpad button is pressed during grace | Terminate without waiting for expiry. Consume the press and release the utility's hold. |
| The consumed button later releases after cleanup or disablement | Consume the matching release. Generate no extra click or unmatched release. |
| A physical button is already held when contacts qualify | Preserve its ordinary action. Do not start a three-finger drag. |
| All physical buttons release while contacts remain | Keep recognition unarmed until zero valid contacts are observed. |
| Contacts lift while a physical button remains held | Keep recognition unarmed until all physical buttons release with zero contacts. |
| Another mouse moves or sends button events during dragging or grace | Let the desktop process them normally without interception or gesture termination. |
| Contacts appear on an unselected touchpad | Preserve its ordinary behavior. Do not use those contacts for recognition. |
| Two selected-touchpad contacts and one other-touchpad contact move | Do not combine them into a three-finger gesture. |
| A second instance starts in the same desktop session | Reject it before device acquisition. Preserve the existing instance and any active gesture. |
| Two instances attempt startup simultaneously | Exactly one acquires ownership. The other exits with an error. |
| Enable, disable, or cancel commands target a running session | Deliver the command to its existing instance without starting another gesture processor. |
| Cancellation occurs during dragging or grace | Stop movement, cancel timers, release, and require fresh contacts. Preserve enabled state. |
| Release fails after physical termination or cancellation | Keep T05's pending cleanup and suppression. Do not resume the gesture. |

Use deterministic tests for physical-button pairing, recognition blocking, and cancellation effects.
Use integration tests for duplicate-instance ownership, command delivery, and desktop behavior with another pointing device.
Observe application outcomes without asserting that cancellation reverses a drop.

### T08 acceptance scenarios

Coverage: RF-4, RF-8, and RF-9. These scenarios define future tests, not executed implementation tests.

| Scenario | Expected result |
| --- | --- |
| Session locks during dragging or grace | Terminate, discard queued movement, cancel timers, release, and block new drags. |
| Suspend preparation occurs during dragging or grace | Apply the same termination and release behavior before suspension where the environment permits. |
| Release fails during lock or suspend preparation | Retain T05 cleanup and affected-touchpad suppression. |
| Unlock or resume occurs with cleanup incomplete | Keep new drags blocked until cleanup and all other readiness checks pass. |
| Unlock or resume occurs with contacts or physical buttons held | Require zero valid contacts and all selected-touchpad physical buttons released before arming. |
| Unlock or resume occurs while explicitly disabled | Keep dragging disabled. Preserve T06 ordinary-input behavior after cleanup. |
| Normal exit occurs during dragging or grace | Release, complete cleanup, and restore ordinary touchpad handling before exit. |
| Normal exit cannot complete release | Retain cleanup and report failure without claiming successful shutdown. |
| The process crashes while holding its drag button | Observe system removal of the held button state and restoration of ordinary touchpad handling. |
| The process receives forced termination while holding its drag button | Verify the same system outcomes without relying on any process cleanup handler. |
| A later restart begins with contacts already down | Use startup configuration and wait for fresh-contact readiness. Do not restore the old gesture or runtime override. |
| Session end or connection loss occurs during a drag | Terminate and block new drags. Apply T05 without sending stale input to another session. |
| A session becomes available but remains locked | Keep recognition unavailable. |
| A monitor disconnects or the display layout changes during dragging | Preserve the gesture while the session remains available. Add no compensating pointer movement. |
| The desktop repositions the pointer or dragged window | Continue subsequent relative motion from the resulting pointer position. |
| Display layout changes during grace | Preserve the existing deadline. Do not resume merely because the pointer moves externally. |
| The dragged window closes or its application exits | Keep normal gesture movement and ending rules. Do not select a replacement target or promise undo. |

Use synthetic lifecycle events for deterministic transition tests.
Run actual crash and forced-termination tests from an external observer in the isolated desktop test session.
Verify both button state and ordinary pointing and scrolling after process death, independently of any later utility restart.
Record aggregate results and scenario outcomes without captured input-event streams.

### T09 baseline verification

Coverage: RF-1, RF-4, and RF-8. T13 records inventories before these acceptance checks run.

| Baseline item | Required evidence |
| --- | --- |
| Linux environments | Fully updated stable Arch Linux and CachyOS, with exact kernel, Hyprland, libinput, XWayland, and application versions. |
| Linux hardware | Device identity, three independent contact positions and identities, fourth-contact detection, axis resolution, and palm-rejection information. |
| Native Wayland applications | Thunar file dragging and Mousepad selection and title-bar scenarios, with verified backend identity. |
| XWayland applications | The same scenarios inside Hyprland, with verified backend identity. |
| Window states | Floating title-bar movement and comparison with ordinary primary-button behavior in tiled, maximized, and fullscreen states. |
| macOS environment | One compatible Mac's model, trackpad, macOS build, and relevant settings, with three-finger dragging enabled. |
| macOS applications | Finder file dragging and TextEdit selection and title-bar scenarios. |
| Run consistency | Fixed Linux versions per acceptance run and fixed Mac hardware, software, and settings throughout comparison. |

Each Linux application/backend combination must run on both distributions.
Verify the actual backend rather than relying solely on launch arguments or environment variables.
If an application cannot run under a required backend, record the blocker rather than claiming that matrix entry passed.
Use ordinary primary-button title-bar dragging as the baseline for each tested window state.
No utility-driven window-state transition is required to overcome desktop movement restrictions.
Record inventories and aggregate outcomes without persisting input events.

### T10 measurement procedure and acceptance scenarios

Coverage: RF-2, RF-3, RF-6, RF-7, RF-9, NFR-1, and NFR-2.
These procedures define future verification. No measurements have been collected for this task.

1. Use T13's recorded environment and verify the application's actual native Wayland or XWayland backend.
2. Warm up the session and record power mode, display refresh rate, settings, measurement precision, and uncertainty.
3. Select one operation and its origin and endpoints from NFR-2's timing table.
4. Map input timestamps, runtime timestamps, and deadlines to the same monotonic clock domain for processing measurements.
5. Use an independent observer to measure actual desktop button or pointer changes against the trigger.
6. Calibrate any cross-clock timing alignment and include its uncertainty in the result.
7. Collect at least 100 valid samples for that operation and environment. Retain slow samples in the calculation.
8. Calculate processing and desktop distributions separately. Use the nearest-rank p95 and the maximum observed sample.
9. Compare both statistics with their limits, including uncertainty. Report pass, fail, or inconclusive.
10. Save only aggregates, sample counts, scenario descriptions, and measurement metadata. Discard temporary in-memory samples afterward.
11. Repeat each check separately on both distributions and both application backends.

For 100 samples, nearest-rank p95 is the 95th sample after sorting by duration.
A missing response is a failed observation, not a sample to omit or a successful output submission.
Do not substitute the time the utility notices restored output for the independently established release opportunity.
Record unavailable-output waiting separately from measured cleanup response. It does not satisfy cleanup by itself.
Do not time nonexistent pointer output at the continuation threshold. T04 requires the pointer to remain still there.
Owner-disappearance confirmation remains a separate T05 cleanup test, not a synthetic button-release latency sample.

For pointer comparison, use matched displacement and speed with the same Linux touchpad, display scale, and settings.
Compare active-motion distances after recognition. Do not count the discarded start or continuation threshold movement as pointer travel.
Use logical desktop pixels consistently. Convert physical display measurements using the recorded display scale when necessary.
Calculate the allowed distance error as the larger of 0.05 times the one-finger reference distance and 2 logical pixels.
Check direction separately and preserve the no-repositioning-jump requirement.

| Scenario | Expected result |
| --- | --- |
| Processing p95 is 5 ms and maximum is 20 ms | Pass only if uncertainty does not permit either limit to be exceeded. |
| Processing p95 exceeds 5 ms or any sample exceeds 20 ms | Fail that timing check. |
| Desktop p95 exceeds 50 ms or any sample exceeds 100 ms | Fail that timing check. |
| A slow sample occurs in a warmed-up run | Include it in the percentile and maximum. |
| An estimate could fall on either side of a limit after uncertainty | Report inconclusive rather than passed. |
| A 300 ms grace interval ends | Measure release response from the fixed deadline, not initial finger lift. |
| Output is unavailable before recovery | Measure cleanup latency from when release becomes possible, excluding the unavailable interval. |
| Continuation qualifies | Measure its state transition internally. Measure pointer response from the next movement event. |
| The one-finger reference travels 100 logical pixels | Allow at most 5 logical pixels of distance difference, with matching direction. |
| The one-finger reference travels 20 logical pixels | Allow at most 2 logical pixels of distance difference, with matching direction. |
| A macOS outcome differs for start, repositioning, continuation, tap-to-end, or release | Record and review the mismatch before claiming equivalence. |

Use Finder and TextEdit for the fixed macOS comparison. Report timing differences separately from outcome matching.
Linux pointer mapping remains relative to ordinary one-finger Linux pointing rather than identical screen travel on different hardware.
Persist neither raw input events nor pointer paths as measurement evidence.

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
Use touchpads that meet RF-1's capabilities and the exact inventories recorded under T13.
Run Thunar file dragging and Mousepad text-selection and title-bar checks under native Wayland and XWayland on both distributions.
Verify application backends and test the approved window states against ordinary primary-button behavior.
Freeze exact software versions for each run. Record a changed environment as a new run rather than mixing its evidence.

| Requirement | Desktop evidence |
| --- | --- |
| RF-1 | Recorded environment versions and results for both distributions. |
| RF-2 | Three-finger movement starts at the current pointer position without a physical press. |
| RF-3 | Pointer motion follows approved direction, scale, and response limits. |
| RF-4 | File drops, text selections, and title-bar window movement complete in representative applications. |
| RF-5 | Every specified contact-count change pauses movement while retaining the drag. |
| RF-6 | Repositioning continues the same drag without a pointer jump. |
| RF-7 | Expiry and taps end the drag without an additional click. |
| RF-8 | Ordinary input works outside active drags and pending cleanup. Cleanup suppresses only the affected touchpad without replay. |
| RF-9 | Disconnection, permission loss, and desktop loss terminate the gesture and satisfy the approved release policy. |

Compare the same authored interaction scripts with the selected macOS reference.
Use Finder and TextEdit with native three-finger dragging enabled.
Keep the recorded Mac model, trackpad, macOS build, and relevant settings fixed throughout the comparison.
Store scenario descriptions, aggregate timings, and pass/fail results only.
Do not capture or persist input-event streams.
Apply T10's approved measurement procedure and limits. Inconclusive measurements do not establish acceptance.
Compare start, repositioning, continuation, tap-to-end, and release outcomes with native macOS and review every mismatch.
Report timing differences separately without relaxing the Linux latency limits.

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

- [Apple three-finger drag enablement](https://support.apple.com/en-us/102341)

Coverage: RF-1, RF-5, RF-7, RF-8, RF-9.

- [libinput three-finger drag behavior](https://wayland.freedesktop.org/libinput/doc/latest/drag-3fg.html)
- [libinput gesture and hold-event limitations](https://wayland.freedesktop.org/libinput/doc/latest/gestures.html)
- [Linux uinput and libevdev guidance](https://docs.kernel.org/input/uinput.html)

These references establish API capabilities and limitations. They do not prove this backend works with the selected Hyprland release.
