<div align="center">

# SML

### Beautiful terminal UIs, written in markup.

**A declarative, lightweight TUI view engine for Go.**  
Compose terminal screens with familiar XML-like elements, supply live data from Go, and keep presentation separate from your application logic.

[Get started](#quick-start) · [Language guide](docs/language.md) · [Go API](docs/go-api.md) · [Components](docs/components.md) · [Examples](examples/) · [Contributing](CONTRIBUTING.md) · [Roadmap](docs/roadmap.md)

[![Go CI](https://github.com/AmpedWasTaken/SML/actions/workflows/go.yml/badge.svg)](https://github.com/AmpedWasTaken/SML/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/AmpedWasTaken/SML.svg)](https://pkg.go.dev/github.com/AmpedWasTaken/SML)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.23+](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

SML lets you define *what* your terminal UI looks like, without hardcoding layouts into hundreds of Go print statements. Its markup is declarative, bounded, and deliberately **non-executable**.

**What it looks like:**

```xml
<screen name="dashboard" version="3">
  <derive key="derived.health"
          op="coalesce"
          source="app.health"
          fallback="unknown"/>

  <section title="DASHBOARD">
    <grid columns="2" gap="2" collapse-below="60">
      <metric label="Requests" value="{{app.requests}}"/>
      <metric label="Errors" value="{{app.errors}}"/>
    </grid>

    <panel title="Service">
      <field label="Status" key="derived.health"/>
      <progress label="CPU" value="{{app.cpu}}"/>
    </panel>
  </section>
</screen>
```

The embedding Go application supplies the `app.*` bindings. Change the data and re-render the view; no embedded scripting or runtime JavaScript is involved.

## Why SML?

| | |
| --- | --- |
| **Declarative UI** | Write screens using `<row>`, `<grid>`, `<panel>`, `<metric>`, and other readable components. |
| **Go-first integration** | Parse markup, supply bindings, and render terminal lines or structured Unicode cell frames. |
| **Reactive-style views** | Recompute bounded `<derive>` values and conditional branches when your host renders updated data. |
| **Reusable components** | Compose with `<include>` or `<use>` using declared props and named slots. |
| **Interactive metadata** | Get cell-accurate hitboxes, stable control IDs and host-owned action descriptors. |
| **Safer by design** | Reject unknown elements and executable directives; limit input size, nesting, repetition and media. |
| **Small dependency footprint** | Uses `uniseg` for correct Unicode grapheme widths; the rest is Go standard library. |

**How it fits into the Go TUI ecosystem:** SML is a *view layer*, not a replacement for a terminal event loop. You can use it inside an SSH dashboard, a custom terminal shell, or alongside another Go event/input system. Terminal sessions, redraw scheduling, navigation, state persistence, and permissions remain the responsibility of your host.

## Quick start

Requires **Go 1.23+**.

```bash
go get github.com/AmpedWasTaken/SML
```

Use SML directly:

```go
package main

import (
    "fmt"
    "strings"

    "github.com/AmpedWasTaken/SML"
)

func main() {
    view, err := sml.Parse([]byte(`
        <screen name="hello">
            <panel title="Hello, SML!">
                <text>Welcome, {{user.name}}</text>
            </panel>
        </screen>
    `))
    if err != nil {
        panic(err)
    }

    lines, err := view.RenderChecked(
        map[string]string{"user.name": "developer"}, 60,
    )
    if err != nil {
        panic(err)
    }
    fmt.Println(strings.Join(lines, "\n"))
}
```

Or run the included CLI from a clone:

```bash
go run ./cmd/sml validate examples/dashboard.sml

go run ./cmd/sml render \
  -width 70 \
  -set app.health=online \
  -set app.requests=4242 \
  -set app.errors=3 \
  -set app.cpu=37 \
  examples/quickstart.sml
```

> The CLI prints rendered lines. It does **not** start a mouse-driven, full-screen interactive application. The host is responsible for events and redrawing.

## Explore the language

| Feature | Example |
| --- | --- |
| Layout | `<row>`, `<column>`, `<grid columns="3">` |
| Sections | `<panel title="Stats">`, `<section title="Overview">` |
| Data | `<metric>`, `<field>`, `<table>`, `<progress>`, `<gauge>`, `<sparkline>` |
| Logic | `<if>`, `<when>`, `<otherwise>`, `<empty>`, `<repeat>` |
| Custom components | `<use>`, `<slot>`, `<fill>` with string props |
| Computed values | `<derive key="derived.status" op="coalesce" .../>` |
| Components | `<include name="header"/>` |
| Inputs and actions | `<button>`, `<link>`, `<choice>`, `<toggle>`, `<input>`, `<number>` |
| Visuals | Inline styling attributes, gradients, text-cell images and GIFs |

SML views are **strictly validated**, not HTML documents; browser APIs, CSS stylesheets, arbitrary JS, shell invocations and remote image URLs are not supported.

### Your first real project

```text
my-app/
├── main.go
└── assets/
    ├── views/
    │   ├── home.sml
    │   └── status.sml
    ├── components/
    │   └── header.sml
    ├── art/                    # optional ASCII art
    └── media/                  # optional local images/GIFs
```

```go
views, err := sml.LoadDir("assets")
if err != nil {
    return err // keep last-known-good assets when hot reloading
}

lines, err := views["home"].RenderChecked(data, 80)
if err != nil {
    return err
}

// Your application decides how and when to paint these lines.
```

See [the Go API guide](docs/go-api.md), [complete component reference](docs/components.md), [example assets](examples/assets/) and [host-owned action example](examples/host-actions/).

## CLI and development

```bash
go run ./cmd/sml validate examples/dashboard.sml
go run ./cmd/sml validate examples/assets
go run ./cmd/sml render -width 80 -set user.name=Tim -assets examples/assets examples/assets/views/status.sml
go test ./...
go vet ./...
```

For commonly used commands, run `make check` (when Make is available). The CLI reference explains `validate`, `render`, flags and error codes: [docs/cli.md](docs/cli.md).

## Documentation

| Guide | Covers |
| --- | --- |
| [Getting started / syntax](docs/language.md) | Screens, bindings, conditions, includes, and versions |
| [Component reference](docs/components.md) | Built-in elements by category with examples |
| [Go integration](docs/go-api.md) | Parsing, rendering, interaction metadata, and errors |
| [Cell layout](docs/layout.md) | Unicode widths, frames, bounds and stable control IDs |
| [CLI](docs/cli.md) | Commands, arguments and asset loading |
| [Architecture](docs/architecture.md) | Parser, validator, renderer, host integration, limitations |
| [Project structure](docs/project-structure.md) | Package boundaries, extension points, file placement |
| [Roadmap](docs/roadmap.md) | Proposed improvements and contribution priorities |
| [Releasing](docs/releasing.md) | Versioning, compatibility, release checklist |
| [Security](docs/security.md) | Resource limits, action trust boundaries and safe embedding |
| [Migrating from Sentinel V5](docs/migration.md) | Package imports and host responsibilities |
| [Contributing](CONTRIBUTING.md) | Local tooling, tests and PRs |

## Status and compatibility

SML was originally developed as the terminal view engine for [Sentinel V5](https://github.com/AmpedWasTaken/Sentinel-V5) and has been extracted so anyone can embed the language in Go applications. The engine accepts SML v1/v2/v3 documents and preserves the existing markup syntax. **Compatibility is the goal, not a promise of a frozen v1.0 API**. Pin a version/commit for production integrations until the first stable release.

The project includes Go tests and GitHub CI. The published package does not contain Sentinel authentication, providers, plugin execution, or backend permission logic; it only describes terminal views and interactions.

## Contributing

Issues, examples, bug reports, and PRs are welcome. Please include a small reproducer for rendering bugs, test both narrow and wide terminals, and document any language-level changes.

Start with [CONTRIBUTING.md](CONTRIBUTING.md), report bugs using [issue templates](.github/ISSUE_TEMPLATE/), and see [SUPPORT.md](SUPPORT.md) for help. Read [the changelog](CHANGELOG.md), [project structure](docs/project-structure.md), [security policy](SECURITY.md), and [code of conduct](CODE_OF_CONDUCT.md) before contributing.

## License

MIT © 2026 [AmpedWasTaken](https://github.com/AmpedWasTaken). See [LICENSE](LICENSE).
