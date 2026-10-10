# Project Constitution

1. **Minimal stack:** Use Go and prefer its standard library; every external dependency or native binding must have a documented necessity.
2. **Specification first:** Define macOS-equivalent drag behavior and verifiable acceptance criteria in the active spec before implementation; update the spec before changing behavior.
3. **Isolated logic:** Keep gesture recognition and drag state transitions independent of device access, display-server backends, configuration, and user interfaces.
4. **Behavioral tests:** Cover gesture transitions and timing with deterministic unit tests, backend interactions with integration tests, and bug fixes with regression tests; run applicable tests before completion.
5. **Minimal persistence:** Persist only explicit configuration under the XDG configuration directory, using atomic writes; keep gesture state in memory and never persist input events.
6. **Consistent language:** Write identifiers, comments, documentation, messages, and errors in English; errors must identify the failed operation and cause, with actionable guidance where available.
