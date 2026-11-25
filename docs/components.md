## Localized interactive labels

Buttons, links and plugin actions can optionally select a host-owned
localization binding without changing their immutable command, page,
action or permission identifiers:

```xml
<screen name="controls" version="3">
  <button label="HOME" label-key="i18n.actions.home" page="home"/>
  <link label="HELP" label-key="i18n.actions.help" command="help"/>
</screen>
```

`label` remains mandatory, ASCII-only and serves as the English fallback.
`label-key` must be a valid `i18n.*` name: arbitrary user, state or plugin
bindings are rejected. The host passes localized display strings in the
`RenderFrame` bindings map. The same display label is used for both the
visual control and its returned hotspot, whose offsets are measured in
Unicode terminal cells (Chinese characters generally occupy two cells).
Long labels are clipped to at most 47 display cells; newlines, tabs and
terminal control sequences cannot alter layout geometry.

Localization changes **presentation only**: validated `page`, `command`,
`action`, plugin identifiers and host authorization remain unchanged.
If the key is absent or empty, the original ASCII label renders.

# Component reference

This is a concise reference to the currently supported XML-like SML elements. Each file begins with a named `<screen>`; `version="3"` enables derived data and v3 control gating.

## Layout and content

| Element | Purpose | Example |
| --- | --- | --- |
| `screen` | Root view | `<screen name="home" version="3">...</screen>` |
| `column` | Vertical stack | `<column><text>A</text><text>B</text></column>` |
| `row` | Horizontal layout | `<row><text>Left</text><text>Right</text></row>` |
| `grid` | Responsive or weighted columns | `<grid columns="3" weights="1,3,2" gap="2" collapse-below="65">...</grid>` |
| `panel` | Bordered section | `<panel title="Stats"><text>Ready</text></panel>` |
| `section` | Titled group | `<section title="Overview">...</section>` |
| `text` | Plain text / binding | `<text>Hello {{user.name}}</text>` |
| `badge` | Compact highlighted text | `<badge tone="success">OK</badge>` |
| `alert` | Highlighted message | `<alert tone="warning">Offline</alert>` |
| `metric` | Label and value | `<metric label="CPU" value="{{system.cpu}}"/>` |
| `field` | Bound key with fallback | `<field label="Status" key="system.health" fallback="unknown"/>` |
| `stat` | Label and referenced value | `<stat value="stats.count">Count</stat>` |
| `progress` | Percent indicator | `<progress label="CPU" value="{{system.cpu}}"/>` |
| `divider` | Horizontal rule | `<divider/>` |
| `spacer` | Empty rows | `<spacer size="1"/>` |
| `prompt` | Text prompt marker | `<prompt prefix="> "/>` |
| `banner` | Local ASCII file | `<banner src="logo.txt"/>` |

### Proportional panes in SML v3

For terminal apps with sidebars, inspectors, request/response panes, or metric
dashboards, set `weights` on a `grid` to distribute *terminal cells* by
relative share. Weights are **integers 1–16**, one for each declared column
(1–8 columns). The structural `RenderFrame` and `RenderInteractive` paths
use weighted widths; legacy `RenderWidth` preserves its older layout.

```xml
<screen name="workspace" version="3">
  <grid columns="3" weights="1,3,2" gap="1" collapse-below="70">
    <panel title="NAVIGATION"><text>Folders</text></panel>
    <panel title="WORKSPACE"><text>Requests or logs</text></panel>
    <panel title="INSPECTOR"><text>Details</text></panel>
  </grid>
</screen>
```

If the terminal is narrower than `collapse-below`, or any weighted pane
would fall below 14 cells, the grid stacks vertically. Short final rows
recalculate weights for their visible children. Hotspots keep the correct
positions after wrapping. Invalid weight vectors fail parsing, not rendering.

## Data and charts

```xml
<screen name="monitor">
  <gauge label="CPU" value="{{system.cpu}}" width="20"/>
  <sparkline label="Latency" values="{{system.latency_history}}" width="24"/>
  <table source="data.rows" style="rounded" sort="name" page-size="10">
    <column key="name" label="NAME" width="18"/>
    <column key="status" label="STATUS" width="12"/>
  </table>
</screen>
```

