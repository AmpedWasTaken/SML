# Typed bindings and batched presentation state

Typed bindings are an **opt-in Go API**, not a new SML markup version or an expression interpreter. Existing `map[string]string`, `RenderFrame` and `RenderInteractive` are unchanged.

## Scalar values

Use `sml.StringValue`, `sml.IntValue`, `sml.FloatValue`, `sml.BoolValue` and `sml.NullValue`. A `sml.Value` is comparable, has a `Kind()` and typed accessors (`Text`, `Int64`, `Float64`, `Bool`). Only finite float64 numbers are valid; string values must be UTF-8 and at most 4096 bytes.

List and record bindings, timestamps and duration types are not implemented in this increment. Those require a reviewed bounded structured-value/data-provider contract, rather than serializing arbitrarily large datasets into view strings.

## Host-managed state

```go
store, err := sml.NewBindingState(map[string]sml.Value{
    "metrics.count": sml.IntValue(1),
    "metrics.online": sml.BoolValue(true),
})
if err != nil { return err }

first := store.Snapshot()
oldFrame, err := view.RenderTypedFrame(first.Values, 80)
if err != nil { return err }

change, err := store.Apply(
    sml.BindingUpdate{Key: "metrics.count", Value: sml.IntValue(2)},
    sml.BindingUpdate{Key: "metrics.online", Value: sml.BoolValue(false)},
)
if err != nil { return err }
// change.Revision increments once; change.ChangedKeys is sorted.
second := store.Snapshot()
newFrame, err := view.RenderTypedFrame(second.Values, 80)
if err != nil { return err }
diff, err := sml.DiffFrames(oldFrame, newFrame)
if err != nil { return err }
_ = diff // The host decides whether to repaint changed cells or everything.
```

Batch updates are all-or-nothing, even if an invalid key or value appears last. A no-op batch does not increment the revision. Deletion uses `BindingUpdate{Key: "metrics.count", Delete: true}` and changes the revision only if the key exists.

`Snapshot` returns a detached map. Concurrent `Apply` and `Snapshot` calls are serialized with a lock; the store starts **no goroutines**, does **no terminal I/O**, and never retains snapshots after returning them. Render from a snapshot owned by the render goroutine; do not concurrently mutate the returned map yourself.

## Limits, safety and compatibility

- At most 256 bindings or updates per batch, with a maximum of 128 KiB of aggregate key/value text, including 4096 bytes per string.
- Names follow the existing SML binding-name grammar. The `derived.*` namespace is protected: only the existing SML derivation layer may produce those keys.
- Input strings are validated for UTF-8; SML's existing interpolation sanitization removes terminal control characters before presentation. Neither values nor markup authorize actions.
- No error reports submitted values or secrets. Do not store passwords in a presentation binding; password inputs need dedicated host-owned safe handling.
- Updates produce changed-key metadata, **not** automatic render invalidation. For now, a host may rerender then call `DiffFrames`. Dependency tracking, scoped component state and selective recomputation remain follow-up work.
- The older string-map renderer is fully supported. No SML v4 syntax is required and no Sentinel-specific dependencies are introduced.

See [frame diffs](frame-diffs.md) and [Go API](go-api.md).
