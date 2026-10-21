# SML — Structured Markup for Terminal Interfaces

A **dependency-free declarative terminal UI library for Go**. Parse small XML-compatible SML views into validated component trees and render them with live application data.

[![Go CI](https://github.com/AmpedWasTaken/SML/actions/workflows/go.yml/badge.svg)](https://github.com/AmpedWasTaken/SML/actions/workflows/go.yml)

## Installation

~~~sh
go get github.com/AmpedWasTaken/SML
~~~

SML is an evolving v0 Go package; pin a released tag or commit for reproducibility.

## Quickstart

Define a view:

~~~xml
<sml>
  <panel title="{{app}}">
    <text>Live service overview</text>
    <grid columns="2">
      <metric label="Jobs" value="{{jobs}}"/>
      <metric label="Healthy" value="{{healthy}}"/>
    </grid>
  </panel>
</sml>
~~~

Render it with dynamic values:

~~~go
package main

import (
    "fmt"
    sml "github.com/AmpedWasTaken/SML"
)

func main() {
    source := []byte("<sml><panel title=\"{{app}}\"><metric label=\"Jobs\" value=\"{{jobs}}\"/></panel></sml>")
    root, err := sml.Parse(source)
    if err != nil {
        panic(err)
    }
    frame, err := sml.RenderData(root, sml.Bindings{
        "app": "Example service",
        "jobs": "42",
    }, sml.Options{Width: 64})
    if err != nil {
        panic(err)
    }
    fmt.Print(frame)
}
~~~

Run the sample in this repository:

~~~sh
go run ./examples/basic
~~~

## Features

- Real, installable Go package; no third-party dependencies.
- Strict XML-compatible parser with limits on source size and tree depth.
- Native panel, grid, text and metric components.
- Nested panel rendering and 1–4 column grids.
- Runtime data bindings using double-brace placeholders in text and attributes.
- Reusable parsed document trees; binding does not mutate the original.
- Plain-text output to io.Writer or string, without terminal escape sequences.
- Automated unit tests, Go vet and runnable example.

## Public API

| API | Purpose |
|---|---|
| Parse(source) | Parse and validate SML syntax |
| Bind(root, values) | Clone and resolve placeholders |
| Render(writer, root, options) | Render a frame to an output writer |
| RenderString(root, options) | Render a frame as text |
| RenderData(root, values, options) | Bind and render using runtime data |

See [language documentation](docs/LANGUAGE.md) and [contributing guide](CONTRIBUTING.md).

## Current limitations

The first module release focuses on **rendering** rather than terminal event processing. Keyboard focus, keybindings, mouse events, scroll virtualization, interactive commands, full-screen updates, ANSI themes and cross-platform input drivers are not implemented. Terminal width currently measures Unicode code points; double-width emoji and some combining characters may produce imperfect alignment.

This library is a standalone, independently maintained core inspired by the SML approach in Sentinel V5. It does **not** claim source-code parity with Sentinel V5's private implementation.

## Safety

- SML expressions are plain-text substitutions, not code execution.
- DTDs, XML processing instructions, namespaces and unknown elements are rejected.
- Source documents are capped at 1 MiB and 32 nesting levels.
- Control characters are removed during rendering; binding cannot inject new markup.

## Develop

~~~sh
go test ./...
go vet ./...
go run ./examples/basic
~~~

Licensed under MIT. See [LICENSE](LICENSE).