`table` receives JSON rows as a *string binding* (`data.rows`). It supports bounded paging, sorting and filtering. Table `style` accepts `compact`, `compact-lite`, `compact-classic`, `default`, `rounded`, `double`, `unicode`, `markdown`, and `ascii`. Omitting `style` uses `rounded`. Use `double` for double-line borders, or `compact-lite` to retain the old compact layout. Narrow tables stack their fields when a complete frame cannot fit. Unlike other visual components, this is an existing table-specific preset, not CSS.

## Control flow and computed fields

```xml
<screen name="status" version="3">
  <derive key="derived.state" op="coalesce" source="app.health" fallback="unknown"/>
  <if field="app.ready" equals="true">
    <badge>Ready</badge>
    <otherwise><alert>Not ready</alert></otherwise>
  </if>
  <when field="app.has_warnings">
    <text>Warnings available</text>
  </when>
  <empty field="app.messages">
    <text>Nothing to display</text>
  </empty>
  <repeat source="app.items" as="item" limit="10">
    <text>{{item}}</text>
  </repeat>
</screen>
```

`derive` supports bounded `coalesce`, `equals`, `concat`, `duration`, `uppercase`, `lowercase` and `sum` operations. Earlier derivations can feed later ones. The `repeat` element consumes comma-separated values (not arbitrary executable expressions).

## Reusable components

```xml
<!-- views/home.sml -->
<screen name="home"><include name="summary"/></screen>
```

```xml
<!-- components/summary.sml -->
<component name="summary"><panel title="Shared"><text>Hello</text></panel></component>
```

Validate the whole asset root with `sml.LoadDir("assets")` or the CLI `sml validate assets/`.

## Native panel and table glow (SML v3)

The existing \`<panel>\` and bounded \`<table>\` accept **opt-in** presentation
attributes. No extra row/column components, arbitrary ANSI or CSS are needed.

\`\`\`xml
<screen name="plugins" version="3">
  <panel title="Installed plugins" border="round" padding="1"
         glow="soft" glow-color="#72D6EA" accent="#A590F1">
    <table source="plugin.rows" style="rounded" glow="medium"
           glow-color="#72D6EA" header-style="accent" row-style="striped">
      <column key="name" label="PLUGIN" width="18"/>
      <column key="state" label="STATE" width="12"/>
      <column key="runtime" label="RUNTIME" width="16"/>
    </table>
  </panel>
</screen>
\`\`\`

The existing \`table source=\` is a JSON **string binding** (up to 200 rows);
the new styles do not execute code and do not relax table or action limits.

| Component | New options |
| --- | --- |
| \`panel\` | \`glow="none|soft|medium|strong"\`, \`glow-color="#RRGGBB"\`, \`accent="#RRGGBB"\` |
| \`table\` | Same glow/hex options, \`header-style="default|accent|muted|solid"\`, \`row-style="plain|striped|compact"\` |

Glow is an approximation using ANSI foreground color on existing border
strokes, a brighter top edge, and a *small* tint only behind nonspace table
header/striped-row characters. **No blur, overdraw, artificial halo columns,
cursor movement, background jobs, or blinking**. This keeps fixed-width
geometry, focus targets, and animated TUI input safe.

The plain \`RenderFrame\`, \`RenderWidth\` and \`RenderInteractive\` APIs remain
colorless and backward-compatible. A color-capable Go host explicitly opts
into styling:

\`\`\`go
frame, err := view.RenderFrame(bindings, 80)
if err != nil { return err }
rows, err := sml.PaintGlowFrame(frame, sml.GlowOptions{TrueColor: true})
if err != nil { return err }
// rows contain ANSI SGR color only: host owns terminal write/diff/cursor.
\`\`\`

Or call \`view.RenderGlowInteractive(bindings, width, opts)\` for styled rows
with the original hotspot metadata. The host must preserve validated SGR in
its final paint path; sending rows through a sanitizer that discards ANSI
will intentionally remove the effect.

On older SSH/PuTTY clients, pass \`GlowOptions{NoColor: true}\` or omit
\`TrueColor\` to get the original geometry and text exactly. Glyph/font
fallback remains the terminal host's responsibility. Color attributes are
validated as full six-digit hexadecimal values and cannot inject escapes.

## Input and interactive metadata

```xml
<screen name="controls" version="3">
  <button label="Help" command="help"/>
  <choice label="Region" key="state.region" options="eu,us"/>
  <toggle label="Enabled" key="state.enabled"/>
  <input label="Name" key="state.name"/>
  <number label="Retries" key="state.retries" min="0" max="5"/>
