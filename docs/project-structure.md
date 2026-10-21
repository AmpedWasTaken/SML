# Project structure

```text
SML/
├── .github/
│   ├── ISSUE_TEMPLATE/      # Structured, reproducible reports and requests
│   └── workflows/           # Go compatibility tests and checks
├── cmd/sml/                 # Standalone validate/render CLI
├── docs/                    # Language, APIs, architecture, roadmap, releases
├── examples/                # Markup assets and runnable Go demonstrations
├── internal/
│   └── textcell/            # Private grapheme width, clipping and wrapping
├── terminal/                # Public, optional terminal gradient helpers
├── *.go                     # Public sml package and tightly coupled engine
├── go.mod / go.sum           # Module and dependency pin
├── Makefile                 # Contributor checks
└── README.md                # Entry point
```

## Why not move every root Go file into its own folder?

In Go, a directory is a package. The public import `github.com/AmpedWasTaken/SML` exposes types such as `Node`, `View`, `Frame`, `Cell`, and `Hotspot`. Moving those declarations into independent folders would create new import paths, circular dependencies, or compatibility-breaking API changes.

Root source files remain grouped by responsibility (`layout.go`, `interaction.go`, `components.go`, `media.go`, `style.go`, etc.) while reusable implementation without root dependencies belongs in `internal/`. Small public delegating wrappers, such as `cells.go`, preserve older imports. This is intentional Go package design, not an unfinished folder migration.

## Responsibilities

| Location | What goes here | Public import |
| --- | --- | --- |
| Root `sml` | Parsing, validation, view tree, render APIs, interaction descriptors | Yes |
| `internal/textcell` | Unicode text measurement and layout primitives | No |
| `terminal` | Optional RGB/gradient helper utilities | Yes |
| `cmd/sml` | CLI binary only | Command |
| `examples` | Minimal host integration and portable markup | For learning |
| `docs` | Reference, guidance, contribution decisions | N/A |

## Extension rules

- Keep SML a *view engine*. Do not bind its core to an event loop, renderer backend, auth system, shell, or Sentinel-specific provider.
- Add new grammar and semantics to the root validation/rendering pipeline with tests.
- Keep pure reusable helpers in `internal/` if doing so avoids cycles. Move public declarations only when a release explicitly documents an API break.
- Maintain `RenderChecked`, `RenderFrame`, and `RenderInteractive` compatibility for current callers.
- Prefer `action="host.action"` metadata over legacy `command`. The host must authorize and dispatch actions.
- Add new examples in their own folders and keep imports reproducible with `go run ./examples/<name>`.

See [architecture](architecture.md), [security](security.md), and [releases](releasing.md).
