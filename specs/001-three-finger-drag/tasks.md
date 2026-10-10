# Tasks: Three-Finger Drag

Intended repository location: `specs/001-three-finger-drag/tasks.md`.
Sources: `docs/constitution.md`, `specs/001-three-finger-drag/spec.md`, and `specs/001-three-finger-drag/plan.md`.

## Task rules

- Each task allows at most 30 minutes of active work, including its focused checks.
- Estimates exclude approval waits and test-machine provisioning. Obtain those prerequisites before starting dependent work.
- If work exceeds its estimate, mark it `partial` and split the remaining work before continuing.
- Rows follow dependency order. Independent rows may proceed separately.
- A dependency requires `done`. A declined task does not satisfy its dependents.
- `done` means the completion criterion has evidence.
- `partial` means work started but the completion criterion remains unmet.
- `declined` means an explicit decision rejected the task. Record the reason and revise affected dependencies.
- `-` means work has not started, including work that awaits a prerequisite.
- All rows start with `-`. Existing specification and plan documents do not prove these tasks complete.
- Record evidence with the task status. Never store captured input events, pointer paths, or runtime state.
- Record authored scenarios, environment versions, aggregate measurements, and pass/fail results instead.
- Future implementation tasks require separate authorization to write code.

For each future Go change, run `make fmt` and applicable tests before marking its task done.
Add regression tests with defect fixes. Use English for identifiers, comments, messages, errors, and documentation.
Update the README tree whenever a task adds files or directories. Run `graphify update .` after code changes.
If these checks exceed the task budget, split the work. Do not omit required checks.

## 1. Resolve the specification