</screen>
```

The legacy command allowlist is intentionally narrow and the standalone engine never invokes commands. The host decides what a supported command, page or plugin action actually does. Call `view.RenderInteractive(...)` to receive hitbox metadata.

## Inline styles, gradients and media

Some visual elements support `tone`, `variant`, `border`, `padding`, `align`, and `case` attributes with validated values. For example:

```xml
<panel title="Health" border="round" tone="neutral">
  <badge variant="outline" tone="success">Online</badge>
</panel>
```

**Note:** styling is limited to documented element/attribute combinations. No arbitrary CSS or `style=` is accepted on general components.

Media uses local, bounded files in `assets/media/`, rendered as text-cell luminance/block approximations:

```xml
<image src="logo.png" width="32" height="12" mode="blocks" alt="Logo"/>
<gif src="loading.gif" width="32" height="12" fps="8" mode="ascii" alt="Loading"/>
```

Appearance gradient declarations use `<screen name="appearance">` with `<gradient .../>` elements, parsed into `terminal.GradientStyle`. Painting colored text is the responsibility of your application.


## Custom components with props and slots

SML supports typed *string props* and named slots in reusable component files. For example, put this in `assets/components/stat-card.sml`:

```xml
<component name="stat-card" props="title:string,value:string">
  <panel title="{{props.title}}">
    <metric label="Current" value="{{props.value}}"/>
    <slot name="footer">
      <text>Default footer</text>
    </slot>
  </panel>
</component>
```

Use it from a view:

```xml
<screen name="dashboard" version="3">
  <use name="stat-card" title="Active users" value="{{app.users}}">
    <fill slot="footer"><badge>Live</badge></fill>
  </use>
</screen>
```

Props are declared explicitly, limited in size, and rendered as data. Missing or undeclared props and unknown fill slots are errors. Slot contents do not execute code. Use `LoadDir` to validate the full component tree, including cycle detection.

## Host actions and element identifiers

Prefer generic `action` IDs for standalone applications:

```xml
<button id="refresh" label="Refresh" action="metrics.refresh"/>
```

`RenderFrame` and `RenderInteractive` return a `Hotspot` with `ID` and `Action`. Applications handle the action after their own authorization checks. Legacy `command` and `page` targets remain for existing SML v1–v3 views; they are not a general-purpose execution interface.

## Global terminal styles with SMLCSS

Put `styles.smlcss` next to the asset directory's `views/` folder (or use
the existing `styles.json` format, but not both). The asset loader validates
the stylesheet together with the views and components.

```css
:root {
    --edge: round;
    --space: 1;
}

panel {
    border: var(--edge);
    padding: var(--space);
}

button { border: var(--edge); }
input { border: var(--edge); }
choice { border: var(--edge); }
toggle { border: var(--edge); }
number { border: var(--edge); }
table { style: rounded; }

.important { border: double; }
.classic { border: ascii; }
```

Use `class="important"` on a compatible SML v3 component.
The precedence is component default, then classes in declaration order on
the element, then explicit inline attributes. Use `border="none"` to
retain an unframed control. Inputs with `appearance="boxed"` retain
their existing full-width input layout.

SMLCSS is a deliberately small terminal styling grammar, not browser CSS.
It does not support imports, scripts, arbitrary nested selectors, or
arbitrary properties. Supported control states and media rules are described below. Bounded animation keyframes are described below. Component selectors and classes accept only validated
presentation attributes. Table presets use `style` on the table selector.
Existing explicit styles remain valid.

### Extended SMLCSS layout and animation

Group component selectors with commas and split their declarations over
several rules. Later declarations for the same component or class replace
earlier ones; the established class/inline precedence remains unchanged.

```css
:root {
    --accent: #67d9a5;
}

