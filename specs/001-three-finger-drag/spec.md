# Specification: Three-Finger Drag

## Context and objective

Touchpad users need to drag files, select text, and move windows without a physical press.
The objective is to match the macOS three-finger drag experience.
The first release supports Arch Linux and CachyOS with Hyprland on Wayland.

A brief grace period lets users lift or reposition their fingers without releasing a drag.
The team must verify the specified timing limits and compare observable behavior with macOS before claiming an exact match.

## Users

- Arch Linux and CachyOS users who use a touchpad with Hyprland on Wayland.
- Users who want to drag without sustained physical pressure on the touchpad.

## User stories

- As a user, I want to drag files with three fingers without holding a physical button.
- As a user, I want to select text with three fingers without holding a physical button.
- As a user, I want to move windows by their draggable title bars without holding a physical button.
- As a user, I want to reposition my fingers so I can drag beyond one touchpad stroke.
- As a user, I want a tap to end the drag without waiting for the grace period.
- As a user, I want ordinary pointing and scrolling outside a drag.
- As a user, I want failures to end a drag so the desktop does not retain a held drag action.

## Functional requirements

### Terms

- **Enabled:** Three-finger dragging is requested by the startup configuration or the latest runtime enable command. Readiness remains a separate condition.
- **Armed:** The utility has observed zero valid contacts with all selected-touchpad physical buttons released after the applicable readiness checks.
- **Ready:** The feature is enabled and armed, required access is available, and neither an active drag nor pending release exists.
  Lifecycle availability, reliable contact state, and released selected-touchpad physical buttons are also required.
- **Lifecycle available:** The desktop session is available and unlocked. The utility is not exiting, preparing for suspend, or suspended.
- **Active drag:** A gesture in dragging or grace that has not terminated. Failure terminates it even if button release remains pending.
- **Terminated gesture:** A gesture that cannot produce further drag movement or resume. Its output cleanup may remain pending.
- **Pending release:** An unresolved cleanup obligation for the terminated gesture's held button, separate from gesture state.
- **Cleanup complete:** The button has been released, or the desktop confirms that its owning virtual device or session no longer exists.
- **Grace period:** A bounded interval that keeps the drag held and pauses gesture-driven pointer movement.
- **Ending tap:** A fresh one-finger tap that satisfies RF-7 during the grace period. Recognition occurs on finger lift.
- **Tap candidate:** A fresh single valid contact after all valid contacts have lifted during the drag's grace period.
- **Tap displacement:** The straight-line distance from the candidate's initial contact position, measured in touchpad millimeters.
- **Supported environment:** Arch Linux or CachyOS with Hyprland on Wayland and a compatible touchpad.
- **Drag failure:** Touchpad disconnection, missing or revoked permissions, desktop connection loss, input synchronization loss, or output-write failure.
- **Valid contact:** A finger contact, including a thumb, that is not a detected palm. Contact counts refer to valid contacts.
- **Group position:** The arithmetic average of the three valid contacts' positions on the touchpad.
- **Start baseline:** The group position when exactly three valid contacts form the current group, before a drag starts.
- **Group displacement:** The straight-line distance from the start baseline to the current group position, measured in touchpad millimeters.
- **Qualifying start movement:** Group displacement that reaches or exceeds the configured threshold while the same three valid contacts remain.
- **Grace deadline:** The time of the first contact-count departure from three during dragging, plus the configured grace duration.
- **Continuation baseline:** The group position when exactly three valid contacts form the current group during the grace period.
- **Qualifying continuation movement:** Displacement from the continuation baseline that reaches the drag-start threshold while the same three valid contacts remain.
- **Logical primary button:** The desktop button action for ordinary selection and dragging, including the desktop's left-handed mapping.
- **Motion baseline:** The group position used to measure subsequent relative pointer movement during dragging.
- **Selected touchpad:** The single touchpad that supplies contacts and physical-button events to this feature. Other touchpads retain normal desktop behavior.

Each requirement has acceptance criteria in EARS notation.
The words WHERE, WHILE, WHEN, IF, THEN, and SHALL identify conditions and required responses.

### RF-1: Supported environments

- WHERE the environment is supported, the utility SHALL provide the three-finger drag behavior in this specification.
- WHERE applications use native Wayland or XWayland within Hyprland, the utility SHALL support the specified drag operations.

#### Compatibility and acceptance baseline

Compatible touchpads must provide all these capabilities:

- Separate positions and stable identities for at least three simultaneous contacts.
- Detection of a fourth contact, so recognition can distinguish three contacts from additional contacts.
- Physical axis resolution sufficient to measure displacement in millimeters.
- Contact information sufficient for palm rejection while preserving the approved thumb behavior.

Touchpads that lack any required capability are deferred beyond the MVP.
These requirements define eligibility. Actual device compatibility still requires acceptance evidence.

Use fully updated stable Arch Linux and CachyOS installations when acceptance testing begins.
Record exact kernel, Hyprland, libinput, XWayland, and application versions before testing.
Keep those versions fixed throughout each acceptance run. Development builds are outside this baseline.

| Operation | Linux application | Required application backends |
| --- | --- | --- |
| File dragging | Thunar | Native Wayland and XWayland within Hyprland |
| Text selection | Mousepad | Native Wayland and XWayland within Hyprland |
| Draggable title-bar movement | Mousepad | Native Wayland and XWayland within Hyprland |

Run these checks on both distributions. Verify each application's actual display backend during testing.
Application availability alone does not establish coverage for both backends.

