# Release process

SML is currently pre-1.0. Do not call an untagged `main` build a stable release.

## Versioning

Use semantic versions for tagged releases. Before v1.0, document compatibility changes explicitly; after v1.0, reserve breaking exported Go API changes for a major version and follow Go module semantic import versioning.

- Patch: bug fix without intended API/markup breaking changes.
- Minor (pre-1.0): additions and explicitly documented compatibility changes.
- Major: breaking API contract that requires a new Go module import path when applicable.

Pin a tag or commit for reproducible production builds. Avoid claiming full SemVer guarantees before a documented stable release.

## Maintainer checklist

1. Review merged changes and ensure `CHANGELOG.md` describes user-visible changes and compatibility.
2. Run CI on Go 1.23 and current stable Go; verify `go test ./...`, `go vet ./...`, race tests, and CLI examples.
3. Check security-sensitive changes, API diffs, docs, and migration notes.
4. Choose a version number and create an annotated Git tag from the intended release commit.
5. Publish release notes with install instructions, notable changes, and migration warnings.
6. Verify package documentation resolves on [pkg.go.dev](https://pkg.go.dev/github.com/AmpedWasTaken/SML) and that a fresh consumer can `go get` the tag.

Tagging and publishing are **manual maintainer decisions**; this document does not imply a release has been made.
