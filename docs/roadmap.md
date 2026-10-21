# Roadmap

This is a proposal list, not a claim that every feature already works. Please discuss major syntax/API changes before implementation.

## Foundation (implemented)

- Standalone Go module, renderer and CLI.
- Cell-based layout, grapheme-aware text, stable action IDs, reusable components.
- Documentation and parser/rendering regression tests.

## Next: reliability and developer experience

- Add more golden render snapshots across narrow/wide layouts and supported terminals.
- Improve actionable parser diagnostics (file, line/column, attribute).
- Benchmark repeated rendering, component expansion, and wide Unicode content.
- Publish a realistic integration example using a mainstream Go TUI host.
- Improve documentation search/navigation and the language reference.

## Later: extensibility

- Discuss a versioned extension API for custom, host-rendered components.
- Explore more accessible keyboard/focus metadata without owning host input.
- Evaluate rendering performance and allocation budgets for frequently refreshed UIs.
- Document a compatibility policy before a 1.0 release.

Each roadmap item needs a minimal problem statement, tests, docs, and a decision about how its functionality stays safe and general-purpose. See [CONTRIBUTING.md](../CONTRIBUTING.md).
