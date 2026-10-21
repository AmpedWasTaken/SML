# Contributing to SML

Thank you for helping make declarative terminal UIs useful beyond any single application. We welcome bug fixes, documentation, examples, performance improvements, and well-scoped feature proposals.

Please read the [Code of Conduct](CODE_OF_CONDUCT.md) and [Security Policy](SECURITY.md). Security reports containing exploit details must not be filed as public issues.

## Find a good first contribution

- Browse the [roadmap](docs/roadmap.md), issues, and open pull requests.
- A small, reproducible bug or documentation improvement can be submitted directly.
- For syntax, architecture, exported Go API changes, or breaking behavior, open a feature request **before** implementation.
- Avoid application-specific commands, permission assumptions, vendor dependencies, or executable directives in SML markup.

## Local setup

Go 1.23 or newer is required.

```bash
git clone https://github.com/AmpedWasTaken/SML.git
cd SML
go mod download
make check
go test -race ./...
```

Without `make`, use `gofmt -l` to check formatting, `go test ./...`, `go vet ./...`, and the example commands in the [README](README.md).

## Code organization

See [project structure](docs/project-structure.md). Maintain exported Go APIs in the root `sml.go` facade (aliases and wrappers). Put the parser, renderer, components, controls, assets and their package-private tests under `internal/engine/`. Pure terminal grapheme helpers remain in `internal/textcell/`. Optional public gradient helpers belong in `terminal/`, examples in `examples/`, and executable tools in `cmd/`. Update root facade tests when changing public behavior.

Keep changes focused. Do not create a folder per root file: in Go, package boundaries are directories, and moving a public declaration to a new package changes import compatibility.

## Tests and documentation

- Add tests for new behavior, edge cases, malformed SML, and previously reported regressions.
- For layout changes, include narrow/wide terminal widths, ANSI-free output, combining marks, emoji, CJK characters, and action hitboxes when applicable.
- For markup syntax changes, test parsing, validation, rendering, error messages, and version compatibility.
- Update the relevant `docs/` page and provide a runnable `.sml` or Go example.
- Keep resource limits and untrusted-input boundaries intact. The host owns dispatch and authorization; SML never executes actions itself.
- Add a note to `CHANGELOG.md` for user-visible changes.

## Pull requests

1. Fork the repository and create a descriptive branch from `main`.
2. Make one focused change and use `gofmt` on changed Go files.
3. Run `make check` and `go test -race ./...` when possible.
4. Open a pull request and fill in the template, including compatibility impact and tests.
5. Address review feedback; maintainers decide on release timing and API stability.

We prefer small PRs that are easy to review, but cross-cutting refactors are welcome with a clear migration plan. Do not commit local build outputs, generated debug files, or secrets.

## Compatibility

SML is pre-1.0. Existing v1/v2/v3 markup should keep working unless a change is explicitly reviewed as breaking. Public API changes need documented migration paths. See [release guidelines](docs/releasing.md).

## License

Contributions are licensed under the repository's [MIT license](LICENSE). Submit only work you have the right to license.
