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
- **Ending tap:** A tap during the grace period that ends the drag. Its recognition rules remain open.
- **Supported environment:** Arch Linux or CachyOS with Hyprland on Wayland and a compatible touchpad.
- **Drag failure:** Touchpad disconnection, missing or revoked permissions, or loss of the desktop connection.

Each requirement has acceptance criteria in EARS notation.
The words WHERE, WHILE, WHEN, IF, THEN, and SHALL identify conditions and required responses.

### RF-1: Supported environments

- WHERE the environment is supported, the utility SHALL provide the three-finger drag behavior in this specification.

Acceptance checks must cover both distributions. Compatible touchpads remain an open question.

### RF-2: Start a drag

- WHILE ready, WHEN exactly three fingers move, the utility SHALL start a drag at the current pointer position.
- WHEN a three-finger drag starts, the utility SHALL require no physical touchpad press.
- WHILE ready, WHEN three fingers remain stationary, the utility SHALL generate neither a click nor a drag.
- WHILE ready, WHEN the contact count differs from three, the utility SHALL refrain from starting a three-finger drag.

### RF-3: Move during a drag

- WHILE a drag is active, WHEN exactly three fingers move, the utility SHALL move the pointer with their movement.
- WHILE exactly three fingers move during a drag, the utility SHALL keep the drag action held.

### RF-4: Drag files, select text, and move windows

- WHEN a drag starts over a draggable file item, the utility SHALL allow the application's ordinary file-drag action.
- WHEN a drag starts over selectable text, the utility SHALL allow the application's ordinary text-selection action.
- WHEN a drag starts over a draggable title bar, the utility SHALL allow the window to follow the pointer.

The application controls valid drop destinations and the result of a file drop.
The gesture starts at the current pointer position, not the fingers' absolute touchpad position.

### RF-5: Pause a drag

- WHILE a drag is active, WHEN all fingers lift, the utility SHALL start the grace period.
- WHEN an active drag's contact count changes from three, the utility SHALL start the grace period.
- WHILE the grace period is active, the utility SHALL keep the drag action held.
- WHILE the grace period is active, the utility SHALL pause gesture-driven pointer movement until RF-6 applies.

These criteria cover zero, one, two, or more than three contacts.
Contact changes within an existing grace period have unresolved timer rules.

### RF-6: Continue a drag

- WHILE the grace period remains unexpired, WHEN exactly three fingers resume movement, the utility SHALL continue the same drag.
- WHEN a drag continues, the utility SHALL cancel the current grace period.
- WHEN a drag continues, the utility SHALL keep the drag action held without an intervening release.
- WHEN repositioned fingers resume movement, the utility SHALL prevent a pointer jump caused solely by their new contact positions.

### RF-7: End a drag

- WHEN the grace period expires, the utility SHALL release the drag action.
- WHILE the grace period is active, WHEN an ending tap occurs, the utility SHALL release the drag action immediately.
- WHEN an ending tap occurs, the utility SHALL consume the tap without generating an additional click.
- WHEN the utility releases the drag action, the utility SHALL end the active drag.

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
| All fingers lift during a drag | The grace period starts. |
| Contact count changes from three to one, two, or more than three | The grace period starts. |
| Exactly three fingers resume movement before expiry | The same drag continues without an intervening release. |
| Three fingers return but remain stationary | Timer behavior remains open. |
| The grace period expires | The drag ends. Later movement must satisfy RF-2 to start another drag. |
| The user taps during the grace period | The drag ends immediately, without an additional click. |
| The pointer starts over an area without drag support | The application's ordinary behavior applies. |
| A file drops over an invalid destination | The application determines the result. |
| A failure prevents immediate release | The active drag ends. The utility releases the held drag action at the earliest opportunity. |
| Movement coincides with grace-period expiry | Event precedence remains open. |
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

## [OPEN QUESTIONS]

- **OQ-1:** What is the grace-period duration? Can users change it?
- **OQ-2:** What finger count, duration, and movement tolerance define an ending tap?
- **OQ-3:** What movement threshold starts a drag? Must all three fingers move? How should staggered contacts behave?
- **OQ-4:** Do three stationary fingers affect the grace timer? Can further contact changes restart it? What happens exactly at expiry?
- **OQ-5:** After a failure, must users explicitly enable the feature again? Can availability recover automatically without resuming the interrupted drag?
- **OQ-6:** How do users enable or disable the feature? What happens when they disable it during a drag?
- **OQ-7:** What happens when users press a physical button or use another pointing device during a drag?
- **OQ-8:** Which touchpads, applications, and macOS reference behavior define the acceptance baseline?
- **OQ-9:** What response-time limits and pointer-motion criteria define acceptable behavior?
