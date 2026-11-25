# Premium widget library

The 32 display-only widgets are available through the host-driven `premiumwidgets` package.
They render bounded terminal text frames with ASCII and Unicode variants, reduced motion,
and local interaction state. Hosts supply the animation phase and redraw schedule.
No widget starts a goroutine, runs commands, accesses a network or reads host data.

## SML integration

Use `premiumwidgets.Render(id, input)` and inject `frame.Bindings("preview")` into
any validated SML view. For example:

```xml
<screen name="sample" version="3">
  <panel title="Widget">
    <text>{{preview.line1}}</text>
    <text>{{preview.line2}}</text>
    <text>{{preview.line3}}</text>
    <text>{{preview.line4}}</text>
    <text>{{preview.line5}}</text>
    <text>{{preview.line6}}</text>
  </panel>
</screen>
```

`Gallery(route, width, phase, state)` returns a parsed SML view, presentation bindings,
an animation flag and an error. The six routes cover all 32 widgets. The gallery starts
in ASCII and supports switching to Unicode and reduced motion. Its data is synthetic.
The hosting application must implement local control updates and use its existing
terminal input/paint loop, at no more than 12 FPS. The visual gradient is a glyph
density effect, not a truecolor gradient.

## Sentinel

Open the fullscreen TUI and run `plugins open premium-widgets`.
The built-in presentation gallery requires the existing signed `premium_themes`
appearance entitlement; it grants no other permissions. No plugin executable,
real telemetry or additional host command is involved. The Sentinel integration
currently uses a source snapshot of the SML package until a compatible SML module
revision can be pinned. The library is not a new native SML XML tag.
