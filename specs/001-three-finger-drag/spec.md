# Specification: Three-Finger Drag

## Context and objective

Touchpad users need to drag files, select text, and move windows without a physical press.
The objective is to match the macOS three-finger drag experience.
The first release supports Arch Linux and CachyOS with Hyprland on Wayland.

A brief grace period lets users lift or reposition their fingers without releasing a drag.
The team must resolve timing details and compare observable behavior with macOS before claiming an exact match.

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

- **Ready:** The feature is enabled, its required access is available, and no drag is active.
- **Active drag:** The utility holds the drag action, including during the grace period.
- **Grace period:** A bounded interval that keeps the drag held and pauses gesture-driven pointer movement.
- **Ending tap:** A fresh one-finger tap that satisfies RF-7 during the grace period. Recognition occurs on finger lift.
- **Tap candidate:** A fresh single valid contact after all valid contacts have lifted during the drag's grace period.
- **Tap displacement:** The straight-line distance from the candidate's initial contact position, measured in touchpad millimeters.
- **Supported environment:** Arch Linux or CachyOS with Hyprland on Wayland and a compatible touchpad.
- **Drag failure:** Touchpad disconnection, missing or revoked permissions, or loss of the desktop connection.
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

Each requirement has acceptance criteria in EARS notation.
The words WHERE, WHILE, WHEN, IF, THEN, and SHALL identify conditions and required responses.

### RF-1: Supported environments

- WHERE the environment is supported, the utility SHALL provide the three-finger drag behavior in this specification.

Acceptance checks must cover both distributions. Compatible touchpads remain an open question.

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
Invalid-configuration recovery remains open under T11. RF-6 uses this same threshold for continuation.

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
T10 defines numerical comparison tolerances and response-time limits.

### RF-4: Drag files, select text, and move windows

- WHEN a drag starts over a draggable file item, the utility SHALL allow the application's ordinary file-drag action.
- WHEN a drag starts over selectable text, the utility SHALL allow the application's ordinary text-selection action.
- WHEN a drag starts over a draggable title bar, the utility SHALL allow the window to follow the pointer.

The application controls valid drop destinations and the result of a file drop.
The gesture starts at the current pointer position, not the fingers' absolute touchpad position.
All three operations use the logical primary-button action, rather than a fixed physical left-button action.

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
Further changes, including lifting all remaining fingers, do not restart or extend the timer.
After successful continuation, a later departure from three starts a new grace period with a new deadline.
The default requires comparison with macOS. T11 defines invalid-configuration recovery and changes during a drag.

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

Movement at or after the deadline cannot continue the expired drag.

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
- WHERE no ending-tap duration limit is configured, the utility SHALL use 200 ms.
- WHERE no ending-tap movement limit is configured, the utility SHALL use 1 mm.
- WHEN reading configuration, the utility SHALL accept independent ending-tap duration and movement limits as finite numbers greater than zero.
- IF either configured limit is zero, negative, nonnumeric, or nonfinite, THEN the utility SHALL reject that value.

Both touchdown and lift must occur during the grace period, before its expiry.
A remaining drag finger cannot become a tap candidate without first lifting.
Exceeding a movement limit cancels the candidate even if the finger returns to its initial position.
Canceling a candidate does not itself end the drag. The existing drag and grace-period rules still apply.
Duration uses milliseconds. Movement uses millimeters and remains independent of the drag-start threshold.
The defaults require comparison with macOS before an equivalence claim.
Expiry takes priority when candidate lift coincides exactly with the grace deadline.
T11 defines invalid-configuration recovery and changes to configuration during a drag.

The application determines the result of the release. A release does not guarantee that a file drop succeeds.

### RF-8: Preserve ordinary gestures

- WHILE no drag is active, the utility SHALL preserve ordinary one-finger pointing and two-finger scrolling.
- WHILE the feature is enabled, the utility SHALL reserve three-finger gestures for this feature.
- WHILE the feature is enabled, the utility SHALL prevent its three-finger gestures from also triggering competing desktop actions.

### RF-9: End a drag after failure

- IF a drag failure occurs, THEN the utility SHALL end any active drag.
- IF a drag failure permits release, THEN the utility SHALL release the held drag action immediately.
- IF a drag failure prevents release, THEN the utility SHALL release the held drag action at the earliest opportunity.

Ending a failed drag does not undo actions that an application already completed.
Failure recovery remains an open question. Recovery must not resume the interrupted drag.

## Non-functional requirements