panel, button, input {
    border: round;
}

panel {
    color: var(--accent);
    background-color: #09110d;
}

box {
    width: 80%;
    min-width: 24ch;
    max-width: 100ch;
}

grid {
    grid-template-columns: 1fr 2fr;
    column-gap: 2ch;
    row-gap: 1;
    align-items: center;
    collapse-below: 60ch;
}
```

`box` widths can be terminal cells (plain integers or CSS `ch` units),
percentages of the parent, or `auto`. The terminal's available width always
wins over a minimum width. `grid-template-columns` accepts 1–8 integer
fractional tracks from `1fr` to `16fr`. Use `RenderFrame` for weighted
geometry; the legacy text renderer retains its equal-width grid behavior.
Rows and grids accept `row-gap` (0–8 rows), `column-gap` (alias for
`gap`), `align-items: start|center|end` and `collapse-below` (20–240 cells).
The collapse threshold is relative to the container width.

CSS aliases map to existing SML attributes: `color` to `foreground`,
`background-color` to `background`, `text-align` to `align`, and
`text-transform: uppercase|lowercase|none` to `case`. The applicable
component must support the underlying property.

### Animation shorthand and custom brightness keyframes

```css
@keyframes breathe {
    from { brightness: 40%; }
    50%  { brightness: 100%; }
    to   { brightness: 40%; }
}

panel {
    border: double;
    glow: medium;
    glow-color: #67d9a5;
    animation: breathe 2s ease-in-out infinite;
}

table {
    glow: soft;
    animation: sweep 1200ms linear infinite;
}
```

The animation shorthand accepts an existing effect (`pulse`, `sweep`,
`fade`) or a defined keyframe name, a duration (250ms–10s), an optional
easing (`linear`, `ease-in`, `ease-out`, `ease-in-out`), and an optional
iteration value (`1`–`100` or `infinite`). Brightness keyframes interpolate brightness; they require 0% and 100% endpoints, at most eight stops,
and brightness values of 30–100%. They cannot change geometry, visibility,
or action metadata. Only panels and tables support animation.

For animation, the host must call `PaintGlowFrame` or
`RenderGlowInteractive` with truecolor, animation enabled and a changing
phase. For example:

```go
rows, err := sml.PaintGlowFrame(frame, sml.GlowOptions{
    TrueColor: true,
    Animate:   true,
    Phase:     elapsed,
})
```

Animation is quantized to the existing 12 FPS rendering phase. Hosts can
disable color or respect reduced-motion preferences. SMLCSS does not own
the event loop or schedule background work. Browser CSS transforms, general property keyframes, and transitions are not
implemented. The following section describes the supported subset of control
states, media queries, and flex layout.

### Responsive media queries, flex rows and control states

```css
row {
    display: flex;
    flex-direction: row;
    gap: 2ch;
    justify-content: space-between;
    align-items: start;
}

.sidebar {
    flex-basis: 24ch;
    flex-grow: 0;
}

.content {
    flex-basis: 20ch;
    flex-grow: 1;
}

@media (max-width: 70ch) {
    row {
        flex-direction: column;
        row-gap: 1;
    }
    grid {
        grid-template-columns: 1fr;
    }
}

button {
    border: round;
    color: #FFFFFF;
}

button:hover {
    color: #67D9A5;
}

button:focus {
    border-color: #00CCFF;
}

button:disabled {
    color: #777777;
}
```

Use flex sizing on box children of a row:

```xml
<screen name="home" version="3">
    <row>
        <box class="sidebar">
            <panel title="Navigation">
                <button id="home" label="Home" action="nav.home"/>
            </panel>
        </box>
        <box class="content">
            <panel title="Workspace">
                <text>Welcome back</text>
            </panel>
        </box>
    </row>
