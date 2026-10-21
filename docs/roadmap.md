# Community roadmap: next-level SML

SML is an open-source declarative terminal view engine for Go. The foundations (cell-based Unicode layout, safe parsing, components, and host-owned actions) already exist. **Everything linked below is a proposed improvement, not a shipped feature.** Follow the GitHub issue status for the source of truth.

This backlog was created on 21 October 2026 to make it straightforward for contributors to pick up self-contained work.

## Choose your first contribution

- [Good first issues](https://github.com/AmpedWasTaken/SML/issues?q=is%3Aissue%20is%3Aopen%20label%3A%22good%20first%20issue%22): smaller, bounded tasks appropriate for a first PR.
- [Help wanted](https://github.com/AmpedWasTaken/SML/issues?q=is%3Aissue%20is%3Aopen%20label%3A%22help%20wanted%22): tasks where design review or multiple-file changes are useful.
- [All open issues](https://github.com/AmpedWasTaken/SML/issues?q=is%3Aissue%20is%3Aopen): check current status; some issues may already be claimed or completed.

Issue bodies describe the problem, starting files, acceptance criteria, and recommended tests. `S` = focused small PR, `M` = multiple touched files, `L` = coordinate the public contract before coding. `P0` = correctness/compatibility, `P1` = valuable next milestone, `P2` = exploratory or longer-term.

**Recommended first foundation work:** #13 (source diagnostics), #15 (render snapshots), #17 (external API compatibility), and #25 (correct hitbox geometry). These support safer implementation of the more ambitious features.

## Contribution rules

1. Read [CONTRIBUTING.md](../CONTRIBUTING.md), [project structure](project-structure.md) and [security boundaries](security.md).
2. Comment on the issue with a brief plan to coordinate or avoid overlapping PRs; assignment is not required to contribute.
3. Branch from the latest `main`; keep public imports of `github.com/AmpedWasTaken/SML` and legacy view behavior stable.
4. Add tests and documentation before requesting review; verify `go test ./...` and `go vet ./...`.
5. For public syntax/API designs (particularly `L` tasks), agree on a limited interface in the issue before adding code.

SML intentionally **does not** run application commands, authenticate users, own terminal event loops or fetch remote resources. Those remain responsibilities of embedding hosts.

## Backlog by area (50 issues)

### Parser and developer tools

- [#11](https://github.com/AmpedWasTaken/SML/issues/11) — Safe `sml fmt` command
- [#12](https://github.com/AmpedWasTaken/SML/issues/12) — `sml lint` warnings
- [#13](https://github.com/AmpedWasTaken/SML/issues/13) — Structured source-position diagnostics
- [#14](https://github.com/AmpedWasTaken/SML/issues/14) — Typos and suggested names
- [#53](https://github.com/AmpedWasTaken/SML/issues/53) — Machine-readable grammar specification
- [#54](https://github.com/AmpedWasTaken/SML/issues/54) — Language server diagnostics/completion
- [#55](https://github.com/AmpedWasTaken/SML/issues/55) — Tree-sitter grammar
- [#58](https://github.com/AmpedWasTaken/SML/issues/58) — `sml inspect` frame JSON

### Tests, performance and shipping

- [#15](https://github.com/AmpedWasTaken/SML/issues/15) — Golden rendering/geometry snapshots
- [#16](https://github.com/AmpedWasTaken/SML/issues/16) — Parser fuzz corpus
- [#17](https://github.com/AmpedWasTaken/SML/issues/17) — External public API fixtures
- [#19](https://github.com/AmpedWasTaken/SML/issues/19) — Reproducible benchmark suite
- [#20](https://github.com/AmpedWasTaken/SML/issues/20) — Reduce render allocations
- [#21](https://github.com/AmpedWasTaken/SML/issues/21) — Cache immutable component templates
- [#59](https://github.com/AmpedWasTaken/SML/issues/59) — CI on Windows/macOS/Linux
- [#60](https://github.com/AmpedWasTaken/SML/issues/60) — Release validation and compatibility gate

### Layout and sizing

- [#22](https://github.com/AmpedWasTaken/SML/issues/22) — Min/max width constraints
- [#23](https://github.com/AmpedWasTaken/SML/issues/23) — Row/column gap
- [#24](https://github.com/AmpedWasTaken/SML/issues/24) — Horizontal distribution
- [#25](https://github.com/AmpedWasTaken/SML/issues/25) — Exact panel padding/border geometry
- [#26](https://github.com/AmpedWasTaken/SML/issues/26) — Viewport clipping and scroll hitboxes

### Tables, data and reusable components

- [#27](https://github.com/AmpedWasTaken/SML/issues/27) — Sticky table header metadata
- [#28](https://github.com/AmpedWasTaken/SML/issues/28) — Host-controlled table pagination
- [#29](https://github.com/AmpedWasTaken/SML/issues/29) — Column alignment and empty fallbacks
- [#30](https://github.com/AmpedWasTaken/SML/issues/30) — Typed component props
- [#31](https://github.com/AmpedWasTaken/SML/issues/31) — Optional slot fallback
- [#33](https://github.com/AmpedWasTaken/SML/issues/33) — Keyed repeat controls
- [#52](https://github.com/AmpedWasTaken/SML/issues/52) — Expandable tree view

### Language and styling

- [#32](https://github.com/AmpedWasTaken/SML/issues/32) — Numeric derive formatting
- [#34](https://github.com/AmpedWasTaken/SML/issues/34) — Semantic theme tokens
- [#35](https://github.com/AmpedWasTaken/SML/issues/35) — Terminal color capability
- [#36](https://github.com/AmpedWasTaken/SML/issues/36) — Style precedence rules

### Interaction and forms

- [#37](https://github.com/AmpedWasTaken/SML/issues/37) — Focus-order metadata
- [#38](https://github.com/AmpedWasTaken/SML/issues/38) — Public frame hit-test helper
- [#39](https://github.com/AmpedWasTaken/SML/issues/39) — Unicode action labels
- [#40](https://github.com/AmpedWasTaken/SML/issues/40) — Input constraints/error metadata
- [#41](https://github.com/AmpedWasTaken/SML/issues/41) — Controlled input change descriptors
- [#42](https://github.com/AmpedWasTaken/SML/issues/42) — Typed host action registry

### Accessibility and inclusivity

- [#43](https://github.com/AmpedWasTaken/SML/issues/43) — Semantic accessibility tree
- [#44](https://github.com/AmpedWasTaken/SML/issues/44) — Disabled/read-only reason metadata
- [#45](https://github.com/AmpedWasTaken/SML/issues/45) — Shortcut hints
- [#46](https://github.com/AmpedWasTaken/SML/issues/46) — Accessible high-contrast recipes
- [#47](https://github.com/AmpedWasTaken/SML/issues/47) — Reduced-motion/no-color preferences

### Charts and rich visuals

- [#48](https://github.com/AmpedWasTaken/SML/issues/48) — Bounded bar chart
- [#49](https://github.com/AmpedWasTaken/SML/issues/49) — Sparse sparkline robustness
- [#50](https://github.com/AmpedWasTaken/SML/issues/50) — Indeterminate progress
- [#51](https://github.com/AmpedWasTaken/SML/issues/51) — Text fallback for local media

### Reference integrations and learning

- [#56](https://github.com/AmpedWasTaken/SML/issues/56) — Bubble Tea integration example
- [#57](https://github.com/AmpedWasTaken/SML/issues/57) — SSH host (Wish) integration example
- [#61](https://github.com/AmpedWasTaken/SML/issues/61) — Ten-recipe SML cookbook

## How we decide what ships

Features should be general-purpose for Go terminal applications, bounded in resource usage, safe with untrusted markup, and tested for Unicode/cell geometry. New language constructs require a backwards-compatibility story. Priorities describe relative direction rather than release promises.

For changes that affect a public API, discuss typed data and versioning before implementation. Do not silently put Sentinel-specific commands or authorization logic back into the core package.

This page indexes issues, not releases: see [CHANGELOG.md](../CHANGELOG.md) and [releasing.md](releasing.md) for what has actually shipped.
