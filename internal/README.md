# Production package boundaries

`gesture` owns platform-independent contact, policy, state, and effect values.
It uses supplied monotonic time and physical millimeters. It does not access clocks, devices, configuration files, or desktops.

Future adapters supply complete frames for the selected touchpad. They must verify the device capabilities required by RF-1.
Synchronization loss produces unreliable input rather than inferred contacts. The runtime will handle termination and cleanup separately.

Output state records pending cleanup independently of gesture lifetime. Effects request logical primary-button actions and relative millimeter movement.
The output adapter will map those values to the approved desktop behavior. Experimental pointer gain is not a production motion policy.

`Engine` implements recognition transitions with supplied time, generation-tagged grace, and ordered abstract effects.
The default grace is 800 ms. Returning three contacts resumes immediately, and a fresh single contact requests release.
The runtime must confirm cleanup before the engine restores ordinary contacts.
The runtime must validate session identity for cleanup confirmations and apply termination events before movement at a shared timestamp.
Use `Process` to submit simultaneous events as one batch. It prioritizes termination and grace expiry before movement.
Cleanup confirmations carry the generation from `ReleasePrimary`. Old confirmations cannot clear a later cleanup obligation.
Ordinary-input restoration includes an owned contact snapshot after suppression. Physical-button consumption remains a separate routing effect.
Motion effects include monotonic timestamps and elapsed time with fractional millimeter deltas.
Adapters must map these values through the approved desktop response. The core does not apply experimental gain or acceleration twice.
Adapters and production output integration remain separate tasks. Movement effects carry millimeters, not the probe's experimental gain.
