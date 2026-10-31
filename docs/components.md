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

`table` receives JSON rows as a *string binding* (`data.rows`). It supports bounded paging, sorting and filtering. Table `style` accepts `compact`, `compact-lite`, `compact-classic`, `default`, `rounded`, `unicode`, `markdown`, and `ascii`. Unlike other visual components, this is an existing table-specific preset, not CSS.

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