- **NFR-1: Behavior match.** Compare repeatable drag scenarios with macOS. Record differences before claiming equivalent behavior.
- **NFR-2: Response time.** Measure start, movement, continuation, and release delays against agreed limits. These limits remain open.
- **NFR-3: Verification.** Use deterministic tests for gesture transitions and timing. Test desktop interactions on both supported distributions.
- **NFR-4: Regression prevention.** Add regression tests for defect fixes. Run applicable tests before completion.
- **NFR-5: Privacy.** Keep gesture state transient. Never persist input events.
- **NFR-6: Persistence.** Persist only explicit user configuration. An interrupted configuration save must preserve a complete valid configuration.
- **NFR-7: Language.** Use English for documentation, messages, and errors.
- **NFR-8: Errors.** Error messages must identify the failed operation and cause. Include actionable guidance where available.

## Limit cases

| Case | Required result |
| --- | --- |
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
| A configured threshold is zero, negative, nonnumeric, or nonfinite | Reject the value. Configuration recovery remains open under T11. |
| Contact count first departs from three during dragging | Start the grace period, including when all fingers lift. |
| Contacts change again during grace, including all remaining fingers lifting | Preserve the original deadline. |
| Qualifying continuation movement occurs before expiry | Continue the same drag without an intervening release and cancel the timer. |
| Three fingers return but remain stationary | Preserve the pause and deadline. |
| Continuation displacement stays below the shared threshold | Preserve the pause and deadline. |
| Contact identities change during grace while three contacts remain | Reset the continuation baseline without changing the deadline. |
| Dragging resumes, then the contact count departs from three again | Start a new grace period from the new interruption. |
| No grace duration is configured | Use 300 ms. |
| Grace duration is zero, negative, nonnumeric, or nonfinite | Reject the value. Configuration recovery remains open under T11. |
| The grace period expires | The drag ends. Later movement must satisfy RF-2 to start another drag. |
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
| A failure prevents immediate release | The active drag ends. The utility releases the held drag action at the earliest opportunity. |
| Movement coincides with grace-period expiry | Expiry wins. Release once without continuing the drag. |
| A physical button or another pointing device acts during a drag | Interaction rules remain open. |

## Out of scope

- X11.
- Desktops other than Hyprland.
- Distributions other than Arch Linux and CachyOS.
- Window movement from arbitrary areas of a window.
- Custom gesture mappings.
- A graphical settings interface.
- Per-application rules.
- Concurrent use of competing three-finger desktop gestures.

## Criteria of completion

- Every functional acceptance criterion passes on both supported distributions.
- Agreed representative applications support file dragging, text selection, and title-bar window movement in acceptance checks.
- Tests cover drag start, finger-count changes, continuation, tap-to-end, and grace-period expiry.
- Failure checks confirm drag termination and release wherever the environment permits release.
- Ordinary pointing and scrolling work outside an active drag.
- Acceptance checks show no competing three-finger desktop actions.
- Response-time measurements meet the agreed limits.
- Privacy and configuration checks satisfy NFR-5 and NFR-6.
- Documentation, messages, and errors satisfy NFR-7 and NFR-8.
- The macOS comparison records observable differences and the user accepts any remaining differences.
- The team resolves release-blocking questions. The user explicitly accepts any deferred questions.

## Resolved decisions

- **OQ-1 (T03):** Default to a configurable 300 ms grace period. Accept only finite numeric durations greater than zero.
- **OQ-2 (T02):** Use a fresh one-finger tap, recognized on lift. RF-7 defines configurable 200 ms and 1 mm limits.
- **OQ-3 (T01):** Use group displacement with a configurable 1 mm default. RF-2 defines contact validity, staggered placement, and baseline resets.
- **OQ-4 (T03):** Preserve each grace deadline until continuation or termination. Require qualifying movement before expiry. Expiry wins exact-time ties.
- **T04:** Match ordinary pointing and the logical primary button. Discard threshold-crossing movement and reset motion baselines after contact replacement.

## [OPEN QUESTIONS]

- **OQ-5:** After a failure, must users explicitly enable the feature again? Can availability recover automatically without resuming the interrupted drag?
- **OQ-6:** How do users enable or disable the feature? What happens when they disable it during a drag?
- **OQ-7:** What happens when users press a physical button or use another pointing device during a drag?
- **OQ-8:** Which touchpads, applications, and macOS reference behavior define the acceptance baseline?
- **OQ-9:** What response-time limits and numerical tolerances verify the approved ordinary-pointing match?
