# Contributing

Thanks for helping build a reusable Go-first terminal markup language.

## Before opening a PR

1. Read [the architecture](docs/architecture.md), [the language guide](docs/language.md), and [the security model](docs/security.md).
2. Open an issue for major new syntax or public API changes so backward compatibility can be discussed.
3. Prefer small, focused PRs with representative SML snippets and tests.
4. Update examples and documentation when changing markup semantics.

## Development

Requires Go 1.23+.

```bash
git clone https://github.com/AmpedWasTaken/SML.git
cd SML
gofmt -w .
go test ./...
go vet ./...
go run ./cmd/sml validate examples/assets
```

Or use `make check` to run formatting checks, tests, vet and CLI examples.

## Changes we value

- Better terminal-width handling, with edge-case tests.
- Clear error messages and diagnostics for malformed SML.
- Performance improvements that preserve bounded resource use.
- Interop examples with real Go TUI input/event systems.
- Accessible, narrow-terminal layouts and complete docs.

SML markup is **not** an executable scripting language. Avoid shell execution, arbitrary eval, unrestricted file paths, user-controlled network access or permissions encoded in markup.

## PR checklist

- [ ] Go is `gofmt`-formatted.
- [ ] `go test ./...` and `go vet ./...` pass.
- [ ] Invalid input and security-sensitive changes have tests.
- [ ] New behavior is documented with an actual `.sml` example.
- [ ] v1/v2/v3 compatibility impact is stated.

## License

By contributing, you agree to release your contributions under the [MIT License](LICENSE).
