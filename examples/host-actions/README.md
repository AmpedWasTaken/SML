# Host-owned actions

Run `go run ./examples/host-actions` from the repository root.

This demo parses a view, binds application data, renders a cell-based `Frame`, and prints descriptors for controls. **No click handler or command is executed.** Your real TUI host maps terminal input to returned `Hotspot` coordinates and checks authorization before acting on `Action`.

For the complete API, see [docs/go-api.md](../../docs/go-api.md).
