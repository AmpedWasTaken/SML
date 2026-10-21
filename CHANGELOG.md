# Changelog

All noteworthy user-facing changes will be recorded here.

The public API is currently pre-1.0. Follow [Semantic Versioning](https://semver.org/) once release tags are published.

## Unreleased

### Added

- GitHub community issue and pull request templates, support and security policies, and documented release process.
- A host-owned action example and private `internal/textcell` Unicode text primitives.
- A Go version CI matrix with race testing for the supported environment.

- Extracted the SML parser, renderer, controls, data tables, media handling and gradient helpers from Sentinel V5.
- Standalone `github.com/AmpedWasTaken/SML` module with Go examples and a CLI.
- `RenderChecked` API for propagating render errors.
- Language, component, API, architecture and security documentation.
- Regression and parser fuzz tests; GitHub Actions test, lint and CLI checks.

### Changed

- Relocated core engine source and all package-private regression tests to `internal/engine/`, keeping the repository root focused on a single public `sml.go` facade.
- Re-exported existing public types and functions through Go aliases/wrappers without introducing new imports for consumers; added external-package API compatibility regression coverage.

- Kept the public cell helpers as backward-compatible wrappers while moving implementation into `internal/textcell`.
- Isolated Sentinel-origin legacy command names without altering the existing parser behavior.

- Reorganized default assets and Sentinel-specific integrations to belong to the host application rather than the language package.
- Formatted the Go source and clarified the standalone library's capabilities and limitations.
