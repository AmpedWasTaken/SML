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
