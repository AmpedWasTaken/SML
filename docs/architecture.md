# Architecture

SML has a deliberately small scope: **markup → validated view tree → rendered terminal lines and interaction metadata**.

```text
  .sml file  ───┐
                ├─> Parse / LoadDir ──> View (validated)
  components/ ──┘                             │
                                            │ bindings
                                   ┌────────▼─────────┐
                                   │  ResolveBindings │
                                   │  layout / render │
                                   └────────┬─────────┘
                                            │
                                ┌───────────┴──────────┐
                                ▼                      ▼
                           []string lines         []Hotspot
                                │                      │
                                └──────────┬───────────┘
                                           ▼
                             Embedding Go TUI application
                        (auth, state, resize, keyboard/mouse,
                          redraw loop, action dispatch)
```

## Parsing

`Parse([]byte)` tokenizes and validates XML markup with `encoding/xml`. Unknown elements, XML namespaces, unsafe directives, excessive nesting, or unsupported control attributes fail before the view becomes usable. `LoadDir(assets)` reads named screens from `assets/views/` and verifies component includes.

## Rendering

The root facade delegates to the private engine; the implementation is isolated behind Go type aliases, preserving public method sets and data types. RenderWidth and RenderChecked interpolate string bindings, evaluate bounded `derived.*` values, select conditional branches, and produce terminal text-cell rows with width-aware layout. DataTable JSON is supplied by the embedding app, not fetched by the language. `RenderChecked` provides error propagation to the host.

## Interactivity

`RenderInteractive` produces logical line and cell geometry for interactive controls. It does not listen for key presses, maintain an event loop, execute commands or manage an SSH connection. The host maps logical coordinates to viewport positions and *re-authorizes* actions using the current authenticated identity.

## Module layout

- Root package `sml` (`sml.go`): thin compatibility facade with public type aliases and function delegation.
- `internal/engine/`: parser, validation, layout, rendering, reusable components, legacy controls, media, assets and their regression tests.
- `internal/textcell/`: private Unicode grapheme width, clipping, padding and wrapping primitives. Public compatibility helpers are forwarded through `sml.go`.
- `internal/engine/legacy_commands.go`: isolated compatibility allowlist for older host commands. New applications should use `action` metadata.
- `terminal/`: optional, independently importable RGB/gradient presentation helpers.
- `cmd/sml/`: CLI for validation and text rendering.
- `examples/`: sample views and a reusable component tree.
- `docs/`: language spec, component catalog, public API, CLI and security boundaries.
- `.github/`: CI, contribution templates, and dependency updates. See [project structure](project-structure.md) for ownership and placement.

## Compatibility and extensions

Views use the same .sml XML-like syntax as the original SML v1/v2/v3 engine. Richer UI controls, authorization, terminal-specific drawing, and runtime integration should be implemented in the host or via separately documented public APIs; silently adding executable tags would violate the language's trust model.

Contributions that change markup must include parser validation tests, rendering tests, malformed-input tests, and documentation.


## Structural rendering path

The new opt-in `RenderFrame` recursively lays out elements and records control bounds before composing terminal rows. `RenderInteractive` now consumes this layout, replacing label-based reverse matching with stable IDs and cell-accurate hitboxes. Width and clipping use `uniseg` grapheme clusters. Legacy `RenderChecked` remains available for v1-v3 string output compatibility.

See [cell-based rendering](layout.md).
