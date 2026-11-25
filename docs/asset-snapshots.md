# Rendering assets and reload

Organize screens inside `assets/views/` in nested folders. A screen's name
must match its filename (without .sml), and names are globally unique.
Referenced components in `assets/components/` are compiled into the same
markup snapshot. Existing flat directories are supported.

Hosts can use `sml.NewAssetStore("assets", "home", "login")` and call
`store.Reload()` after a local edit. A failed reload retains the last good
revision, including its component definitions; required screens cannot
silently disappear. `store.View(name)` returns a detached view, revision,
and presence flag. Existing sessions can retain the old view until their
next render boundary. File watching, terminal painting, media assets and
action authorization remain host responsibilities.

Validation reports the relative source file and, for XML syntax errors or
unknown elements, the decoder line and column. Semantic diagnostics still
use descriptive errors without exact node coordinates.

## Shared design tokens and classes

Optionally create `assets/styles.json`:

```json
{
  "tokens": {"space": "1", "accent": "#aabbcc"},
  "classes": {
    "card": {"padding": "@space", "border": "round", "accent": "@accent"}
  }
}
```

A screen or component loaded through the asset loader can use
`<panel class="card" title="Overview">...</panel>`. Classes are applied
left to right; inline attributes override class defaults. Tokens may also
be referenced directly in visual attributes. The usual node-level checks
still apply. Actions, bindings and executable properties cannot be set
by styles. Shared classes and tokens are available through asset loading;
standalone Parse/ParseFile intentionally have no external stylesheet.

## Preview and inspect

```sh
sml preview -assets assets -width 80 -height 24 home
sml preview -assets assets -watch -ascii home
sml debug -assets assets -width 60 -scroll 5 home
```

The preview accepts repeatable `-set key=value` bindings and prints changed
frames in watch mode, retaining the last display on invalid edits. Debug
reports element bounds, visible hit targets and binding dependencies. No
actions are dispatched. `Frame.ASCII()` also exposes a geometry-preserving
fallback for host-selected older terminal profiles.