</screen>
```

A flex row supports row/column direction, up to eight horizontal children,
integer grow factors (0–16), cell or percentage bases, gaps,
`justify-content: start|center|end|space-between`, and vertical
`align-items: start|center|end`. Items use a minimum horizontal allocation
of 14 cells. When their bases do not fit, the row stacks instead of
clipping controls. This is a bounded terminal-specific subset, not browser
Flexbox: shrink factors and wrap lines are opt-in as described below; arbitrary
positioning is not implemented. Use flex-basis rather than width on flex item boxes when
allocating space in the parent row.

Media rules support one inclusive min-width or max-width condition
(12–240 cells, relative to the full terminal), and box, grid, or row
component selectors with layout properties. Matching media rules replace
global component defaults in source order; explicit inline attributes and
classes still win. Grid track changes must specify the full track vector.

State selectors accept control names followed by `:hover`, `:focus` or
`:disabled`. They can change foreground, background and border colors only.
They cannot change border geometry, enabled state, or action metadata.
The host supplies focus and hover IDs from the rendered frame:

```go
rows, err := sml.PaintGlowFrame(frame, sml.GlowOptions{
    TrueColor: true,
    FocusedID: focusedID,
    HoveredID: hoveredID,
})
```

Disabled state is derived from the rendered control, not from CSS.
Disabled paint takes precedence over focus, which takes precedence over
hover for overlapping properties. No-color clients retain the plain frame.
CSS does not manage mouse or keyboard events or grant permissions.

### Bounded scroll containers

```css
.scroll-area {
    overflow: auto;
    max-height: 20;
}
```

```xml
<screen name="logs" version="3">
    <box class="scroll-area" scroll-key="state.scroll">
        <repeat source="app.messages" as="message" limit="100">
            <text>{{message}}</text>
        </repeat>
    </box>
</screen>
```

The host supplies `state.scroll` as a nonnegative integer row offset.
The renderer clamps the offset to the available content, clips the box
to its maximum height (1–3000 terminal rows), translates visible element
bounds and removes off-screen hotspots. `overflow: hidden` clips without
scrolling; `overflow: visible` preserves normal box behavior. The host owns
wheel/key events and scroll state. This is a bounded post-render viewport,
not lazy rendering or a terminal event loop. `max-height` is measured in
rows, not columns.

### Flexible shrinking and wrapping

```css
row {
    display: flex;
    flex-wrap: wrap;
    gap: 2ch;
    row-gap: 1;
}

.card {
    flex-basis: 32ch;
    flex-grow: 1;
    flex-shrink: 1;
    min-width: 16ch;
}
```

`flex-wrap: wrap` forms additional rows in source order when the natural
item widths do not fit. Each line contains at most eight items. Growth and
justification are calculated independently for each line. The default
`nowrap` preserves earlier behavior.

`flex-shrink` accepts an integer 0–16 (default 0). If a line's natural
width exceeds available space, shrinkable items give up cells in
proportion to their base width multiplied by the shrink factor. They
never shrink below their minimum allocation (at least 14 cells); if the
line still cannot fit, the engine stacks it. Integer rounding is
deterministic. `collapse-below` and column direction still take
precedence. When the terminal is narrower than a single item's minimum,
the full available width wins.

These properties are supported on flex rows and their box children
respectively. Layout uses terminal cells, not browser pixels. The
structural renderer maintains original child paths and translates
element bounds and hitboxes for every wrapped line. This does not
implement browser flex-wrap reverse, order, or arbitrary positioning.

### Positioned layers

A box with `position: relative` establishes a local coordinate system.
Direct child boxes with `position: absolute` are removed from the normal
vertical flow and painted over its ordinary children.

```css
.workspace {
    position: relative;
}

.notice {
    position: absolute;
    left: 12ch;
    top: 2;
    width: 36ch;
    z-index: 10;
}
```

```xml
<screen name="workspace" version="3">
    <box class="workspace">
        <panel title="Workspace">
            <text>Preferences</text>
            <text>Appearance</text>
            <text>Accessibility</text>
        </panel>
        <box class="notice">
            <panel title="Notification" border="double">
                <text>Preferences saved</text>
                <button label="Dismiss" action="notice.dismiss"/>
            </panel>
        </box>
    </box>
