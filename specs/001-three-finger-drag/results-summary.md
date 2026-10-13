# Backend Feasibility Summary

## Scope

The user and assistant tested the libevdev proxy and virtual-pointer approach on CachyOS with Hyprland.
Manual desktop observations complement automated probe tests. They are not release acceptance for every supported environment.
Detailed local experiment records remain outside version control. This summary preserves the results and outstanding obligations.

## Environment

The tested session used kernel `7.2.9-2-cachyos`, Hyprland `0.56.2`, libinput `1.32.0`, and libevdev `1.14.0`.
Thunar `4.20.10` and Mousepad `0.7.0` launched under native Wayland and XWayland.
The physical touchpad advertised five contact slots, tracking IDs, fourth-contact detection, and 12 units/mm position resolution.
Arch Linux and the native macOS reference remain deferred.

## Observed results

| Tasks | Result |
| --- | --- |
| T13a | Local package, application-backend, and static device metadata collected. |
| T14 | Native adapter API, dependency baseline, and resource ownership documented. |
| T15–T18 | Synthetic devices appeared in Hyprland. Physical grabs suppressed motion and restored input. The probe rejected its own virtual touchpad. |
| T19–T20 | The user reported no apparent ordinary pointing or scrolling difference through the proxy. Numerical equivalence was not established. |
| T21–T22 | One-finger pointing and two-finger scrolling transitioned into three-finger selection without reported unwanted actions. Ordinary input returned afterward. |
| T23 | Revised fresh one-finger return released the hold and restored pointing. One run recorded 12 such endings and four continuations. |
| T24 | Repositioning continued the existing selection. One run recorded 22 continuations without a reported pointer jump. |
| T25 | Motion measurement tools are available. Manual trajectories differed, so numerical equivalence remains inconclusive. |
| T26 | The user confirmed active selection ended and normal pointing and scrolling returned after both Ctrl+C and SIGKILL. |
| T27 | Session-loss and cross-session cleanup testing remains deferred. |
| T28 | Controlled finger and thumb holds retained counts and identities. Valid mixed-palm scenarios reported three fingers and a palm together. |
| T29 | The user approved backend adoption for implementation, with T25 and T27 retained as release obligations. |

## Behavior changes approved during experiments

- Increase the default grace period from 300 ms to 500 ms.
- Resume the held drag as soon as exactly three valid contacts return before expiry.
- Reset the motion baseline on return. Require movement only for subsequent pointer motion, without another recognition threshold.
- After all valid contacts lift, a fresh one-finger return releases the drag immediately and restores ordinary pointing.
- Replace ending-tap recognition and remove its duration and displacement settings from the active policy.

Returning fingers first reported as one contact ends the drag, even if two more follow.
Specification and plan revision notices identify the current policy. Older tap-based scenarios require consolidation before production implementation.

## Measurement findings

Three controlled fingertip holds had maximum group displacements of 0.111, 0.169, and 0.056 mm.
Three thumb-plus-two-finger holds measured 0.274, 0.124, and 0.202 mm.
Each retained three non-palm contacts with no measured count or identity changes.
These measurements include involuntary movement and do not isolate sensor noise.

Two mixed-palm runs reported four non-palm contacts. The user identified incorrect palm placement in those runs.
Three corrected-placement runs retained three non-palm contacts and a palm in every measured frame.
Those results support the tested scenarios, not universal palm rejection.

The motion probe exposed a possible gain mismatch, especially for faster strokes.
Manual distance, direction, and speed were not sufficiently matched to establish its size or compliance with the specified tolerance.
An initial synchronization-loss attempt stopped before READY. A subsequent attempt completed, but the preparation-time cause remains unverified.

## Probe limitations

- Experimental output gain does not establish ordinary-touchpad acceleration equivalence.
- Logical primary-button mapping requires explicit selection in the probe. Left-handed behavior is not validated.
- The probe does not implement production compositor-confirmed pending cleanup, automatic recovery, or lock/suspend handling.
- Native integration tests require explicit device access. A skipped device test is not a passing hardware test.
- Measurements emit aggregates and derived stroke statistics. Raw input events, contact identities, and absolute pointer paths are not saved.

## Adoption and remaining work

T29 changes the implementation gate, not the release requirements. T30 may proceed.
Before release, complete T25, T27, Arch Linux acceptance, and native macOS comparison.
Resolve behavioral mismatches under the constitution before claiming equivalence.

The executable source and reproduction instructions are in [the probe runbook](../../experiments/inputprobe/README.md).
