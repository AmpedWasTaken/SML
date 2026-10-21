# Component reference

This is a concise reference to the currently supported XML-like SML elements. Each file begins with a named \`<screen>\`; \`version="3"\` enables derived data and v3 control gating.

## Layout and content

| Element | Purpose | Example |
| --- | --- | --- |
| \`screen\` | Root view | \`<screen name="home" version="3">...</screen>\` |
| \`column\` | Vertical stack | \`<column><text>A</text><text>B</text></column>\` |
| \`row\` | Horizontal layout | \`<row><text>Left</text><text>Right</text></row>\` |
| \`grid\` | Responsive columns | \`<grid columns="3" gap="2" collapse-below="65">...</grid>\` |
| \`panel\` | Bordered section | \`<panel title="Stats"><text>Ready</text></panel>\` |
| \`section\` | Titled group | \`<section title="Overview">...</section>\` |
| \`text\` | Plain text / binding | \`<text>Hello {{user.name}}</text>\` |
| \`badge\` | Compact highlighted text | \`<badge tone="success">OK</badge>\` |
| \`alert\` | Highlighted message | \`<alert tone="warning">Offline</alert>\` |
| \`metric\` | Label and value | \`<metric label="CPU" value="{{system.cpu}}"/>\` |
| \`field\` | Bound key with fallback | \`<field label="Status" key="system.health" fallback="unknown"/>\` |
| \`stat\` | Label and referenced value | \`<stat value="stats.count">Count</stat>\` |
| \`progress\` | Percent indicator | \`<progress label="CPU" value="{{system.cpu}}"/>\` |
| \`divider\` | Horizontal rule | \`<divider/>\` |
| \`spacer\` | Empty rows | \`<spacer size="1"/>\` |
| \`prompt\` | Text prompt marker | \`<prompt prefix="> "/>\` |
| \`banner\` | Local ASCII file | \`<banner src="logo.txt"/>\` |

## Data and charts

\`\`\`xml
<screen name="monitor">
  <gauge label="CPU" value="{{system.cpu}}" width="20"/>
  <sparkline label="Latency" values="{{system.latency_history}}" width="24"/>
  <table source="data.rows" style="rounded" sort="name" page-size="10">
    <column key="name" label="NAME" width="18"/>
    <column key="status" label="STATUS" width="12"/>
  </table>
</screen>
\`\`\`

\`table\` receives JSON rows as a *string binding* (\`data.rows\`). It supports bounded paging, sorting and filtering. Table \`style\` accepts \`compact\`, \`compact-lite\`, \`compact-classic\`, \`default\`, \`rounded\`, \`unicode\`, \`markdown\`, and \`ascii\`. Unlike other visual components, this is an existing table-specific preset, not CSS.

## Control flow and computed fields

\`\`\`xml
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
\`\`\`

\`derive\` supports bounded \`coalesce\`, \`equals\`, \`concat\`, \`duration\`, \`uppercase\`, \`lowercase\` and \`sum\` operations. Earlier derivations can feed later ones. The \`repeat\` element consumes comma-separated values (not arbitrary executable expressions).

## Reusable components

\`\`\`xml
<!-- views/home.sml -->
<screen name="home"><include name="summary"/></screen>
\`\`\`

\`\`\`xml
<!-- components/summary.sml -->
<component name="summary"><panel title="Shared"><text>Hello</text></panel></component>
\`\`\`

Validate the whole asset root with \`sml.LoadDir("assets")\` or the CLI \`sml validate assets/\`.

## Input and interactive metadata

\`\`\`xml
<screen name="controls" version="3">
  <button label="Help" command="help"/>
  <choice label="Region" key="state.region" options="eu,us"/>
  <toggle label="Enabled" key="state.enabled"/>
  <input label="Name" key="state.name"/>
  <number label="Retries" key="state.retries" min="0" max="5"/>
</screen>
\`\`\`

The legacy command allowlist is intentionally narrow and the standalone engine never invokes commands. The host decides what a supported command, page or plugin action actually does. Call \`view.RenderInteractive(...)\` to receive hitbox metadata.

## Inline styles, gradients and media

Some visual elements support \`tone\`, \`variant\`, \`border\`, \`padding\`, \`align\`, and \`case\` attributes with validated values. For example:

\`\`\`xml
<panel title="Health" border="round" tone="neutral">
  <badge variant="outline" tone="success">Online</badge>
</panel>
\`\`\`

**Note:** styling is limited to documented element/attribute combinations. No arbitrary CSS or \`style=\` is accepted on general components.

Media uses local, bounded files in \`assets/media/\`, rendered as text-cell luminance/block approximations:

\`\`\`xml
<image src="logo.png" width="32" height="12" mode="blocks" alt="Logo"/>
<gif src="loading.gif" width="32" height="12" fps="8" mode="ascii" alt="Loading"/>
\`\`\`

Appearance gradient declarations use \`<screen name="appearance">\` with \`<gradient .../>\` elements, parsed into \`terminal.GradientStyle\`. Painting colored text is the responsibility of your application.
