# Conservative binding-driven invalidation

`View.BindingDependencies()` is a pure, opt-in descriptor computed from parsed SML markup. A host can use the resulting `BindingDependencies.AffectedBy(changedKeys...)` to decide if **binding-driven** updates require another render. No event loop, goroutine, action dispatch or terminal I/O is added to SML.

```go
usage := view.BindingDependencies()
if usage.AffectedBy("host.status") {
    // Host: render another Frame, then optionally use sml.DiffFrames.
}
```

## Dependency semantics

- `Keys` contains sorted host binding keys referenced by interpolation, field/input/session controls, conditional gates, table/repeat sources, prompts and direct control state gates.
- Derivations are traced transitively. A view using `derived.health` sourced from `host.status` depends on `host.status`; an unused derivation does not force a render.
- Both branches of a conditional count: state changes can reveal a previously hidden control.
- `Conservative=true` means a complete dependency graph is unavailable (for example a local `include`, a reusable `use` component or an unrecognized programmatically constructed node). With any changed key, `AffectedBy` then returns true.
- With no changed keys, `AffectedBy()` returns false even in conservative mode.
- The descriptor is derived from the current parsed `View`. Recalculate it after hot reload, asset changes, or swapping to a new view.

These static dependencies are intentionally an **over-approximation**. They may request renders that ultimately paint identical cells, but must not skip known host-bound changes. Avoid using them as a security or authorization signal.

## What this does not cover

Binding-driven invalidation is not automatic rendering or incremental layout. Regardless of `AffectedBy`, a host must independently redraw when terminal size, scroll position, focus presentation, theme, animation frame, asset content, app navigation or terminal capability changes.

No action data is executed or authorized by dependency metadata. After any redraw, the host must still validate the current user's permissions for every action descriptor. For a host-owned state batch, pass its `ChangedKeys` slice to `AffectedBy` once the typed-state API is available.

Use [`sml.DiffFrames`](frame-diffs.md) to compute changed terminal cells **after** a necessary render. A binding usage descriptor need not be persisted and can be discarded on unmount.
