# Cell-based rendering and interaction geometry

SML provides two render paths for compatibility:

- `RenderChecked(values, width)` / `RenderWidth`: preserve v1-v3 string-rendering semantics.
- **`RenderFrame(values, width)`**: render a structural layout with Unicode grapheme-aware terminal cell measurement, text wrapping and exact node-derived interaction coordinates.

The frame API is currently opt-in so existing Sentinel-style views remain compatible. `RenderInteractive` now uses structural geometry internally.

```go
frame, err := view.RenderFrame(values, 80)
if err != nil { return err }
for _, line := range frame.Lines { fmt.Println(line) }
for _, hit := range frame.Hotspots {
    fmt.Printf("action=%s id=%s row=%d columns=[%d,%d)\n",
        hit.Action, hit.ID, hit.Line, hit.Start, hit.End)
}
```

## Data model

`Frame.Width` is the terminal width in cells; `Frame.Height` is rendered row count. `Frame.Cells[y][x]` contains a grapheme at the leading cell, with `Continuation=true` on the remaining cells of a full-width grapheme. Cells associated with an interactive control contain the corresponding `ElementID`. `Frame.Elements` exposes stable tree paths and rectangular layout bounds. `Frame.Hotspots` has exact display-column coordinates.

Layouts recursively compose children through columns, rows, grids, panels, sections, conditions and repeats. When a grid collapses for narrow terminals, its controls move with the actual layout rather than being found again by label text. Repeated controls receive distinct path-derived IDs. To give a control a user-owned stable ID, set `id="refresh"`.

## Unicode behavior

Terminal columns are calculated with `github.com/rivo/uniseg` using grapheme clusters: East Asian wide glyphs, combining marks, flag sequences and joined emoji do not split while clipping. `CellWidth`, `ClipCells`, `PadCells`, and `WrapText` are available for embedding applications.

The new renderer wraps `<text>` content on word boundaries when possible and breaks long words only between complete graphemes.

## Weighted grid panes

SML v3 structural grids support `weights="1,3,2"`, used as ratios rather
than fixed pixels. Each rendered row receives its complete available width,
with gaps accounted for and remainder allocated deterministically. Panes
cannot overlap; layouts collapse into one column when a 14-cell minimum
cannot be met. The single-column fallback still works with 12-cell frames.

Use this for compact monitors, left navigation + main editors, and a third
inspector pane. It doesn't implement adjustable dividers, independently
scrolling panes, overlays, ANSI colors, or an event loop. Those require
separate runtime capabilities in the host application.

## Scope and future work

`RenderFrame` is a **structural frame**, not a terminal input loop. Hosts still manage cursor movement, scrolling, key events, mouse support and authorization. The renderer currently targets up to 240 columns / 3,000 rows and does not implement arbitrary CSS flex properties, syntax-colored ANSI escape sequences inside bindings, bidirectional text layout, or clipping of complex multi-row controls. These remain separate engineering work.
