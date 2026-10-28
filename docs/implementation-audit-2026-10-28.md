# SML next-generation implementation audit — 2026-10-28

> Source of truth: the current SML default branch, existing test files and open GitHub issues. This is an implementation baseline, **not** a claim that SML 4.0 has shipped or that any platform has passed a release gate.

## Existing capabilities versus target

| Area | Existing implementation | Status / gap | Relevant issues |
| --- | --- | --- | --- |
| Parse / validation | Strict XML-like markup, size/depth/node limits, v1–v3 documents, validation for controls and assets | Working foundation; diagnostics and authoring tools need work | #13, #14, #16 |
| Public Go API | Parse, ParseFile, LoadDir, View methods, RenderWidth, RenderChecked, RenderInteractive, RenderFrame; public type aliases | Working; external-consumer compile fixtures should expand | #17 |
| Layout / Unicode | Structural rows, columns, weighted grids, panels, grapheme cells (uniseg), responsive collapse | Working with gaps in exact geometry, sizing constraints and scroll windows | #22–#26 |
| Viewport and hit testing | Frame.Cells, Frame.Hotspots, cell positions | This branch introduces post-render Window and HitTest; **not** nested scrolling or virtual data providers | #26, #38 |
| Focus / keyboard | Stable-ish hotspot IDs, named control IDs and host action fields | Missing reusable focus order, groups, keyboard metadata and adapters | #37, #42, #45 |
| State / reactive bindings | String-map bindings, derive, conditional and repeated output, bounded state controls | No typed values or incremental dependency-invalidation public contract | #32, #33, #41 |
| Forms | Existing input/number/toggle/choice controls; host-owned state | No complete multiline editor, form validation or overlay focus lifecycle | #39–#41 |
| Data tables | Bounded JSON rows (200 max), column schema, sorting, filters, pages | Not a virtual DataGrid; needs host-provided windows and richer metadata | #27–#29 |
| Hierarchical navigation | Reusable markup components | Bounded TreeView not implemented | #52 |
| Media and charts | Local bounded image/GIF and simple gauges/sparklines | More text fallbacks, charts and animation controls pending | #48–#51 |
| Styling | Inline attributes, border themes and gradient profiles | Semantic tokens / capability adaptation incomplete | #34–#36, #47 |
| Tooling | CLI validate/render; docs and examples | Missing fmt, lint, inspect, preview runtime, LSP and editor grammar | #11–#12, #53–#58 |
| Tests and performance | Engine tests for Unicode, bindings, interactive controls, panels, grid and media | Need golden geometry snapshots, benchmarks and full multi-OS CI | #15, #19–#21, #59 |
| Sentinel integration | Reusable independent module; rendering/actions are declarative | Host-owned app navigation, auth, worker RPC and app lifecycle remain in Sentinel | Sentinel-V5 |

## Work sequence

1. **Current increment:** pure frame viewport slicing and hit testing; preserve existing Parse / Render methods and v1–v3 markup without new syntax.
2. Verify panel border/padding and focus geometry at 40/60/80/120 columns; add goldens and measure clipping costs.
3. Design focus/keyboard metadata and typed state without moving terminal event processing into SML.
4. Introduce bounded provider-based DataGrid and TreeView, then overlay composition.
5. Add semantic themes, developer CLI/inspector and reference applications before considering an SML 4 contract.

## Deliberate limits of this increment

- Window is a **post-render** projection of an existing Frame. Large datasets still need host-side pagination/virtual data supply, not a 1,000,000-line RenderFrame call.
- Window does not create nested viewport markup, scroll input handling, independent panes, cursor management or new SML language constructs.
- HitTest returns a typed Hotspot descriptor; invoking and authorizing it is strictly host responsibility.
- No terminal emulator was accessed by the audit. No benchmark figures or CI success are claimed until actually measured.
- Keep security, accessibility and the base declarative language available to all users.

## Verification gates

Run Go 1.23+ unit tests, vet, race tests, and external public API tests. Add Linux, Windows and macOS CI, followed by manual host integration tests for Windows Terminal, PuTTY and narrow SSH windows. See docs/layout.md for the implemented viewport contract. SML 4.0 remains untagged pending compatibility, security and API review.