These tasks capture user decisions. They do not select new behavior without approval.

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T01 | Define contact recognition and drag-start thresholds. | RF-2, RF-3 | 25 min | None | Made when: Approved criteria cover contact validity, staggered placement, stationary fingers, and required finger movement. Evidence: [T01 decisions](#t01-completion-evidence). | done |
| T02 | Define the ending tap. | RF-7 | 20 min | T01 | Made when: Approved criteria specify finger count, duration, movement tolerance, recognition time, and tap consumption. Evidence: [T02 decisions](#t02-completion-evidence). | done |
| T03 | Define grace timing and event precedence. | RF-5, RF-6, RF-7 | 25 min | T02 | Made when: Approved criteria cover duration, timer restarts, stationary returns, repeated changes, and events exactly at expiry. | - |
| T04 | Define motion and primary-button behavior. | RF-2, RF-3, RF-4, RF-6 | 25 min | T01 | Made when: Approved criteria define initial movement, direction, scaling, acceleration, repositioning, and left-handed button behavior. | - |
| T05 | Define failed-drag termination and pending release. | RF-8, RF-9 | 25 min | T03 | Made when: Approved terms distinguish terminated gestures from pending releases and define ordinary-input routing during cleanup. | - |
| T06 | Define enablement, recovery, and rearming. | RF-2, RF-7, RF-8, RF-9 | 25 min | T05 | Made when: Approved criteria cover startup, disablement, recovery, fresh contacts, and contacts remaining after expiry. | - |
| T07 | Define physical-button and other-device interactions. | RF-3, RF-7, RF-8, RF-9 | 25 min | T04, T06 | Made when: Approved criteria cover physical buttons, other pointers, multiple touchpads, duplicate instances, and explicit cancellation. | - |
| T08 | Define lifecycle and desktop interruption behavior. | RF-4, RF-8, RF-9 | 25 min | T05, T06 | Made when: Approved criteria cover exit, crash, lock, suspend, session loss, monitor changes, and target-window disappearance. | - |
| T09 | Select the compatibility and macOS baseline. | RF-1, RF-4, RF-8 | 25 min | None | Made when: The user approves device capabilities, versions, applications, XWayland scope, window states, and the macOS reference. | - |
| T10 | Define measurable responsiveness and comparison rules. | RF-2, RF-3, RF-6, RF-7, RF-9 | 25 min | T03, T04, T09 | Made when: Approved limits and measurement methods cover start, motion, continuation, release, and recoverable failure cleanup. | - |
| T11 | Define explicit configuration behavior. | RF-1, RF-5, RF-7, RF-9 | 20 min | T03, T06 | Made when: Approved settings have defaults, bounds, missing-file behavior, invalid-value behavior, and rules for changes during a drag. | - |
| T12 | Complete the specification consistency review. | RF-1 through RF-9 | 30 min | T01 through T11 | Made when: RF contradictions are resolved, release-blocking questions are closed, and completion criteria preserve the constitution. | - |

### T01 completion evidence

The user approved these decisions in the T01 clarification discussion:

- Use the average position of exactly three valid contacts. Individual fingers may remain stationary.
- Exclude detected palms and include thumbs.
- Allow staggered placement without a time limit. Establish the baseline when the third valid contact arrives.
- Start at a straight-line group displacement of at least 1 mm, configurable through the configuration file.
- Accept finite numeric thresholds greater than zero. Reject zero, negative, nonnumeric, and nonfinite values.
- Do not accumulate traveled distance or count movement before the baseline.
- Before drag start, reset the baseline when valid contact count or identities change.
- Three stationary contacts generate neither a click nor a drag.

Evidence: [RF-2 and RF-3](spec.md#rf-2-start-a-drag), [resolved OQ-3](spec.md#resolved-decisions), and [plan decisions](plan.md#approved-t01-decisions).
The [T01 acceptance scenarios](plan.md#t01-acceptance-scenarios) define future verification. T01 completes specification decisions, not implementation testing.

### T02 completion evidence

The user approved these decisions in the T02 clarification discussion:

- Require a fresh one-finger tap after all valid contacts lift. Include thumbs and ignore detected palms.
- Recognize the tap on finger lift within 200 ms, including exactly 200 ms.
- Keep displacement within 1 mm of the initial contact position throughout contact. Exactly 1 mm qualifies.
- Cancel the candidate when either limit is exceeded or a second valid contact appears.
- After cancellation, require all valid contacts to lift before another attempt.
- Configure duration and displacement independently from the drag-start threshold.
- Accept only finite numeric values greater than zero for both limits.
- Require touchdown and lift before grace expiry. Touchdown does not extend the grace period.
- Consume the recognized tap and release the drag without an additional click, as RF-7 requires.

Evidence: [RF-7 recognition](spec.md#ending-tap-recognition), [resolved OQ-2](spec.md#resolved-decisions), and [plan decisions](plan.md#approved-t02-decisions).
The [T02 acceptance scenarios](plan.md#t02-acceptance-scenarios) define future verification. T02 completes specification decisions, not implementation testing.
T03 retains the decision about finger lift exactly at grace expiry.

## 2. Establish backend feasibility

Prepare existing Arch Linux, CachyOS, and macOS test environments before these tasks.
Machine installation and hardware procurement are external prerequisites, not hidden 30-minute tasks.
Select one approved Linux session for initial experiments. Validate the other distribution during acceptance.

Experiments may use existing tools or isolated prototypes after code authorization.
A completed experiment records a pass or failure. Only passing evidence permits backend adoption at T29.
If an experiment needs more work, add a bounded follow-up task and update T29's dependencies.

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T13 | Check the prepared test environments. | RF-1, RF-4 | 20 min | T09 | Made when: Version inventories confirm access to both Linux sessions, selected devices, applications, and the macOS reference. | - |
| T14 | Define the native adapter API and dependency. | RF-1, RF-8, RF-9 | 25 min | T12 | Made when: The plan lists the required libevdev calls, version constraints, ownership rules, and dependency justification. | - |
| T15 | Establish a bounded experiment harness. | RF-1, RF-8, RF-9 | 30 min | T13, T14 | Made when: The harness can open and close one synthetic device without recording input or capturing unrelated devices. | - |
| T16 | Probe virtual-touchpad recognition. | RF-1, RF-8 | 25 min | T15 | Made when: Evidence records Hyprland recognition, seat assignment, capabilities, axis ranges, resolution, and applicable settings. | - |
| T17 | Probe virtual-pointer recognition. | RF-1, RF-3, RF-4 | 25 min | T15 | Made when: Evidence records pointer recognition, seat assignment, primary-button behavior, and relative-motion delivery. | - |
| T18 | Probe exclusive physical-input ownership. | RF-1, RF-8 | 25 min | T16 | Made when: Evidence records grab acquisition, duplicate-event suppression, virtual-device exclusion, and ownership release. | - |
| T19 | Compare ordinary one-finger pointing. | RF-8 | 20 min | T18 | Made when: A repeatable comparison records pointing behavior with and without the proxy under identical settings. | - |
| T20 | Compare ordinary two-finger scrolling. | RF-8 | 20 min | T18 | Made when: A repeatable comparison records scrolling behavior with and without the proxy under identical settings. | - |
| T21 | Probe one-contact to three-contact ownership transfer. | RF-2, RF-8 | 30 min | T17, T19 | Made when: Evidence records whether transfer causes an extra click, pointer jump, stale contact, or competing action. | - |
| T22 | Probe two-contact to three-contact ownership transfer. | RF-2, RF-8 | 30 min | T17, T20 | Made when: Evidence records whether transfer leaves scrolling, stale contacts, or competing gestures active. | - |
| T23 | Probe ending-tap suppression. | RF-7, RF-8 | 25 min | T21, T22 | Made when: Evidence records drag release and whether the desktop receives an additional click. | - |
| T24 | Probe contact repositioning and return to ordinary input. | RF-5, RF-6, RF-8 | 30 min | T21, T22 | Made when: Evidence records continuation displacement and stale-contact behavior after release. | - |
| T25 | Probe drag-motion scaling. | RF-3, RF-6 | 25 min | T17, T24 | Made when: Measurements identify acceleration stages and show whether fractional motion meets the approved mapping. | - |
| T26 | Probe process and virtual-device teardown. | RF-7, RF-9 | 25 min | T17, T18 | Made when: Evidence records button state and ordinary-input restoration after graceful exit and forced termination. | - |
| T27 | Probe desktop-session loss and restoration. | RF-9 | 25 min | T26 | Made when: Evidence records cleanup behavior and whether stale movement or releases cross session boundaries. | - |
| T28 | Probe contact classification on baseline hardware. | RF-1, RF-2, RF-5, RF-8 | 25 min | T18 | Made when: Evidence identifies palm, thumb, noise, and partial-multitouch behavior against the approved device baseline. | - |
| T29 | Approve or reject the provisional backend. | RF-1 through RF-9 | 25 min | T19 through T28 | Made when: All eight plan feasibility conditions pass and the plan records backend adoption with evidence links. | - |

If the backend fails, mark T29 `partial` or `declined` with the reason.
Revise the plan and task dependencies before production implementation. Failed experiments do not establish feasibility.

## 3. Build the isolated gesture logic

Each behavior task includes focused deterministic unit tests.
Use synthetic contacts and an injected clock. Do not use real sleeps or captured user-input fixtures.

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T30 | Establish package boundaries and formatting checks. | RF-1 through RF-9 | 25 min | T29 | Made when: The module builds, `make fmt` works, and gesture logic has no platform, configuration, or interface dependencies. | - |
| T31 | Define contact, state, policy, and effect records. | RF-2 through RF-9 | 25 min | T30 | Made when: Records match the approved model and distinguish gesture termination from pending output release. | - |
| T32 | Add the synthetic frame and clock test harness. | RF-2 through RF-9 | 25 min | T31 | Made when: A test advances time explicitly and checks ordered effects without hardware, real sleeps, or persisted events. | - |
| T33 | Implement the approved contact-validity rule. | RF-1, RF-2, RF-5 | 25 min | T32 | Made when: Tests cover accepted fingers, rejected contacts, noise, and approved hardware limitations. | - |
| T34 | Implement drag-start recognition. | RF-2, RF-4 | 30 min | T33 | Made when: Boundary tests prove three-contact movement starts one hold and stationary or other-count contacts do not. | - |
| T35 | Implement relative drag movement. | RF-3, RF-4 | 25 min | T34 | Made when: Tests verify initial motion, approved scaling, fractional movement, and no repeated button press. | - |
| T36 | Implement contact-count pause behavior. | RF-5 | 25 min | T35 | Made when: Tests cover departures to zero, one, two, and additional contacts with no gesture-driven movement. | - |
| T37 | Implement grace-period timer rules. | RF-5, RF-6, RF-7 | 30 min | T36 | Made when: Tests verify approved duration, stationary returns, repeated changes, and stale-timer rejection. | - |
| T38 | Implement drag continuation and new baselines. | RF-6 | 25 min | T37 | Made when: Tests prove continuation retains the same hold and excludes repositioning displacement. | - |
| T39 | Implement ending-tap recognition. | RF-7 | 25 min | T37 | Made when: Tests cover accepted taps and duration, movement, and contact-count rejection boundaries. | - |
| T40 | Implement release and contact rearming. | RF-2, RF-7, RF-8 | 25 min | T38, T39 | Made when: Tests verify one release, no extra tap click, and approved behavior for contacts remaining after release. | - |
| T41 | Implement exact-boundary event precedence. | RF-3, RF-5, RF-6, RF-7 | 25 min | T40 | Made when: Tests cover events before, at, and after expiry and apply the approved ordering consistently. | - |
| T42 | Implement gesture termination after failure. | RF-8, RF-9 | 25 min | T41 | Made when: Tests prove failure cancels motion and timers while preserving only the required cleanup obligation. | - |
| T43 | Check gesture invariants with synthetic sequences. | RF-2 through RF-9 | 25 min | T42 | Made when: Generated sequences verify balanced button effects, stale-event rejection, and no continuation after failure. | - |

## 4. Build device adapters and runtime coordination

Reuse only experiment code that satisfies the approved adapter contract and production checks.

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T44 | Implement device selection and ownership lifecycle. | RF-1, RF-8, RF-9 | 30 min | T30 | Made when: Adapter tests verify selection, grab/release, native-resource cleanup, and exclusion of virtual devices. | - |
| T45 | Assemble complete contact frames. | RF-2, RF-3, RF-5, RF-7 | 30 min | T31, T44 | Made when: Synthetic adapter tests verify slots, contact IDs, complete-frame boundaries, and approved contact classification. | - |
| T46 | Handle lost synchronization and device removal. | RF-1, RF-9 | 25 min | T45 | Made when: Injected discontinuities produce the approved resynchronization or failure outcome without stale contacts. | - |
| T47 | Implement virtual-touchpad lifecycle. | RF-1, RF-8, RF-9 | 30 min | T44 | Made when: Adapter tests verify capabilities, settings, creation, delayed recognition, and destruction. | - |
| T48 | Implement ordinary-input forwarding. | RF-8 | 25 min | T45, T47 | Made when: Adapter tests preserve complete ordinary-input frames without duplicate physical input or feedback loops. | - |
| T49 | Implement virtual-pointer output. | RF-2, RF-3, RF-4, RF-7 | 30 min | T44 | Made when: Tests verify ordered hold, relative motion, release, and approved button mapping on the selected seat. | - |
| T50 | Track output submission and release failures. | RF-7, RF-9 | 25 min | T49 | Made when: Fault injection distinguishes submitted output, uncertain button state, and pending release without claiming application receipt. | - |
| T51 | Implement gesture ownership transfer. | RF-2, RF-7, RF-8 | 30 min | T43, T48, T50 | Made when: Integration checks match the verified transfer procedure without extra clicks, stale contacts, or competing gestures. | - |
| T52 | Implement contact re-entry after a drag. | RF-6, RF-7, RF-8 | 25 min | T51 | Made when: Checks verify continued dragging and return to ordinary input without pointer jumps or stale contacts. | - |
| T53 | Implement Hyprland readiness checks. | RF-1, RF-8, RF-9 | 25 min | T47, T49 | Made when: Tests distinguish supported sessions, missing devices, wrong seats, and unavailable output. | - |
| T54 | Implement desktop-session identity and loss detection. | RF-9 | 25 min | T53 | Made when: Tests detect session loss and reject output effects associated with an obsolete session. | - |
| T55 | Implement monotonic event-time mapping. | RF-2, RF-3, RF-5, RF-6, RF-7 | 25 min | T45, T41 | Made when: Tests compare frame times and timer deadlines in one clock domain without wall-clock dependence. | - |
| T56 | Implement the ordered runtime event loop. | RF-2 through RF-9 | 30 min | T43, T46, T50, T52, T54, T55 | Made when: Integration checks preserve approved event precedence and effect order across frames, timers, failures, and output results. | - |
| T57 | Implement failure cleanup and approved recovery. | RF-8, RF-9 | 30 min | T56 | Made when: Fault tests discard queued motion, retry permitted releases, and prevent interrupted gestures from resuming. | - |
| T58 | Implement approved lifecycle and external-input rules. | RF-3, RF-7, RF-8, RF-9 | 30 min | T57 | Made when: Focused tests match approved cancellation, physical-button, other-device, shutdown, and session-interruption rules. | - |

If T58 contains multiple independent handlers, split it before implementation using the decisions from T07 and T08.
The 30-minute limit applies to each resulting task, including its focused tests.

## 5. Add configuration and the command interface

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T59 | Implement XDG configuration lookup. | RF-1, RF-9 | 20 min | T30 | Made when: Tests verify absolute XDG paths, fallback paths, missing files, and no file creation during ordinary startup. | - |
| T60 | Implement strict configuration validation. | RF-1, RF-5, RF-7, RF-9 | 25 min | T59 | Made when: Tests cover approved defaults, bounds, unknown fields, unsupported versions, and invalid-configuration behavior. | - |
| T61 | Implement atomic configuration replacement. | RF-1, RF-9 | 30 min | T60 | Made when: Tests verify same-directory temporary writes, file flush, close, rename, directory flush, and temporary-file cleanup. | - |
| T62 | Test configuration-save failure boundaries. | RF-1, RF-9 | 25 min | T61 | Made when: Fault tests preserve complete files and distinguish failures before replacement from durability failures after replacement. | - |
| T63 | Implement the approved command interface. | RF-1, RF-8, RF-9 | 25 min | T60 | Made when: Interface tests verify approved commands, invalid arguments, startup state, and English error output. | - |
| T64 | Connect startup, enablement, and disablement. | RF-1, RF-8, RF-9 | 30 min | T58, T62, T63 | Made when: Integration tests cover resource acquisition, readiness, disablement during a drag, and rollback after startup failure. | - |
| T65 | Audit configuration and runtime persistence. | RF-1 through RF-9 | 20 min | T64 | Made when: Checks show only explicit configuration persists and runtime paths store no input events or gesture state. | - |
| T66 | Audit language, errors, and package boundaries. | RF-1 through RF-9 | 20 min | T64 | Made when: Review confirms English text, operation-and-cause errors, justified dependencies, and isolated gesture logic. | - |

T59 through T62 also cover NFR-6. T65 covers NFR-5. T66 covers NFR-7 and NFR-8.

## 6. Verify the integrated application

These tasks use the prepared Linux test session and authored synthetic scenarios.
For failures, record a separate bounded fix task with a regression test before repeating the affected check.

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T67 | Verify start, motion, and button ordering. | RF-2, RF-3, RF-4 | 25 min | T64 | Made when: Desktop-observed synthetic scenarios confirm correct start position, one hold, approved movement, and no physical press requirement. | - |
| T68 | Verify pauses and continuation. | RF-5, RF-6 | 25 min | T67 | Made when: Scenarios cover every approved contact-count departure, timer rule, and repositioned continuation without an intervening release. | - |
| T69 | Verify tap, expiry, and boundary release. | RF-7 | 25 min | T68 | Made when: Desktop observations confirm approved expiry precedence, one release, and no extra click from the ending tap. | - |
| T70 | Verify ordinary gestures and ownership transfer. | RF-8 | 25 min | T69 | Made when: Pointing and scrolling match the baseline and transfer scenarios produce no competing three-finger action. | - |
| T71 | Verify input-access and device failures. | RF-9 | 25 min | T70 | Made when: Disconnection and denied or revoked access terminate the gesture and meet the approved cleanup policy. | - |
| T72 | Verify output failures and delayed release. | RF-7, RF-9 | 25 min | T71 | Made when: Injected write failures preserve cleanup obligations and release at the approved opportunity without stale motion. | - |
| T73 | Verify process and session interruptions. | RF-8, RF-9 | 25 min | T72 | Made when: Exit, forced termination, and approved session interruptions meet cleanup rules without cross-session effects. | - |
| T74 | Verify external actions and desktop edge cases. | RF-3, RF-4, RF-7, RF-8, RF-9 | 25 min | T73 | Made when: Approved physical-button, other-pointer, cancellation, target-disappearance, and monitor-change scenarios meet their acceptance criteria. | - |
| T75 | Run final automated checks. | RF-1 through RF-9 | 25 min | T65, T66, T74 | Made when: Formatting, unit, race, and integration checks pass with no required scenario counted as passed after a skip. | - |

T75 uses `make fmt`, `go test ./...`, `go test -race ./...`, and `go test -tags=integration ./...`.
NFR-3 and NFR-4 also apply to all implementation and verification tasks.

## 7. Verify both distributions and the macOS comparison

Run each row against the baseline selected in T09.
If that baseline has multiple hardware or application combinations, expand the row into separate tasks before execution.
Each expanded task covers one bounded scenario group on one combination, with at most 30 minutes of work.
The final evidence review depends on every expanded task.

| Task | Description | RF coverage | Time | Depends on | Completion criteria | Status |
| --- | --- | --- | --- | --- | --- | --- |
| T76 | Verify Arch Linux file dragging. | RF-1, RF-2, RF-3, RF-4, RF-7 | 20 min | T75 | Made when: Valid and invalid destination scenarios match application behavior without a physical press or extra click. | - |
| T77 | Verify Arch Linux text selection. | RF-1, RF-2, RF-3, RF-4, RF-7 | 20 min | T75 | Made when: Approved selection scenarios produce the expected text range and release behavior. | - |
| T78 | Verify Arch Linux title-bar movement. | RF-1, RF-2, RF-3, RF-4, RF-7 | 20 min | T75 | Made when: Each approved window-state scenario meets its movement and release criteria. | - |
| T79 | Verify Arch Linux gesture transitions. | RF-1, RF-5, RF-6, RF-7, RF-8 | 25 min | T75 | Made when: Grace, repositioning, ending taps, ordinary gestures, and competing-action checks pass on baseline hardware. | - |
| T80 | Verify Arch Linux failures. | RF-1, RF-8, RF-9 | 25 min | T75 | Made when: Approved access, device, output, and lifecycle failure scenarios meet termination and release criteria. | - |
| T81 | Verify CachyOS file dragging. | RF-1, RF-2, RF-3, RF-4, RF-7 | 20 min | T75 | Made when: Valid and invalid destination scenarios match application behavior without a physical press or extra click. | - |
| T82 | Verify CachyOS text selection. | RF-1, RF-2, RF-3, RF-4, RF-7 | 20 min | T75 | Made when: Approved selection scenarios produce the expected text range and release behavior. | - |
| T83 | Verify CachyOS title-bar movement. | RF-1, RF-2, RF-3, RF-4, RF-7 | 20 min | T75 | Made when: Each approved window-state scenario meets its movement and release criteria. | - |
| T84 | Verify CachyOS gesture transitions. | RF-1, RF-5, RF-6, RF-7, RF-8 | 25 min | T75 | Made when: Grace, repositioning, ending taps, ordinary gestures, and competing-action checks pass on baseline hardware. | - |
| T85 | Verify CachyOS failures. | RF-1, RF-8, RF-9 | 25 min | T75 | Made when: Approved access, device, output, and lifecycle failure scenarios meet termination and release criteria. | - |
| T86 | Measure Arch Linux response times. | RF-1, RF-2, RF-3, RF-6, RF-7, RF-9 | 25 min | T76 through T80 | Made when: Aggregate start, movement, continuation, release, and cleanup timings satisfy the approved limits. | - |
| T87 | Measure CachyOS response times. | RF-1, RF-2, RF-3, RF-6, RF-7, RF-9 | 25 min | T81 through T85 | Made when: Aggregate start, movement, continuation, release, and cleanup timings satisfy the approved limits. | - |
| T88 | Run the macOS reference scenarios. | RF-2 through RF-8 | 25 min | T12, T13 | Made when: Authored scenario results and aggregate timings document the approved macOS baseline without input recordings. | - |
| T89 | Review equivalence against macOS. | RF-2 through RF-8 | 25 min | T86, T87, T88 | Made when: Comparisons satisfy the approved equivalence criteria and every remaining difference has an explicit disposition. | - |
| T90 | Align user documentation with the verified release. | RF-1, RF-8, RF-9 | 25 min | T89 | Made when: README requirements, configuration, commands, limitations, and project tree match verified behavior and approved scope. | - |
| T91 | Complete the RF and constitution evidence review. | RF-1 through RF-9 | 25 min | T65, T66, T75, T76 through T90 | Made when: Every RF criterion has passing evidence on both distributions and all constitutional obligations have evidence. | - |

T86 and T87 cover NFR-2. T88 and T89 cover NFR-1.
T91 also verifies every approved expansion and defect-fix task. Missing evidence prevents completion.

## RF traceability

| Requirement | Core implementation | Integrated checks | Distribution acceptance |
| --- | --- | --- | --- |
| RF-1 | T44, T47, T53, T59, T60, T63, T64 | T75 | T76 through T87 |
| RF-2 | T33, T34, T40, T49, T51 | T67 | T76 through T78, T81 through T83 |
| RF-3 | T35, T49, T55, T56 | T67, T74 | T76 through T78, T81 through T83, T86, T87 |
| RF-4 | T34, T35, T49 | T67, T74 | T76 through T78, T81 through T83 |
| RF-5 | T36, T37 | T68 | T79, T84 |
| RF-6 | T38, T52 | T68 | T79, T84, T86, T87 |
| RF-7 | T39, T40, T41, T49, T50 | T69, T72 | T76 through T79, T81 through T84 |
| RF-8 | T48, T51, T52, T57, T58, T64 | T70, T73, T74 | T79, T80, T84, T85 |
| RF-9 | T42, T46, T50, T54, T57, T58, T64 | T71 through T74 | T80, T85, T86, T87 |
