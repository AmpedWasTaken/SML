# SML language reference

SML uses strict XML-like markup. The root is a named `screen`, with an optional `version="1"`, `version="2"` or `version="3"` (v3 enables derived bindings). The name is used for locating reusable views.

```xml
<screen name="home" version="3">
  <text>Welcome {{user.name}}</text>
  <panel title="Summary">
    <metric label="Online" value="{{stats.online}}"/>
  </panel>
</screen>
```

## Elements

| Group | Elements |
| --- | --- |
| Layout | `screen`, `component`, `column`, `row`, `grid`, `panel`, `section`, `spacer`, `divider`, `include` |
| Text and metrics | `text`, `banner`, `badge`, `alert`, `prompt`, `stat`, `metric`, `field`, `progress` |
| Data | `table`, `gauge`, `sparkline` |
| Branching | `if`, `when`, `otherwise`, `empty`, `repeat`, `derive` |
| Interaction | `button`, `link`, `plugin-action`, `toggle`, `choice`, `input`, `number` |
| Media and appearance | `image`, `gif`, `gradient` |

The parser validates supported elements and bounded attributes. Many elements accept style attributes, including `tone`, `align`, `case`, `padding`, and `border`. Unsupported combinations fail validation.

## Bindings

`{{user.name}}` reads a string from the map passed by your Go host. Unknown keys are replaced with blank text. Dynamic values are sanitized before presentation. The host decides which data to provide.

## Conditions

```xml
<if field="service.ready" equals="true">
  <badge>Ready</badge>
  <otherwise><alert>Not ready</alert></otherwise>
</if>
```

`<otherwise>` must be the last child of an `if` or `when`.

## Derived bindings (v3)

```xml
<derive key="derived.health" op="coalesce"
        source="service.health" fallback="unknown"/>
<field label="Health" key="derived.health"/>
```

Operators: `coalesce`, `equals`, `concat`, `duration`, `uppercase`, `lowercase`, and `sum`. Computations are ordered, bounded, and non-executable. At most 24 derivations may be declared per view.

## Components and files

```text
assets/
  views/
    home.sml
  components/
    header.sml
  art/
    logo.txt
  media/
    status.gif
```

Load the entire tree with `sml.LoadDir("assets")`. `<include name="header"/>` resolves `assets/components/header.sml` and validates bounded recursion. Only literal, safe component filenames are supported. Use `view.WithAssets("assets")` when manually parsing a view.

## Tables and media

A `<table>` supports a **bounded table-specific** `style` preset (`compact`, `compact-lite`, `compact-classic`, `default`, `markdown`, `rounded`, `unicode`, or `ascii`). This is distinct from generic inline styling: a `style` attribute on other components is rejected.


`table` supports data columns, sorting, filtering and bounded pages; supply table data through host-provided bindings (typically JSON data according to the table schema). `image` and `gif` decode local files under `assets/media/` into text-cell artwork, not true terminal pixels. For detailed examples, consult the original engine tests in this repository.

## Interaction

`button`, `link`, and `plugin-action` describe actions for the Go host to authorize and dispatch. `toggle`, `choice`, `input`, and `number` describe session-local state widgets. The engine doesn't run an input event loop, network handler or command runner. See [Go integration](go-api.md) and [Security](security.md).

## Stable keyed repeats (SML v3, opt-in)

By default, repeat instances keep their existing position-based control IDs. A
dynamic list that can reorder may opt into value keys to preserve per-item
hotspot IDs and focus metadata.

```xml
<screen name="inventory" version="3">
  <repeat source="items" as="item" key-by="value" index-as="idx" limit="10">
    <text>{{idx}}: {{item}}</text>
    <button id="inspect" label="Inspect" action="inventory.inspect"/>
  </repeat>
</screen>
```

The host supplies `items` as a comma-separated list of **unique**, nonempty
identifiers such as `server-1,server-2,server-3`. Each item is trimmed and its
bounded UTF-8 value (at most 128 bytes) determines a non-executable, opaque,
hash-derived path suffix. Reordering these identifiers retains each
control's generated ID, and `index-as="idx"` exposes its current **zero-based**
visible index to child bindings. Explicit child IDs such as `inspect` remain
per-item scoped; a reorder cannot transfer the old row's ID to another item.

Duplicate or missing keys fail rendering deterministically, rather than
silently sharing a control state. Keyed source input is capped at 32 KiB and
displays at most the existing 100-item limit. Only **visible items within the
configured repeat limit** are checked for duplicates. Commas are separators
and cannot be part of a key in this mode; a future typed data-provider
interface can support more complex records.

`index-as` must be a distinct safe local identifier and cannot shadow
reserved presentation namespaces. Neither `key-by` nor `index-as` exists in
SML v1/v2. Unkeyed repeats preserve their established positional IDs and
rendering semantics. Both `RenderChecked` and `RenderFrame` validate keyed
values consistently. No markup may execute code or authorize a host action.
