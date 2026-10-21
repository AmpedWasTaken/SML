# Contributing to SML

Issues and pull requests are welcome. Please keep changes focused on portable, testable Go UI primitives.

## Local checks

~~~sh
go test ./...
go vet ./...
go run ./examples/basic
~~~

For parser changes, add tests covering valid and malformed input. For rendering changes, include snapshots or exact expected layout checks for short and wide terminal widths. Avoid adding runtime code execution or networking as a feature of the markup language.

## API stability

The module has a v0 evolving API. Explain any compatibility break in the pull request and update README and docs/LANGUAGE.md. Code style follows gofmt.

## Security reports

For security-sensitive issues, use GitHub private vulnerability reporting instead of publishing exploit details.
