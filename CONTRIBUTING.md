# Contributing to SML

Thanks for helping make the standalone SML library useful beyond its original host.

## Local checks

Use Go 1.23+.

```sh
go fmt ./...
go test ./...
go vet ./...
go run ./cmd/sml validate examples/dashboard.sml
go run ./cmd/sml validate examples/assets
```

Keep the public API small. Submit parsing and rendering changes with regression tests, including malformed-input cases where appropriate. Please do not add shell execution, unchecked dynamic evaluation, or unbounded network/filesystem access to markup features.

## Pull requests

Explain the behavior change, provide a minimal `.sml` example, update the reference docs, and describe backwards-compatibility impact. Preserve v1/v2/v3 markup compatibility when possible. Avoid coupling the library to Sentinel's server, plugin manager, or authentication system.

## Licensing

No open-source license has been selected at this stage. Contributions and redistribution terms will need to be clarified before a stable open-source release.
