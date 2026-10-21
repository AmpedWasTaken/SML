# Security model

SML is a **declarative presentation language**, not a sandboxed shell or scripting language. The parser rejects arbitrary XML directives, processing instructions, unknown elements, and invalid attribute combinations. It enforces a 64 KiB view size limit, maximum parsing depth and node limits, bounded list/repeat rendering, and output sanitization for dynamic text.

## Trust boundaries

- **Markup** describes layout, reactive display-only derivations and action metadata.
- **Bindings** are strings provided by the host; only pass authorized data.
- **Actions** must be checked and executed by the embedding application. Never treat `enabled-when`, `page`, or action metadata as authorization.
- **Assets** are read from the local asset root; validate the asset directory and manage it as application-controlled content. The library does not download remote media.
- **Rendering** returns printable lines and hotspots, but the embedding app owns any actual terminal I/O and interaction dispatch.

SML has no arbitrary script interpreter, external URL fetching, shell execution, general expression evaluation, or Go code evaluation.

## Practical requirements for hosts

1. Treat editor-supplied markup as untrusted until `Parse`/`LoadDir` succeeds.
2. Keep permissions and role checks in the Go host. Validate *again* immediately before executing any action.
3. Decide which binding namespaces are available to a user/session. Do not expose secrets in `values`.
4. Keep asset directories private to the intended application and consider symlink/TOCTOU defenses appropriate to your deployment.
5. If a hot reload fails, preserve the previously validated view set.
6. Verify terminal output and interaction under the clients you support. The engine does not include its own SSH server or terminal driver.

## Compatibility warning

The extracted engine retains legacy read-only command names and typed plugin-action metadata originating from Sentinel V5. They are presentation contracts only: the standalone library **does not** provide the corresponding services or authorize their dispatch.


## Generic actions and components

`action="app.refresh"` is a validated **descriptor**, never executable text. The embedding host must authorize and dispatch it. `id` is a stable UI identifier, not a capability token. Legacy command allowlisting continues only for backwards compatibility.

Custom component names and prop declarations are validated, and component files are local-only and depth-bounded. Props are string interpolation, not Go expressions or shell commands. Reusing a component does not grant access to any additional binding.