</screen>
```

Horizontal offsets are terminal cells (0–240), vertical offsets are rows
(0–3000), and layer order is 0–100. Equal z-index values paint in source
order, with the later box on top. A relative box supports up to 16 absolute
children. The layer's horizontal origin is clamped to leave at least 12
cells in narrow terminals, and output remains subject to the 3000-row limit.
Overlays are opaque, including whitespace. A lower hotspot that intersects
a higher layer is removed entirely; partially obscured lower element
paint metadata is also omitted to prevent color leaking through overlays.
Wide graphemes crossing a layer edge are cleared as whole clusters.

This is a local compositing mechanism, not a complete browser stacking
context implementation. Absolute positioning must be on a direct box
child of a relative box; position properties cannot be altered by media
queries. Fixed positioning, transparent layers, focus traps and host-managed modal
navigation are not implemented. Sticky scroll headers are described below.

### Sticky scroll headers

```css
.scroll-area {
    overflow: auto;
    max-height: 18;
}

.header {
    position: sticky;
    top: 0;
}
```

```xml
<screen name="activity" version="3">
    <box class="scroll-area" scroll-key="state.scroll">
        <box class="header">
            <text>Recent activity</text>
            <divider/>
        </box>
        <repeat source="app.messages" as="message" limit="100">
            <text>{{message}}</text>
        </repeat>
    </box>
</screen>
```

The sticky box must be the first direct child of a non-relative box
with `overflow: auto`. Its rows remain at the top of the viewport
while the body scrolls independently. `top` reserves rows above the
header; the header and offset consume part of the parent's maximum
height. The body receives the remaining height, and off-screen
hitboxes are clipped. If the header fills the viewport, the body is
not rendered. This is a pinned header, not general browser sticky
positioning: arbitrary in-flow sticky elements, sticky footers and
fixed screen positioning are not included.

### CSS color transitions

```css
button {
    color: #888888;
    transition: color 200ms ease-in-out;
}

button:hover {
    color: #67D9A5;
}

