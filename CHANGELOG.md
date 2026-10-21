# Changelog

All noteworthy user-facing changes will be recorded here.

The public API is currently pre-1.0. Follow [Semantic Versioning](https://semver.org/) once release tags are published.

## Unreleased

### Added

- Extracted the SML parser, renderer, controls, data tables, media handling and gradient helpers from Sentinel V5.
- Standalone `github.com/AmpedWasTaken/SML` module with Go examples and a CLI.
- `RenderChecked` API for propagating render errors.
- Language, component, API, architecture and security documentation.
- Regression and parser fuzz tests; GitHub Actions test, lint and CLI checks.

### Changed

- Reorganized default assets and Sentinel-specific integrations to belong to the host application rather than the language package.
- Formatted the Go source and clarified the standalone library's capabilities and limitations.
