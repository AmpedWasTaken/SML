# CLI

Use the standalone CLI without requiring a full-screen TUI host.

```sh
go run ./cmd/sml validate examples/dashboard.sml
go run ./cmd/sml render -width 70 -set service.health=online -set stats.requests=42 -set stats.users=9 -set stats.capacity=34 examples/dashboard.sml
go run ./cmd/sml validate examples/assets
go run ./cmd/sml render -assets examples/assets -set user.name=developer -set tasks.active=7 examples/assets/views/status.sml
```

Or build an executable with `go build -o sml ./cmd/sml`.

Usage: `sml <validate|render> [-assets DIR] [-width N] [-set key=value] PATH`

- `validate`: parse a `.sml` file or validate all views in an asset directory.
- `render`: print the specified file as terminal lines; no input event loop.
- `-assets DIR`: supply the root directory for reusable components, images and banners.
- `-width N`: render width between 12 and 240, default 80.
- `-set key=value`: supply a view binding; repeat for multiple values.

Place options **before** PATH. When validating a single file, external component resolution is not exhaustive unless you also validate the asset directory; for projects with includes, run `validate assets/` as part of CI.
