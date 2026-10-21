# SML

**Declarative terminal interfaces in Go.** SML is a small XML-like markup language and rendering library extracted from the SML engine in Sentinel V5.

[![Go CI](https://github.com/AmpedWasTaken/SML/actions/workflows/go.yml/badge.svg)](https://github.com/AmpedWasTaken/SML/actions/workflows/go.yml)

Define layouts in editable `.sml` files, bind host-provided data, compose reusable components, and render terminal-friendly text. The library is intentionally separate from databases, networking, authentication, and any particular terminal server.

> **Status: pre-release extraction.** The first standalone version is under review. There is no tagged stable release yet. This is a markup parser and renderer, **not** an independent full-screen input/event loop.

## A first view

```xml
<screen name="dashboard" version="3">
  <derive key="derived.health" op="coalesce"
          source="service.health" fallback="unknown"/>
  <section title="Overview">
    <field label="Health" key="derived.health"/>
    <grid columns="2" collapse-below="60">
      <metric label="Requests" value="{{stats.requests}}"/>
      <metric label="Users" value="{{stats.users}}"/>
    </grid>
    <progress label="Capacity" value="{{stats.capacity}}"/>
  </section>
</screen>
```

## Quick start

Requires Go 1.23+.

```sh
go get github.com/AmpedWasTaken/SML
go run github.com/AmpedWasTaken/SML/cmd/sml validate examples/dashboard.sml
```

The `go run` command above assumes you are inside this repository with access to its `examples/` folder; in another project, validate your own file path.

Or from a clone:

```sh
go run ./cmd/sml validate examples/dashboard.sml
go run ./cmd/sml render -width 70 -set service.health=online -set stats.requests=42 -set stats.users=9 -set stats.capacity=34 examples/dashboard.sml
```

Embed it in a Go application:

```go
package main

import (
    "fmt"
    "strings"

    "github.com/AmpedWasTaken/SML"
)

func main() {
    v, err := sml.Parse([]byte(`<screen name="demo"><text>Hello {{user.name}}</text></screen>`))
    if err != nil { panic(err) }
    fmt.Println(strings.Join(v.RenderWidth(map[string]string{"user.name": "developer"}, 80), "\n"))
}
```

## What is included

- **Markup:** `screen`, `text`, `row`, `column`, `grid`, `panel`, `section`, `metric`, `badge`, `progress`, `table`, `if`, `when`, `repeat`, and more.
- **SML v3:** bounded `derive` computations, conditional fallbacks, style attributes, and state-gated controls.
- **Presentation:** ASCII/Unicode tables, gauges, sparklines, text-based image/GIF rendering, and gradient profiles.
- **Interaction metadata:** `RenderInteractive` returns view lines and hit-test metadata; the embedding application owns keyboard/mouse routing, permissions, and action execution.
- **Assets:** `LoadDir` reads `views/*.sml`, and `include` resolves `components/*.sml` with bounded validation.
- **Developer tools:** `sml validate` and `sml render`, sample views, regression tests, and CI.

## Documentation

| Guide | Description |
| --- | --- |
| [Language reference](docs/language.md) | Syntax, elements, expressions, files and versions |
| [Go integration](docs/go-api.md) | Parsing, loading, rendering and interactive metadata |
| [CLI](docs/cli.md) | Validation and rendering commands |
| [Security model](docs/security.md) | Limits, trust boundaries, and host authorization |
| [Sentinel migration](docs/migration.md) | Compatibility boundaries and integration checklist |
| [Contributing](CONTRIBUTING.md) | Development workflow and pull requests |

## Important boundaries

SML markup **cannot execute arbitrary Go, shell commands, JavaScript, or network requests**. Action declarations only describe controls. Rendered output and control metadata are returned to the embedding host; the host must authorize and dispatch any requested action. The `enabled-when` property only controls presentation, not permissions.

Existing Sentinel V5 v1/v2/v3 SML documents are the compatibility target. Server bindings, session state persistence, plugins, terminal resize/mouse support, and live invalidation are **host responsibilities**, not automatically provided by this library.

## Project history

The initial implementation originates from Sentinel V5's `internal/sml` package and its terminal gradient utilities. The reusable parser, renderer and engine tests are extracted here; Sentinel's server runtime, bundled administration views, and application-specific defaults are intentionally not included.

**License:** not selected yet. Until a license is added, standard copyright restrictions apply; check with the maintainer before redistributing.

**Maintainer:** [AmpedWasTaken](https://github.com/AmpedWasTaken)