Use native macOS three-finger dragging on one compatible Mac as the comparison reference.
Record its model, trackpad, exact macOS build, and relevant settings. Keep them fixed throughout the comparison.
Enable three-finger dragging according to [Apple's instructions](https://support.apple.com/en-us/102341).
Use Finder for file dragging and TextEdit for text selection and draggable title-bar movement.
T13 records the actual Linux devices and Mac environment before acceptance testing.

#### Configuration and touchpad selection

Coverage: RF-1, RF-2, RF-5, RF-7, and RF-9.

| Setting | Default | Accepted values |
| --- | --- | --- |
| Configuration version | 1 | Integer 1 only |
| Enabled startup | Enabled | Boolean |
| Drag-start and continuation threshold | 1 mm | Finite positive number, subject to conversion bounds |
| Grace duration | 300 ms | Finite positive number, subject to conversion bounds |
| Ending-tap duration limit | 200 ms | Finite positive number, subject to conversion bounds |
| Ending-tap displacement limit | 1 mm | Finite positive number, subject to conversion bounds |
| Touchpad selection | Automatic when exactly one compatible touchpad exists | Explicit identification of one compatible touchpad |

- IF the configuration file is missing, THEN the utility SHALL use defaults without creating a file.
- WHEN a setting is omitted, the utility SHALL use that setting's default without writing it to disk.
- IF existing configuration is unreadable, malformed, or invalid, THEN the utility SHALL fail startup while preserving normal touchpad handling.
- WHEN reporting configuration failure, the utility SHALL identify the file and cause, and the setting when possible, in English.
- WHEN reporting configuration failure, the utility SHALL include available corrective guidance without silently substituting defaults for invalid values.
- IF configuration contains unknown or duplicate settings or an unsupported version, THEN the utility SHALL reject it.
- WHEN validating configuration, the utility SHALL require actual Boolean and numeric values without converting strings.
- IF a numeric value overflows or rounds to zero during conversion, THEN the utility SHALL reject it and report the supported range.
- WHEN validating numeric settings, the utility SHALL reject invalid values rather than silently clamp them.
- WHEN configuration changes during execution, the utility SHALL retain its loaded settings until restart.
- WHEN the utility restarts, the utility SHALL read and validate configuration before taking control of the touchpad.
- WHILE running, the utility SHALL continue to apply enable, disable, and cancel commands under their existing rules.
- WHERE an explicit touchpad selection exists, the utility SHALL select that compatible device without substituting another.
- IF the selected device is missing or incompatible, THEN the utility SHALL fail startup and report the selection and reason.
- WHERE selection is omitted and exactly one compatible touchpad exists, the utility SHALL select it automatically.
- IF selection is omitted and multiple compatible touchpads exist, THEN the utility SHALL fail startup and list the choices.
- IF no compatible touchpad exists, THEN the utility SHALL fail startup while preserving normal touchpad handling.

All selection failures preserve normal touchpad handling. Configuration changes during a drag cannot change its thresholds or deadlines.
Conversion bounds refine the positive finite-number rules throughout this specification.
The implementation must document its supported numeric ranges and report them for conversion failures.
Runtime device loss follows RF-9 and the same-device recovery rules below.

- WHEN the selected touchpad disconnects, the utility SHALL retain that device selection and apply RF-9 cleanup.
- WHILE the selected touchpad is absent, the utility SHALL wait for that same device without selecting a substitute.
- WHEN the same compatible device returns, the utility SHALL recover only after cleanup and readiness checks pass.
- WHEN changing the selected touchpad, the user SHALL change configuration and restart the utility.

### RF-2: Start a drag

- WHILE ready, WHEN qualifying start movement occurs, the utility SHALL start a drag at the current pointer position.
- WHEN a three-finger drag starts, the utility SHALL require no physical touchpad press.
- WHEN a drag starts, the utility SHALL hold the desktop's logical primary button, respecting left-handed mode.
- WHEN start movement reaches the threshold, the utility SHALL discard movement through the threshold-crossing event without moving the pointer.
- WHEN a drag starts, the utility SHALL establish the motion baseline at the current group position.
- WHILE ready, WHEN three fingers remain stationary, the utility SHALL generate neither a click nor a drag.
- WHILE ready, WHEN the contact count differs from three, the utility SHALL refrain from starting a three-finger drag.
- WHEN counting contacts, the utility SHALL exclude detected palms and include thumbs.
- WHILE ready, WHEN exactly three valid contacts form a group, the utility SHALL establish its start baseline.
- WHILE ready, the utility SHALL allow staggered finger placement without a placement time limit.
- WHILE ready, WHEN contact count or identities change, the utility SHALL discard the previous start baseline.
- WHILE ready, WHEN exactly three valid contacts remain after a change, the utility SHALL establish a new start baseline.
- WHEN evaluating start movement, the utility SHALL use group displacement without requiring each finger to move.
- WHEN evaluating start movement, the utility SHALL exclude movement before baseline creation and shall not accumulate traveled distance.
- WHILE group displacement remains below the threshold, the utility SHALL refrain from starting a drag.
- WHERE no drag-start threshold is configured, the utility SHALL use a default threshold of 1 mm.
- WHEN reading configuration, the utility SHALL accept a drag-start threshold expressed as a finite number greater than zero, in millimeters.
- IF a configured threshold is zero, negative, nonnumeric, or nonfinite, THEN the utility SHALL reject that value.

The threshold applies in every direction and measures touchpad distance, not screen distance.
Detected palms do not change the valid contact count, group position, or baseline.
The 1 mm default requires comparison with macOS before any equivalence claim.
Invalid configuration fails startup under RF-1. RF-6 uses this same threshold for continuation.

#### Enablement and rearming

- WHERE no startup enablement setting is configured, the utility SHALL request enabled operation at startup.
- WHERE startup configuration requests disabled operation, the utility SHALL start with three-finger dragging disabled.
- WHEN startup, enablement, or recovery occurs, the utility SHALL require available access and completed cleanup before arming.
- WHEN evaluating readiness, the utility SHALL also require lifecycle availability.
- WHILE a selected-touchpad physical button is held, the utility SHALL prevent three-finger drag recognition.
- WHEN such a button is pressed outside a drag, the utility SHALL clear armed recognition and preserve the ordinary button action.
- WHEN arming recognition, the utility SHALL require all selected-touchpad physical buttons released and zero valid contacts.
- WHEN those prerequisites hold and zero valid contacts are observed, the utility SHALL arm recognition for fresh contacts.
- WHILE awaiting zero valid contacts, the utility SHALL prevent existing contacts from starting a drag.
- WHEN a new drag starts, the utility SHALL consume the current armed state.
- WHEN a drag completes normally, the utility SHALL require zero valid contacts before arming another drag.
- WHERE zero valid contacts remain at normal completion, the utility SHALL treat the zero-contact requirement as satisfied.
- WHEN the user requests enablement, the utility SHALL apply the request while running and perform the readiness checks.
- WHEN the user requests disablement, the utility SHALL disable new three-finger drags while running.
- WHILE running, the utility SHALL provide explicit enable and disable commands.
- WHEN an enable or disable command changes state, the utility SHALL retain that override only until restart.
- WHEN the utility restarts, the utility SHALL determine startup enablement from configuration rather than the previous runtime override.

All arming criteria include released physical buttons on the selected touchpad.
Arming permits subsequent fresh contacts to form a gesture. Readiness does not require zero contacts throughout recognition.
Runtime enable and disable commands do not rewrite startup configuration.

- WHEN enablement is requested while already enabled, the utility SHALL preserve any active gesture without resetting its recognition state.
- WHEN disablement is requested while already disabled, the utility SHALL preserve the disabled state without generating additional button events.
- WHEN enable, disable, or cancel commands repeat, the utility SHALL preserve any pending-cleanup obligation.

### RF-3: Move during a drag

- WHILE dragging with the same three contacts, WHEN subsequent group movement occurs, the utility SHALL move the pointer with it.
- WHILE dragging with three valid contacts, WHEN their group position moves, the utility SHALL keep the drag action held.
- WHILE dragging, the utility SHALL match ordinary one-finger pointing direction, sensitivity, and acceleration on the same touchpad.
- WHEN the group moves rightward, the utility SHALL move the pointer rightward.
- WHEN the group moves upward, the utility SHALL move the pointer upward.
- WHILE dragging, the utility SHALL ignore natural-scrolling direction settings for pointer movement.
- WHEN contact identities change while dragging with three contacts, the utility SHALL reset the motion baseline without moving the pointer.
- WHEN that baseline resets, the utility SHALL keep the drag held and use only subsequent group movement.
- WHILE three contacts remain stationary outside grace, the utility SHALL keep the drag held without a stationary-contact timeout.
- WHILE those contacts remain stationary, the utility SHALL stop gesture-driven pointer movement without starting a grace timer.

Individual fingers may remain stationary. RF-5 and RF-6 govern paused movement and continuation during the grace period.
Contact replacement while the count stays three does not require another displacement threshold or start a grace period.
If the count departs from three, RF-5 applies instead.
The pointer has no continued gesture-driven motion after the fingers stop.
Matching sensitivity and acceleration means matching the response to equivalent displacement and speed during ordinary one-finger pointing.
NFR-2 defines numerical comparison tolerances and response-time limits.
Other pointing devices may move the desktop pointer while the utility's primary-button hold remains active, including during grace.
The movement restrictions above apply to movement generated by this gesture, not other devices.

### RF-4: Drag files, select text, and move windows

- WHEN a drag starts over a draggable file item, the utility SHALL allow the application's ordinary file-drag action.
- WHEN a drag starts over selectable text, the utility SHALL allow the application's ordinary text-selection action.
- WHEN a drag starts over a draggable title bar, the utility SHALL allow the window to follow the pointer.

The application controls valid drop destinations and the result of a file drop.
The gesture starts at the current pointer position, not the fingers' absolute touchpad position.
All three operations use the logical primary-button action, rather than a fixed physical left-button action.

Window-movement acceptance covers floating windows with draggable title bars.
Tiled, maximized, and fullscreen windows follow normal Hyprland behavior.
Require movement wherever an ordinary primary-button title-bar drag works in the tested state.
The utility must not force a window into another state to make it movable.

#### Desktop changes during a drag

- WHILE the session remains available, WHEN monitors or display layout change, the utility SHALL retain the gesture's normal lifetime rules.
- WHEN the desktop repositions or constrains the pointer or windows, the utility SHALL accept the resulting positions without compensating movement.
- WHEN subsequent group movement occurs, the utility SHALL continue relative pointer movement from the desktop's resulting pointer position.
- WHEN the dragged window disappears or its application exits, the utility SHALL retain the gesture's normal movement and ending rules.
- WHEN a drag target disappears, the utility SHALL leave target handling to the desktop and application without selecting a replacement target.

Display changes do not extend a grace deadline or override lock, suspend, or session-loss termination.
Target disappearance does not guarantee undo. The utility does not restore the closed window or reverse application actions.

### RF-5: Pause a drag

- WHILE dragging outside the grace period, WHEN contact count departs from three, the utility SHALL start the grace period.
- WHEN the grace period starts, the utility SHALL set its deadline to the interruption time plus the configured duration.
- WHILE the grace period is active, the utility SHALL keep the drag action held.
- WHILE the grace period is active, the utility SHALL pause gesture-driven pointer movement until RF-6 applies.
- WHILE the grace period is active, WHEN contacts change, the utility SHALL preserve the existing deadline.
- WHILE three valid contacts remain stationary during the grace period, the utility SHALL keep the existing timer running.
- WHERE no grace duration is configured, the utility SHALL use 300 ms.
- WHEN reading configuration, the utility SHALL accept a grace duration in milliseconds as a finite number greater than zero.
- IF the configured grace duration is zero, negative, nonnumeric, or nonfinite, THEN the utility SHALL reject that value.

These criteria cover zero, one, two, or more than three contacts.
- WHEN a valid contact-count change is reported, the utility SHALL apply it without an additional debounce delay.
- WHEN even a brief departure from three valid contacts occurs during dragging, the utility SHALL start grace under RF-5.

Returning contacts follow RF-6. Detected palms remain excluded from the valid count.
Further changes, including lifting all remaining fingers, do not restart or extend the timer.
After successful continuation, a later departure from three starts a new grace period with a new deadline.
The default requires comparison with macOS. RF-1 rejects invalid configuration and applies file changes only after restart.

### RF-6: Continue a drag

- WHILE the grace period is active, WHEN exactly three valid contacts form a group, the utility SHALL establish its continuation baseline.
- WHILE the grace period is active, WHEN contact count or identities change, the utility SHALL discard the previous continuation baseline.
- WHEN exactly three valid contacts remain after that change, the utility SHALL establish a fresh continuation baseline during grace.
- WHEN qualifying continuation movement occurs strictly before the grace deadline, the utility SHALL continue the same drag.
- WHILE continuation displacement remains below the drag-start threshold, the utility SHALL preserve the pause and the grace deadline.
- WHEN a drag continues, the utility SHALL cancel the current grace period.
- WHEN a drag continues, the utility SHALL keep the drag action held without an intervening release.
- WHEN repositioned fingers resume movement, the utility SHALL prevent a pointer jump caused solely by their new contact positions.
- WHEN continuation qualifies, the utility SHALL discard movement through the threshold-crossing event and retain the current pointer position.
- WHEN continuation qualifies, the utility SHALL reset the motion baseline and use only subsequent group movement for pointer motion.

Continuation uses straight-line displacement of the group's average position in touchpad millimeters, not accumulated traveled distance.
Individual fingers may remain stationary. The shared threshold defaults to 1 mm and follows the drag-start configuration.
Contact changes reset the continuation baseline without changing the deadline. Detected palms do not change valid contacts.

### RF-7: End a drag

- WHEN the grace period expires, the utility SHALL release the drag action.
- WHILE the grace period is active, WHEN an ending tap occurs, the utility SHALL release the drag action immediately.
- WHEN an ending tap occurs, the utility SHALL consume the tap without generating an additional click.
- WHEN the utility releases the drag action, the utility SHALL end the active drag.
- WHEN movement coincides exactly with the grace deadline, the utility SHALL process expiry before evaluating continuation.
- WHEN candidate lift coincides exactly with the grace deadline, the utility SHALL end through expiry without recognizing an ending tap.
- WHEN expiry coincides with movement or candidate lift, the utility SHALL release the held drag action only once.
- WHEN disablement occurs during dragging or grace, the utility SHALL immediately terminate the gesture and stop its movement.
- WHEN disablement terminates a gesture, the utility SHALL cancel its timers and release the held button without waiting for grace expiry.
- IF release fails during disablement, THEN the utility SHALL apply RF-9's pending-cleanup rules.

Movement at or after the deadline cannot continue the expired drag.
Contacts remaining after normal completion cannot start another drag until all valid contacts lift.

#### Simultaneous termination and movement

- WHEN a termination event coincides with qualifying movement, the utility SHALL prevent that movement from starting or continuing a drag.
- WHEN simultaneous events terminate the same gesture, the utility SHALL avoid duplicate successful releases while preserving required cleanup retries.

Termination events include failure, lock, suspend preparation, exit, disablement, cancellation, and a selected-touchpad physical press.
RF-7's grace-expiry priority still applies. RF-7's physical-button consumption rules still apply to presses that terminate a gesture.

#### Physical-button termination and explicit cancellation

- WHILE dragging or in grace, WHEN a selected-touchpad physical button is pressed, the utility SHALL terminate the gesture immediately.
- WHEN that press terminates the gesture, the utility SHALL stop movement, cancel timers, and release its held button.
- WHEN that press terminates the gesture, the utility SHALL consume the physical press and its matching release without another click.
- WHILE running, the utility SHALL provide an explicit cancel command for the current gesture.
- WHEN cancellation is requested during dragging or grace, the utility SHALL stop movement, cancel timers, and release its held button.
- WHEN the user cancels a gesture, the utility SHALL preserve the feature's requested enablement state.
- WHEN cancellation is requested without an active gesture, the utility SHALL preserve enablement and generate no button events.
- WHEN physical-button termination or explicit cancellation occurs, the utility SHALL require fresh contacts under RF-2's rearming rules.
- IF release fails during either termination, THEN the utility SHALL apply RF-9's pending-cleanup rules.

Consuming the matching physical release remains necessary even if cleanup finishes or the feature becomes disabled before that release.
Physical-button termination releases only the utility's held action. Existing ordinary button actions retain their normal behavior.
Explicit cancellation ends the utility's gesture, not an application transaction. The application may treat release as a drop.
Neither cancellation nor physical-button termination guarantees undo.

#### Ending-tap recognition

- WHILE the grace period is active, WHEN contacts change from zero to one, the utility SHALL start a tap candidate.
- WHEN evaluating a tap candidate, the utility SHALL include thumbs and ignore detected palms.
- WHILE a tap candidate remains down, the utility SHALL defer ending-tap recognition until that contact lifts.
- WHEN evaluating tap duration, the utility SHALL measure elapsed time from the candidate's touchdown to its lift.
- WHEN evaluating tap movement, the utility SHALL use the maximum tap displacement throughout the candidate's contact.
- IF tap duration exceeds its configured limit, THEN the utility SHALL cancel the tap candidate.
- IF tap displacement exceeds its configured limit, THEN the utility SHALL cancel the tap candidate.
- WHEN a second valid contact appears, the utility SHALL cancel the tap candidate.
- WHEN a candidate is canceled, the utility SHALL require zero valid contacts before accepting another tap candidate.
- WHEN an uncanceled candidate lifts within both limits before grace expiry, the utility SHALL recognize an ending tap.
- WHEN evaluating either tap limit, the utility SHALL accept equality with that limit.
- WHEN a tap candidate touches down, the utility SHALL preserve the existing grace-period deadline.
- IF grace expiry precedes candidate lift, THEN the utility SHALL end the drag through expiry without recognizing an ending tap.
- WHEN grace expires with a tap candidate still down, the utility SHALL consume its later lift without generating an ordinary click.
- WHERE no ending-tap duration limit is configured, the utility SHALL use 200 ms.
- WHERE no ending-tap movement limit is configured, the utility SHALL use 1 mm.
- WHEN reading configuration, the utility SHALL accept independent ending-tap duration and movement limits as finite numbers greater than zero.
- IF either configured limit is zero, negative, nonnumeric, or nonfinite, THEN the utility SHALL reject that value.

Both touchdown and lift must occur during the grace period, before its expiry.
A remaining drag finger cannot become a tap candidate without first lifting.
Exceeding a movement limit cancels the candidate even if the finger returns to its initial position.
Canceling a candidate does not itself end the drag. The existing drag and grace-period rules still apply.
The late lift completes the old contact sequence. Rearming still requires all readiness checks, including zero valid contacts.
Duration uses milliseconds. Movement uses millimeters and remains independent of the drag-start threshold.
The defaults require comparison with macOS before an equivalence claim.
Expiry takes priority when candidate lift coincides exactly with the grace deadline.
RF-1 rejects invalid configuration and applies configuration-file changes only after restart.

The application determines the result of the release. A release does not guarantee that a file drop succeeds.

### RF-8: Preserve ordinary gestures

- WHILE no drag or release is pending, the utility SHALL preserve ordinary pointing and scrolling when touchpad access is available.
- WHILE release remains pending, the utility SHALL suppress input from the affected touchpad.
- WHEN suppressing input during cleanup, the utility SHALL discard it without later replay.
- WHILE suppressing the affected touchpad, the utility SHALL leave other devices outside that restriction.
- WHEN cleanup completes and touchpad access is available, the utility SHALL restore ordinary one-finger pointing and two-finger scrolling.
- WHEN ordinary input returns after cleanup, the utility SHALL restore it independently of three-finger drag availability.
- WHILE awaiting fresh contacts after cleanup, the utility SHALL preserve ordinary pointing and scrolling when access is available.
- WHILE disabled with cleanup complete, the utility SHALL stop reserving three-finger gestures and allow normal desktop gesture behavior.
- WHILE the feature is enabled, the utility SHALL reserve the selected touchpad's three-finger gestures for this feature.
- WHILE the feature is enabled, the utility SHALL prevent its three-finger gestures from also triggering competing desktop actions.

#### Other devices and duplicate instances

- WHILE enabled, the utility SHALL provide three-finger dragging on only one selected touchpad at a time.
- WHEN evaluating a gesture, the utility SHALL use contacts from the selected touchpad only.
- WHILE operating, the utility SHALL preserve normal desktop behavior on other touchpads.
- WHEN another pointing device produces input, the utility SHALL let the desktop process it without intercepting its events.
- WHEN another pointing device produces input, the utility SHALL leave the current gesture active under its existing termination rules.
- IF another utility instance already runs in the same desktop session, THEN the new instance SHALL exit with an error.
- WHEN rejecting a duplicate instance, the new instance SHALL leave the existing instance and its gesture undisturbed.
- WHEN enable, disable, or cancel commands are issued, the utility SHALL direct them to the existing instance.

Three-finger gesture reservation applies only to the selected touchpad. Contacts from different devices never combine into a gesture.

### RF-9: End a drag after failure

- IF a drag failure occurs, THEN the utility SHALL immediately terminate any active gesture and stop its drag movement.
- WHEN a failure terminates a gesture, the utility SHALL cancel its timers and discard its queued movement.
- WHEN a gesture terminates through failure, the utility SHALL prevent that gesture from resuming.
- IF a drag failure permits release, THEN the utility SHALL release the held drag action immediately.
- IF a drag failure prevents release, THEN the utility SHALL track a separate pending-release obligation.
- WHEN release becomes possible in the owning session, the utility SHALL release the held button at the earliest opportunity.
- WHEN the button is released, the utility SHALL mark cleanup complete.
- WHEN the desktop confirms the owning virtual device or session no longer exists, the utility SHALL mark cleanup complete.
- IF device or session loss lacks confirmation, THEN the utility SHALL retain the pending-release obligation.
- WHEN a new session appears, the utility SHALL prevent delivery of an old session's pending release into it.
- WHEN a drag failure occurs, the utility SHALL report an English error identifying the failed operation, cause, and pending-release status.
- WHERE recovery guidance is available, the utility SHALL include it in the failure error.
- WHEN cleanup attempts repeat without a changed failure, the utility SHALL suppress repeated copies of the same error.
- WHEN failure occurs, the utility SHALL clear armed recognition and preserve the user's requested enablement state.
- WHEN access returns and cleanup completes, the utility SHALL automatically prepare enabled dragging for fresh contacts.
- WHILE the user has explicitly disabled dragging, the utility SHALL prevent automatic recovery from enabling it.
- WHEN recovery prerequisites hold, the utility SHALL require an observed zero-contact state before arming a new gesture.
- IF contact state becomes unreliable or output cannot be delivered, THEN the utility SHALL terminate the gesture and apply pending cleanup.
- WHEN recovering after synchronization loss or output failure, the utility SHALL require reliable state and fresh-contact readiness checks.

Ending a failed drag does not undo actions that an application already completed.
Failure termination and output cleanup are separate outcomes. Unconfirmed loss is not proof that a held button no longer exists.
Pending cleanup keeps the affected touchpad suppressed even though the gesture has terminated.
After cleanup, restore ordinary input as soon as touchpad access permits it.
Automatic recovery requires enabled operation, available access, completed cleanup, and zero valid contacts before a new gesture.
It never resumes a terminated gesture. Ordinary input restoration remains independent of drag readiness.

#### Lifecycle interruptions

- WHEN the session locks or the system prepares to suspend, the utility SHALL terminate any active gesture and block new drags.
- WHEN that interruption occurs, the utility SHALL stop gesture movement, cancel timers, and release its held button.
- IF release remains unresolved, THEN the utility SHALL apply the pending-cleanup rules above.
- WHILE locked, preparing for suspend, or suspended, the utility SHALL prevent new three-finger drags.
- WHEN the session unlocks or the system resumes, the utility SHALL apply RF-2's readiness and fresh-contact checks.
- WHEN normal exit is requested, the utility SHALL stop gesture movement, cancel timers, and release any held button.
- WHEN normal exit cleanup completes, the utility SHALL restore normal touchpad handling before exiting.
- IF release fails during normal exit, THEN the utility SHALL apply pending cleanup rather than report a successful shutdown.
- IF the process crashes or is forcibly terminated, THEN the system SHALL clear the utility's held button state.
- IF the process crashes or is forcibly terminated, THEN the system SHALL restore normal touchpad handling.
- WHEN the utility restarts after termination, the utility SHALL follow startup configuration and fresh-contact readiness checks.
- WHEN the desktop session ends or its connection is lost, the utility SHALL terminate the gesture and block new drags.
- WHEN session loss occurs, the utility SHALL apply pending cleanup and prevent old input from entering a new session.
- WHEN recovering from a lifecycle interruption, the utility SHALL require an available, unlocked session before arming recognition.

Recovery also requires completed cleanup, required access, zero valid contacts, and released physical touchpad buttons.
The user's requested enablement state still applies under RF-2 and RF-9.
Crash and forced-termination cleanup depend on system behavior because the terminated process cannot execute cleanup.
Acceptance tests must verify those system outcomes. They must not assume successful cleanup from process death alone.

## Non-functional requirements

- **NFR-1: Behavior match.** Compare macOS outcomes for drag start, repositioning, continuation, tap-to-end, and release. Unresolved behavioral mismatches block release.
- **NFR-2: Responsiveness and motion.** Meet the timing limits, measurement conditions, and ordinary-pointing tolerances below.
- **NFR-3: Verification.** Use deterministic unit tests for gesture transitions and timing. Use integration tests for backend interactions on both supported distributions.
- **NFR-4: Regression prevention.** Add regression tests for defect fixes. Run applicable tests before completion.
- **NFR-5: Privacy.** Keep gesture state in memory. Never persist input events.
- **NFR-6: Persistence.** Persist only explicit configuration in the XDG configuration directory, using atomic writes. Interrupted saves must preserve complete valid configuration.
- **NFR-7: Language.** Use English for identifiers, comments, documentation, messages, and errors.
- **NFR-8: Errors.** Error messages must identify the failed operation and cause. Include actionable guidance where available.

### Responsiveness and comparison acceptance

Coverage: RF-2, RF-3, RF-6, RF-7, and RF-9. These are acceptance limits, not measured results or macOS timing values.

| Metric | 95th-percentile limit | Maximum observed sample |
| --- | --- | --- |
| Utility processing delay | 5 ms | 20 ms |
| Desktop response | 50 ms | 100 ms |

Equality with a limit passes, subject to measurement uncertainty. Each metric must satisfy both its percentile and maximum limits.

| Operation | Timing origin | Processing endpoint | Desktop endpoint |
| --- | --- | --- | --- |
| Drag start | Qualifying threshold-crossing input event | Submission of the primary-button hold | Independently observed desktop button hold |
| Motion | Subsequent group-movement event | Submission of corresponding pointer output | Independently observed pointer change |
| Continuation transition | Qualifying continuation event before grace expiry | Completion of the continuation state transition | No pointer output is expected at this event under T04 |
| Continued pointer motion | First movement event after qualifying continuation | Submission of corresponding pointer output | Independently observed pointer change |
| Ending-tap release | Qualifying finger-lift event | Submission of button release | Independently observed desktop button release |
| Grace expiry | Fixed grace deadline | Submission of button release | Independently observed desktop button release |
| Other immediate release | Qualifying termination event or accepted command | Submission of button release | Independently observed desktop button release |
| Recoverable failure cleanup | Time release becomes possible in the owning session | Submission of release or its retry | Independently observed desktop button release |

Exclude the intentional grace interval by measuring expiry response from its deadline.
Exclude unavailable-output waiting from cleanup latency by measuring from the release opportunity.
These exclusions do not change gesture deadlines, expiry precedence, or pending-cleanup rules.
Continuation's internal transition uses the processing limits. Its first subsequent movement uses both timing limits.

Collect at least 100 samples per operation, distribution, and application backend for each timing check.
Calculate each result separately. Do not pool operations, distributions, or native Wayland and XWayland results.
Include slow samples. Keep only aggregate timings, sample counts, and measurement details after evaluation.
Never persist input-event recordings.

Use a warmed-up session with no deliberate background workload.
Keep power mode, display refresh rate, and settings fixed throughout each run.
Record refresh rate, measurement precision, and uncertainty with the results.
Use monotonic timestamps for processing delay and an independent observer for desktop response.
Successful output submission alone does not establish desktop response or completed cleanup.
If uncertainty could change a result from pass to fail, report it as inconclusive rather than passed.

Compare three-finger pointer movement with ordinary one-finger pointing on the same Linux touchpad.
Use equivalent touchpad travel distances and speeds, with the same settings.
Allow a distance difference of at most 5% of the one-finger result or 2 logical pixels, whichever is larger.
Direction must match. This tolerance does not permit movement caused solely by contact replacement or repositioning.
Exclude discarded threshold-crossing movement from the active-motion comparison.

Use the same interaction scenarios for native macOS outcome comparisons under T09's fixed baseline.
Report macOS timing differences separately. They do not replace the approved Linux timing limits.
Linux sensitivity remains relative to ordinary Linux pointing, as T04 requires.
Review alone does not establish macOS equivalence. Resolve behavioral mismatches before release under the current constitution.
Accepting different behavior requires explicit changes to both the constitution and specification first.
Timing differences remain subject to T10's separate timing rules.

## Limit cases

| Case | Required result |
| --- | --- |
| The selected touchpad disconnects and another compatible device is present | Apply cleanup and wait for the same device. Do not substitute the other device. |
| The same compatible touchpad returns | Recover only after cleanup, reliable-state, and readiness checks pass. |
| A valid contact-count change lasts only briefly | Apply it immediately without added debounce. Returning contacts follow continuation rules. |
| A termination event coincides with qualifying movement | Termination prevents drag start or continuation at that timestamp. |
| Input synchronization is lost or an output write fails | Terminate and apply pending cleanup. Recover only with reliable state and fresh contacts. |
| Enable repeats while a drag is active | Preserve the drag and recognition state. |
| Disable repeats while already disabled | Remain disabled without additional button events. |
| Cancel arrives with no active gesture | Preserve enablement and generate no button events. Do not clear pending cleanup. |
| A tap candidate lifts after grace already expired | Consume the lift without an ordinary click or another drag release. Apply normal rearming checks. |
| Configuration is missing or a setting is omitted | Use the applicable defaults without creating or rewriting configuration. |
| Existing configuration is unreadable, malformed, or invalid | Fail startup with an actionable English error. Preserve normal touchpad handling. |
| A setting repeats, has an unknown name, or uses a string instead of its required type | Reject configuration rather than infer or convert a value. |
| A positive numeric value overflows or converts to zero | Fail validation and report the supported range without clamping. |
| Configuration changes during dragging, grace, or idle operation | Keep loaded settings until restart. Runtime commands retain their existing behavior. |
| Selection is omitted with multiple compatible touchpads | Fail startup and list the choices. |
| An explicit selection is missing or incompatible | Fail startup and report the selection and reason without substituting another device. |
| No compatible touchpad exists | Fail startup while preserving normal touchpad handling. |
| Three stationary fingers while ready | No generated click or drag. |
| Contact count differs from three while ready | No three-finger drag starts. |
| Three contacts arrive separately | Establish the baseline when the third valid contact arrives, without a placement time limit. |
| Movement occurs before the third valid contact arrives | Exclude that movement from group displacement. |
| One finger moves while two remain stationary | Start when their average position reaches the configured displacement threshold. |
| Group displacement equals the threshold | Start the drag. |
| A movement event crosses the start or continuation threshold | Discard all movement through that event. Move only on subsequent group movement. |
| Left-handed mode swaps physical buttons | Hold the logical primary button for normal selection and dragging. |
| Natural-scrolling direction changes | Preserve ordinary pointer direction. |
| Contact identities change during dragging while the count stays three | Keep the hold, reset the motion baseline, and discard replacement displacement. |
| Three contacts stop moving outside grace | Stop pointer movement. Keep the drag held indefinitely without starting a grace timer. |
| Back-and-forth movement stays below the threshold | No drag starts, regardless of total distance traveled. |
| Valid contact count or identities change before drag start | Discard the baseline. Establish a new baseline when exactly three valid contacts remain. |
| A detected palm touches beside three valid contacts | Exclude the palm. Keep the existing baseline. |
| One thumb and two other fingers form the group | Count all three contacts unless a contact is a detected palm. |
| A configured threshold is zero, negative, nonnumeric, or nonfinite | Reject configuration and fail startup while preserving normal touchpad handling. |
| Contact count first departs from three during dragging | Start the grace period, including when all fingers lift. |
| Contacts change again during grace, including all remaining fingers lifting | Preserve the original deadline. |
| Qualifying continuation movement occurs before expiry | Continue the same drag without an intervening release and cancel the timer. |
| Three fingers return but remain stationary | Preserve the pause and deadline. |
| Continuation displacement stays below the shared threshold | Preserve the pause and deadline. |
| Contact identities change during grace while three contacts remain | Reset the continuation baseline without changing the deadline. |
| Dragging resumes, then the contact count departs from three again | Start a new grace period from the new interruption. |
| No grace duration is configured | Use 300 ms. |
| Grace duration is zero, negative, nonnumeric, or nonfinite | Reject configuration and fail startup while preserving normal touchpad handling. |
| Grace expires with contacts remaining | End the drag. Require all valid contacts to lift before a new RF-2 gesture. |
| A drag completes with zero valid contacts | The zero-contact requirement is already satisfied for the next gesture. |
| Startup or enablement occurs with contacts already down | Wait for access, cleanup, and zero valid contacts before arming. |
| Failure clears while enabled, with cleanup complete and access available | Automatically arm after observing zero valid contacts. Never resume the failed gesture. |
| Failure clears while explicitly disabled | Keep dragging disabled. Restore ordinary input when cleanup and access permit. |
| Disablement occurs during dragging or grace | End immediately, cancel timers, and release. Use pending cleanup if release fails. |
| Disablement occurs while release remains pending | Keep the touchpad suppressed until cleanup completes. Disablement does not bypass cleanup. |
| Dragging is disabled and cleanup is complete | Allow the desktop's normal three-finger gestures. |
| A runtime override differs from startup configuration | Keep the override until restart, then use startup configuration. |
| A valid ending tap completes during the grace period | The drag ends on finger lift, without an additional click. |
| One drag finger remains and then lifts | No ending tap occurs. A fresh contact requires all valid contacts to lift first. |
| A candidate lasts exactly 200 ms with default limits | Accept its duration if it lifts before grace expiry. |
| A candidate reaches exactly 1 mm with default limits | Accept its movement if displacement never exceeds 1 mm. |
| A candidate exceeds 1 mm and returns with default limits | Cancel the candidate. Returning does not restore it. |
| A second valid contact appears and then lifts | Cancel the candidate. Require all valid contacts to lift before another attempt. |
| A detected palm appears beside the candidate | Ignore the palm without canceling the candidate. |
| A candidate lasts longer than its duration limit | Cancel the candidate. Do not treat the later lift as an ending tap. |
| The grace period expires while a candidate remains down | End the drag through expiry. Touchdown does not extend the deadline. |
| Candidate lift coincides exactly with grace expiry | Expiry wins. Release once without recognizing an ending tap. |
| The pointer starts over an area without drag support | The application's ordinary behavior applies. |
| A file drops over an invalid destination | The application determines the result. |
| A failure prevents immediate release | Terminate the gesture and retain a separate pending release. Suppress input from the affected touchpad. |
| Touchpad input arrives during pending cleanup | Discard it without replay. Leave other devices outside the restriction. |
| The owning device or session disappears without confirmation | Keep cleanup pending. Do not assume release succeeded. |
| The desktop confirms that the owning virtual device or session no longer exists | Complete cleanup without sending an old release into a new session. |
| A new session appears while old cleanup remains pending | Do not send the old release into it. Keep the cleanup obligation until completion criteria hold. |
| Cleanup completes while touchpad access remains unavailable | Restore ordinary input when access becomes available. |
| Cleanup completes while three-finger dragging remains unavailable | Restore ordinary pointing and scrolling when access permits. |
| Repeated cleanup attempts encounter the same failure | Do not repeatedly emit the same error. |
| Movement coincides with grace-period expiry | Expiry wins. Release once without continuing the drag. |
| A selected-touchpad physical button is pressed during dragging or grace | End the gesture, release its hold, and consume that press and its matching release. |
| A consumed physical release arrives after cleanup or disablement | Consume it without forwarding an unmatched release or generating another click. |
| A physical touchpad button is already held | Preserve ordinary button behavior. Block recognition until all buttons release and zero valid contacts are observed. |
| A cancel command arrives during dragging or grace | Stop movement, cancel timers, and release. Keep requested enablement and require fresh contacts. |
| Release fails after cancellation or physical-button termination | Apply pending cleanup and affected-touchpad suppression under RF-9. |
| Another pointer moves or produces button input during a drag | Let the desktop process it normally. Do not terminate the utility's gesture. |
| Contacts appear on multiple touchpads | Recognize only the selected touchpad's contacts. Preserve normal behavior on the others. |
| A second instance starts in the same desktop session | Exit with an error without disturbing the existing instance. |
| Session lock or suspend preparation interrupts dragging or grace | Terminate, stop movement, cancel timers, release, and block new drags. Apply pending cleanup when needed. |
| Unlock or resume occurs with contacts or physical buttons held | Wait for an available unlocked session and all readiness checks before arming fresh recognition. |
| Normal exit occurs during a drag | Release the held button and restore normal touchpad handling before exit. |
| Release fails during normal exit | Retain pending cleanup. Do not report successful shutdown. |
| The process crashes or receives forced termination | Verify system cleanup of held button state and restoration of normal touchpad handling. |
| The utility restarts after termination | Use startup configuration and fresh-contact checks. Never resume the old gesture. |
| Session loss interrupts a drag | Terminate and block new drags. Apply pending cleanup without sending old input into another session. |
| A monitor disconnects or display layout changes while the session remains available | Keep normal gesture rules. Accept desktop repositioning without compensating pointer movement. |
| The dragged window closes or its application exits | Keep normal gesture rules. Let the desktop handle target disappearance without selecting a replacement. |

## Out of scope

- Standalone X11 desktop sessions. XWayland applications within Hyprland remain in scope.
- Touchpads that lack RF-1's required contact tracking, fourth-contact detection, physical resolution, or palm-rejection information.
- Development builds in the acceptance baseline.
- Desktops other than Hyprland.
- Distributions other than Arch Linux and CachyOS.
- Window movement from arbitrary areas of a window.
- Custom gesture mappings.
- A graphical settings interface.
- Per-application rules.
- Concurrent use of competing three-finger desktop gestures.
- Three-finger drag recognition on multiple touchpads at once.

## Criteria of completion

- Every functional acceptance criterion passes on both supported distributions.
- Thunar file dragging and Mousepad text selection and title-bar checks pass under verified native Wayland and XWayland backends.
- Each tested touchpad meets RF-1's capability requirements, with device identity and capability evidence recorded.
- Linux versions remain fixed for each run. The Mac model, trackpad, macOS build, and settings remain fixed for comparison.
- Finder and TextEdit provide the native macOS reference results.
- Tests cover drag start, finger-count changes, continuation, tap-to-end, and grace-period expiry.
- Tests cover startup state, runtime overrides, immediate disablement, automatic recovery, and fresh-contact rearming.
- Tests cover physical-button consumption, existing button actions, other-device independence, selected-touchpad isolation, cancellation, and duplicate-instance rejection.
- Tests cover lock, suspend, normal exit, crash, forced termination, session loss, display changes, and target disappearance.
- Crash and forced-termination checks verify system cleanup and restored touchpad handling without relying on process cleanup code.
- Failure checks confirm immediate gesture termination and separate cleanup through release or confirmed owner disappearance.
- Cleanup checks verify touchpad suppression, discarded input, session-bound releases, and error reporting without repeated identical errors.
- Ordinary pointing and scrolling work outside active drags and pending cleanup when touchpad access is available.
- Acceptance checks show no competing three-finger desktop actions.
- Each timing check meets NFR-2's percentile and maximum limits with at least 100 samples under the specified conditions.
- Independent desktop observations support response measurements. Inconclusive results do not count as passes.
- Pointer-distance differences meet the ordinary-pointing tolerance, with matching direction and no repositioning-induced movement.
- Privacy and configuration checks satisfy NFR-5 and NFR-6.
- Configuration checks verify defaults, strict validation, conversion bounds, restart-only changes, and unambiguous touchpad selection.
- Documentation, messages, and errors satisfy NFR-7 and NFR-8.
- The macOS comparison records gesture outcomes and separate timing differences. No unresolved behavioral mismatch remains at release.
- Different accepted behavior requires explicit constitution and specification changes before release. Review alone cannot waive equivalence.
- Tests cover same-device recovery, immediate contact transitions, simultaneous termination, repeated commands, and late tap-lift consumption.
- The team resolves release-blocking questions. The user explicitly accepts any deferred questions.

## Resolved decisions

- **T12:** Same-device recovery, termination priority, immediate contact transitions, failure coverage, repeated commands, and late tap-lift rules complete the consistency review.
- **T11:** RF-1 defines defaults, strict validation, restart-only configuration, and explicit or unambiguous automatic touchpad selection.
- **OQ-1 (T03):** Default to a configurable 300 ms grace period. Accept only finite numeric durations greater than zero.
- **OQ-2 (T02):** Use a fresh one-finger tap, recognized on lift. RF-7 defines configurable 200 ms and 1 mm limits.
- **OQ-3 (T01):** Use group displacement with a configurable 1 mm default. RF-2 defines contact validity, staggered placement, and baseline resets.
- **OQ-4 (T03):** Preserve each grace deadline until continuation or termination. Require qualifying movement before expiry. Expiry wins exact-time ties.
- **OQ-5 (T06):** Recover enabled dragging automatically after access, cleanup, and zero-contact checks. Never resume a terminated gesture.
- **OQ-6 (T06):** Default startup to enabled. Runtime enable/disable commands override configuration until restart. Disablement terminates active drags immediately.
- **OQ-7 (T07):** Touchpad presses terminate and are consumed. Other devices remain independent. Use one selected touchpad and reject duplicate instances.
- **OQ-8 (T09):** RF-1 defines device capabilities, fixed stable versions, Linux applications and backends, and the native macOS comparison policy.
- **OQ-9 (T10):** NFR-2 defines processing and desktop timing limits, sample conditions, uncertainty handling, and ordinary-pointing tolerance.
- **T04:** Match ordinary pointing and the logical primary button. Discard threshold-crossing movement and reset motion baselines after contact replacement.
- **T05:** Separate failed-gesture termination from pending release. Suppress the affected touchpad until cleanup completes, then restore ordinary input independently.
- **T08:** Lifecycle interruptions terminate gestures and require fresh-contact recovery. Display changes and target disappearance retain normal gesture rules.

## [OPEN QUESTIONS]

No product-behavior questions remain from T01 through T12.
Backend feasibility and actual acceptance results remain unverified. Technical contracts must document numeric ranges and device identity before implementation uses them.
