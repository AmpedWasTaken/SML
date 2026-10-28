# Integrating SML into a Go app

## Parse and render an in-memory view

```go
view, err := sml.Parse([]byte(`<screen name="home"><text>Hello {{name}}</text></screen>`))
if err != nil { return err }
lines := view.RenderWidth(map[string]string{"name": "developer"}, 80)
for _, line := range lines { fmt.Println(line) }
```

`Parse([]byte)` validates the syntax, supported elements, styles, and declared controls. `Render(values)` uses the default width 80; `RenderWidth(values, width)` renders terminal text lines. Rendering uses the host's string bindings and never executes markup.

## Parse an on-disk SML view

```go
view, err := sml.ParseFile("views/home.sml")
if err != nil { return err }
```

`ParseFile` rejects symlink/non-regular input and files larger than the language's 64 KiB limit, then calls the same parser as `Parse`. For views with includes, call `view.WithAssets("assets")`, or prefer `LoadDir("assets")` to validate the complete asset tree.

## Load reusable project assets

```go
views, err := sml.LoadDir("assets")
if err != nil { return err }
view := views["home"]
lines := view.RenderWidth(bindings, 100)
```

`LoadDir` validates `assets/views/*.sml` and resolves component includes; it returns views only after the directory is successfully validated. If parsing a file directly, call `view.WithAssets("assets")` to enable local includes, art and media.

## Interactive metadata

```go
lines, hotspots, hasAnimation := view.RenderInteractive(bindings, 100)
_ = lines
_ = hotspots
_ = hasAnimation
```

A `Hotspot` exposes logical line and cell coordinates, a label, kind and declarative action fields (`Page`, `Command`, `PluginID`, `PluginAction`, etc.). The host must translate those coordinates to its own viewport and scroll offset. Disabled or hidden controls do not receive hotspots.

The host owns authorization, navigation, mouse support, keyboard focus, form state changes, resizing and redraw cadence. This API does not grant permission to execute a command or plugin action.

## Appearance profiles

Call `view.GradientStyles()` for an `appearance` view. The returned values are `terminal.GradientStyle` (colors, mode, direction, speed and FPS). Optional helpers in `github.com/AmpedWasTaken/SML/terminal` can apply the profile to text in a compatible terminal.

## Error handling

Invalid parse input returns an error. `LoadDir` returns errors for invalid views or includes. Use `view.RenderChecked(values, width)` when your host needs errors as Go `error` values. For compatibility, `RenderWidth` reports late render failures as a visible `SML render error:` line. The CLI uses `RenderChecked` and exits nonzero on render errors.

For production use, maintain an immutable last-known-good view snapshot and swap only when `LoadDir` succeeds.


## Structured frames and cells

For new applications, call `view.RenderFrame(bindings, width)`. Its `Frame` includes `Lines`, a two-dimensional `Cells` matrix, `Elements` with bounds, and `Hotspots` with stable IDs and host-owned actions. The frame measures positions in terminal cells, not Go runes or byte offsets. See [layout and interaction geometry](layout.md).

## Host-owned actions

Use `<button label="Refresh" id="refresh" action="app.refresh"/>` to emit an action ID without embedding application-specific commands into the library. The host checks the current user's permissions, maps `hotspot.Action` to an authorized handler, and performs any side effects. SML never dispatches the action.


## Scroll-aware structural frames (new optional API)

The Frame alias exposes two additional methods, without changing existing rendering or v1/v2/v3 SML syntax:

    full, err := view.RenderFrame(bindings, 80)
    if err != nil { return err }
    viewport, err := full.Window(0, scrollY, 60, 20)
    if err != nil { return err }
    if spot, ok := viewport.HitTest(localMouseX, localMouseY); ok {
        // Authenticate/authorize the action in the host before dispatch.
        _ = spot.Action
    }

Window clips by terminal grapheme cells; translated hotspots are only active where their painted cells remain visible. The viewport contains its own Lines, Cells, Elements and Hotspots. Width is 1..240, height 1..3000, and scroll offsets must be nonnegative. Invisible right-edge fragments of wide characters are blank, never clickable.

HitTest uses zero-based x/y, an exclusive right edge and last-hotspot precedence. Mouse coordinates from SSH clients must first be translated into the viewport's local frame.

These helpers are post-render projection and metadata resolution, not a lazy DataGrid, event loop, permission system, or replacement for host-side scroll state. The full frame is still rendered first; use bounded provider/pagination data for large record sets.

## Pure frame differences

To update only changed cells, compare two `RenderFrame` snapshots with `sml.DiffFrames(before, after)`. The result exposes `FrameDiff.Runs` (copies of changed row-cell intervals), `FullRedraw` for resize/new baseline or unrepresented style changes, and separate `HotspotsChanged` / `ElementsChanged` flags. The host decides whether partial repainting is worthwhile and still owns all terminal I/O, focus and authorization. See [frame-diffs.md](frame-diffs.md) for the complete contract and benchmark command.
