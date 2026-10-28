# Frame diffs for Go hosts

`sml.DiffFrames(previous, current)` compares two already-rendered structural frames without writing ANSI, moving the cursor, running actions or creating goroutines. It is compatible with v1/v2/v3 markup and does not introduce a new versioned syntax.

```go
before, err := view.RenderFrame(oldBindings, 80)
if err != nil { return err }
after, err := view.RenderFrame(newBindings, 80)
if err != nil { return err }
diff, err := sml.DiffFrames(before, after)
if err != nil { return err }
if diff.FullRedraw {
    // Host: clear and repaint using after.Lines (or host style renderer).
} else {
    for _, run := range diff.Runs {
        // Host: draw all run.Cells from (run.Start, run.Row), including
        // blanks to erase old text and full wide-grapheme clusters.
        _ = run
    }
}
if diff.HotspotsChanged || diff.ElementsChanged {
    // Host: refresh cached input/focus/layout metadata.
}
```

## Semantics

- `FrameDiff.Width` and `Height` describe the **current** frame in display cells.
- A `CellRun` contains `Row`, `Start` (zero-based, inclusive) and a **copy** of the current `Cells` that replace that row's range. The ending column is `Start + len(Cells)`, exclusive. Blank cells must be cleared by the host.
- Changed graphemes are expanded to cover the entire old **and** new cluster, including continuation cells. Do not write continuation cells as separate Unicode characters.
- Identical rendered cells produce no runs. Changed actions/hotspots or element geometry are signaled separately without forcing a text redraw.
- A missing/invalid previous frame or changed width/height requests `FullRedraw=true`. The host should clear prior dimensions before repainting after resize, especially if the old screen was larger.
- `Frame.Cells` currently has no structured foreground/background/style information. When raw `Lines` differ in a way not represented in the cell model (for example ANSI styling), `FullRedraw=true` prevents silently losing the style update. Hosts must repaint from the current styled lines rather than reconstructing style from `Cell`.
- The API validates bounded frame shapes (width 1..240, height 0..3000) and returns errors for malformed current frames. It does not mutate either input.
- The host owns synchronized terminal output, terminal capabilities, cursor, focus, overlays, cancellation and authorization. A diff is **not** itself a terminal renderer.
- Frame comparison remains linear in frame size. This API avoids large output writes when little changes, but is not automatic dependency tracking or lazy virtualization.

## Performance and compatibility

For a repeatable local microbenchmark:

```sh
go test ./internal/engine -run '^$' -bench BenchmarkSmallFrameDiff -benchmem
```

Report the Go version, CPU, platform, dimensions, allocated bytes and ns/op; compare to full renders instead of claiming frame diffing is always faster. A full snapshot and a small 120x34 change are covered by regression tests. A real terminal benchmark remains a separate task.

For scrollable screens use `Frame.Window(...)` before comparing two same-sized viewport projections. All new APIs are additive and remain independent from Sentinel.