button:focus {
    color: #FFFFFF;
}
```

Transitions are supported on control component selectors. The shorthand
accepts `color`, `foreground`, `background-color`, `background`,
`border-color` or `all`; one duration (16ms–10s); and optional easing
(`linear`, `ease-in`, `ease-out`, `ease-in-out`). Only explicit RGB
endpoints interpolate; terminal defaults and gradients switch directly.
Disabled styling takes effect immediately. Reduced-motion mode skips
interpolation.

The host supplies previous and current focus/hover IDs, elapsed time,
and redraw scheduling. For example:

```go
rows, err := sml.PaintGlowFrame(frame, sml.GlowOptions{
    TrueColor:         true,
    Transition:        true,
    PreviousHoveredID: previousID,
    HoveredID:         currentID,
    TransitionElapsed: elapsed,
})
```

SML performs deterministic RGB interpolation without changing glyphs,
bounds, hit targets or bindings. There is no internal timer or retained
transition state. Hosts that interrupt a running transition should
manage their own timing; this API interpolates between state endpoints,
not arbitrary in-flight colors. Layout transitions and transition lists are not yet supported. Color
keyframes are described below.

### RGB keyframes and playback controls

```css
@keyframes aurora {
    from { border-color: #67D9A5; }
    50%  { border-color: #627BFF; }
    to   { border-color: #B56CFA; }
}

panel {
    border: double;
    border-color: #67D9A5;
    animation: aurora 2s ease-in-out;
    animation-delay: 250ms;
    animation-iteration-count: 4;
    animation-direction: alternate;
}
```

Each animation can interpolate one property: brightness, foreground
(`color`), background (`background-color`), border-color or glow-color.
RGB keyframes use explicit six-digit hex colors and two to eight distinct
stops including 0% and 100%. All stops in a track must animate the same
property. RGB interpolation is performed in the paint stage without
altering frame geometry or hotspots.

Playback options apply to built-in effects and custom keyframes:

- `animation-delay`: 0ms–10s.
- `animation-iteration-count`: 1–100 or infinite. Specify either this
  property or the iteration field of the animation shorthand, not both.
- `animation-direction`: normal, reverse, alternate, alternate-reverse.

Cycle boundaries and finite final endpoints are preserved exactly;
intermediate frames are quantized to the existing 12 FPS cadence.
Before the delay, RGB keyframes retain static styles. No-color and
reduced-motion modes do not animate. Host-provided phase and redraw
scheduling remain required. Multiple concurrent animation tracks,
named CSS colors, transform animations and arbitrary geometry
keyframes are not implemented.

### Compound selectors and specificity

```css
button {
    border: round;
    color: #CCCCCC;
}

.primary {
    color: #67D9A5;
}

button.primary {
    background-color: #14291E;
}

.primary.compact {
    padding: 0;
}

#save {
    color: #FFFFFF;
}

button#save.primary {
    border: double;
}
```

```xml
<button id="save" class="primary compact" label="Save" action="app.save"/>
```

SMLCSS supports a component name, multiple classes and one explicit ID
in a compound selector. Classes can be declared solely by appearing in
a compound selector, so no empty marker rule is necessary. Grouped
selectors also support compound rules.

The cascade preserves existing behavior for backwards compatibility:
component defaults, simple classes in markup order, matching compound
rules by specificity, then explicit inline attributes. Specificity
compares ID count, class count and component name in that order; ties
are resolved by stylesheet source order. Media defaults remain below
class, compound and inline declarations.

Selectors match literal markup IDs before rendering, not generated
repeat-instance IDs. They cannot assign IDs, action names, enabled
conditions or other behavioral metadata. Descendant/child combinators and compound pseudo-selectors are described below. Attribute selectors are described below. Universal selectors and `!important` are not supported.

### Contextual selectors and compound control states

```css
panel.workspace button {
    color: #AAB8B0;
}

panel.workspace > button {
    border: round;
}

panel.workspace button.primary:hover {
    color: #67D9A5;
}

button#save.primary:focus {
    border-color: #FFFFFF;
}
```

A space matches a descendant; `>` matches an immediate child.
Selector chains support up to four component/class/ID segments.
Specificity includes every segment; equal scores use stylesheet order.
The matcher uses source markup ancestry during asset loading, including
component definitions in their own source tree, not dynamically expanded
call-site ancestry.

The final segment may include `:hover`, `:focus`, or `:disabled`
when it explicitly names a supported control component. Contextual state
rules override component-wide state rules by specificity. These rules
can change only foreground, background or border colors; they cannot
change layout, bindings or action metadata. State paint is determined
at load time and applied at render time using host-provided state IDs.
Disabled paint remains higher priority than hover and focus.

This is selector chaining, not nested CSS declaration block syntax.
Contextual media selectors, sibling combinators and `!important` are not included. Attribute selectors and explicit inheritance are described below.

### Visual attribute selectors

```css
button[variant="outline"] {
    color: #67D9A5;
}

button[border][variant=outline]:hover {
    color: #FFFFFF;
}

panel.workspace button[variant='solid'] {
    background-color: #14291E;
}
```

Attribute selectors match literal source markup before any style defaults
are applied. They support presence and exact equality on the visual
attributes variant, tone, appearance, border, align and case. Equality
values may be unquoted or surrounded by matching single or double
quotes; whitespace inside brackets and substring or pattern operators
are not supported. Each segment can have up to four attribute filters.
Each filter contributes class-level specificity. Behavior attributes
and dynamic bindings are deliberately excluded.

### Explicit color inheritance

```css
panel.workspace {
    color: #67D9A5;
    background-color: #14291E;
}

panel.workspace button[variant=outline] {
    color: inherit;
    background-color: inherit;
}
```

The foreground and background properties on boxes, panels, buttons
and text may explicitly use `inherit`. They resolve to the closest
ancestor's computed explicit RGB color in the source tree, through
unpainted wrapper nodes. If no such color exists, the property is
removed and normal terminal painting applies. Existing elements do
not automatically acquire new inherited declarations; the opt-in
keyword keeps previously published layouts compatible. The usual
inline precedence still applies.

Inheritance is resolved during asset loading, not by dynamic render
state. Interactive state selectors require explicit RGB colors.
Layout values, gradients and arbitrary CSS properties cannot inherit.
