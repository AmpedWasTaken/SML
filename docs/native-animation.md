# Native panel and table animations

SML v3 supports declarative, host-timed visual effects on panels and tables:

```xml
<screen name="animation-preview" version="3">
  <panel title="Preview" glow="medium" glow-color="#A694EE"
         animate="sweep" duration-ms="1800" easing="ease-in-out">
    <text>Stable content and interaction geometry</text>
  </panel>
</screen>
```

Supported effects: pulse (looping brightness), sweep (looping highlight),
fade (one-shot entrance). Duration is 250–10000 ms. Easing: linear,
ease-in, ease-out, ease-in-out. The host chooses the phase and resets it on
view entry if desired. Animation is quantized to 12 FPS, uses truecolor SGR
only, never hides glyphs or changes layout, and has no timers or I/O.

Use RenderGlowInteractive with GlowOptions{TrueColor:true, Animate:true,
Phase:elapsed}. ReducedMotion or NoColor disables motion or all color,
respectively. Existing RenderFrame and RenderChecked remain plain.
No executable callbacks or changes to authorization are introduced.
