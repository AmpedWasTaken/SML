# Extracting SML from Sentinel V5

The standalone repository preserves the SML parser/renderer and most original regression tests. It deliberately excludes the Sentinel server, authentication, plugin runtime, default administration screens, provider configuration, and private application integration.

| Original | Standalone |
| --- | --- |
| `internal/sml/*.go` (engine) | Root `sml` Go package |
| `internal/terminal/gradient*.go` | Small `terminal` helper package |
| `assets/views/*.sml` | Host-owned assets; examples under `examples/` |
| Sentinel TUI event handling | Host responsibility |
| Authenticated server bindings | Host responsibility |
| `ui.publish` / session invalidation | Host responsibility |

## Host migration checklist

1. Add `github.com/AmpedWasTaken/SML` to your module dependencies once released.
2. Replace `github.com/AmpedWasTaken/Sentinel-V5/internal/sml` imports with the standalone package.
3. Update gradient type imports to `github.com/AmpedWasTaken/SML/terminal`.
4. Keep `assets/views`, `assets/components`, `assets/art`, and `assets/media` with the embedding application.
5. Preserve your existing server-side permission and role checks for every action.
6. Run the original application test suite and check visual output on terminals with narrow widths, styles, mouse, and GIF fallback.
7. Pin a commit or release tag for reproducible builds before switching production deployments.

The standalone project does **not** automatically modify Sentinel V5. Make the dependency switch in the main application only after verifying API equivalence and CI.
