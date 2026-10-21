# Integrating SML into a Go app

## Parse and render an in-memory view

```go
view, err := sml.Parse([]byte(`<screen name="home"><text>Hello {{name}}</text></screen>`))
if err != nil { return err }
lines := view.RenderWidth(map[string]string{"name": "developer"}, 80)
for _, line := range lines { fmt.Println(line) }
```

`Parse([]byte)` validates the syntax, supported elements, styles, and declared controls. `Render(values)` uses the default width 80; `RenderWidth(values, width)` renders terminal text lines. Rendering uses the host's string bindings and never executes markup.

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
