# Project structure

SML uses a small root package for external consumers and private packages for its implementation. In Go, **each directory is a separate package**, so moving types without a facade would break imports and exported method sets.

```text
SML/
├── sml.go                    # Single public API facade (import github.com/AmpedWasTaken/SML)
├── api_test.go               # External-package compatibility tests
├── internal/
│   ├── engine/               # Core implementation + package-private regression tests
│   │   ├── parse.go          # XML grammar and validation
│   │   ├── types.go          # Internal view and node types
│   │   ├── render_legacy.go  # v1-v3 compatible rendering
│   │   ├── layout.go         # Cell-based layout and hitboxes
│   │   ├── components.go     # Custom components, props, slots
│   │   ├── interaction.go    # Declarative controls and action metadata
│   │   ├── assets.go         # Loading local view trees
│   │   ├── media.go          # Images/GIFs
│   │   ├── style.go          # Styling
│   │   └── *_test.go         # Co-located engine regression and fuzz tests
│   └── textcell/             # Pure Unicode grapheme width, wrapping and clipping
├── terminal/                 # Optional public gradient helpers
├── cmd/sml/                  # CLI: validate and render
├── examples/                 # Markup + executable Go host demonstrations
├── docs/                     # Public docs, architecture, roadmap
└── .github/                  # CI, issue templates, and dependency updates
```

## Public API stays stable

Existing users still import:

```go
import sml "github.com/AmpedWasTaken/SML"

view, err := sml.Parse([]byte(`<screen name="hello"><text>Hello</text></screen>`))
if err != nil {
    panic(err)
}
_ = view.RenderWidth(nil, 80)
```

The root `sml.go` exposes `Node`, `View`, `Frame`, `Cell`, `Rect`, `Element`, and `Hotspot` as *type aliases*, with small delegates for public package functions. An alias preserves its original method set, fields, and type identity, rather than creating a second representation. That means `View.RenderFrame`, `View.RenderInteractive`, `View.GradientStyles` and other existing methods remain available. The CLI and examples use the same public module path.

The `internal/` packages cannot be imported by applications outside this module tree. Keep implementation details there rather than promising an unstable internal API.

## Where should new code go?

| Location | Purpose |
| --- | --- |
| `sml.go` | Public aliases/functions only; keep it small |
| `api_test.go` | External consumer compatibility and import-path tests |
| `internal/engine` | Parsing, rendering, component system, action descriptors, layout and regression tests |
| `internal/textcell` | Independent grapheme-aware terminal string algorithms |
| `terminal` | Optional, independently usable public presentation helpers |
| `cmd` | Executable tools |
| `examples` | Runnable host integrations |
| `docs` | Language semantics, architecture, guides and contribution policies |

## Additional package boundaries

Do **not** create a new directory just to hide files: every new Go package adds an import boundary. Further splitting the tightly coupled engine (e.g. parser/AST or media/render interfaces) should happen when it measurably reduces dependencies without creating cycles or leaking internal types. Keep compatibility tests and the public facade in sync whenever a split happens.

SML is a declarative view engine: the embedding host must still authorize and dispatch actions. See [architecture](architecture.md), [security](security.md) and [release process](releasing.md).
