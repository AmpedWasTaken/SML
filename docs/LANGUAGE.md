# SML v0 — language reference

SML is an XML-compatible markup language for declarative static terminal views.

## Grammar

Exactly one root element is required: <sml>. Elements are case-sensitive. Only the tags below are supported; arbitrary HTML, scripts, inline XML DTDs, processing instructions and namespaced elements are not supported.

- sml: root container, accepts child components.
- panel: bordered container. Optional title attribute. Accepts nested components.
- grid: column layout. Optional columns attribute (integer 1 to 4; default 2). Automatically collapses columns if the width would make individual cells narrower than 12 code points.
- text: plain textual content. Whitespace is normalized before rendering.
- metric: renders a label and value separated by flexible whitespace. Requires label and value attributes.

Panels can nest. Grids arrange their immediate children in rows. Empty panels render a blank interior row. This is a static renderer, not a full terminal event loop.

## Dynamic data

Double-brace placeholders, such as {{version}} or {{host_name}}, can appear in text or attribute values. Names must begin with an ASCII letter or underscore, followed by letters, digits, underscores, periods or hyphens.

Call Parse once, then call RenderData with Bindings for each frame. The Bind method deep-copies a parsed tree without mutating it, then substitutes values. Missing bindings return an error; each binding value is limited to 4096 UTF-8 bytes. Runtime content is treated as plain text, not parsed as SML, and output strips control characters.

## Public Go API

~~~go
root, err := sml.Parse(data)
bound, err := sml.Bind(root, sml.Bindings{"app": "My app"})
frame, err := sml.RenderString(bound, sml.Options{Width: 80})
frame, err = sml.RenderData(root, sml.Bindings{"app": "Updated"}, sml.Options{Width: 80})
err = sml.Render(os.Stdout, bound, sml.Options{Width: 80})
~~~

By default the rendering width is 80; accepted widths are 24–240 Unicode code points. The output uses Unicode box-drawing characters without emitting ANSI escape sequences.

## Input validation

- Maximum input size: 1 MiB.
- Maximum nesting: 32 element levels.
- Unknown elements, multiple root elements, XML directives, XML namespaces and malformed markup produce errors.
- The grammar does not run commands, issue network requests or evaluate arbitrary Go expressions.
- Embedded newlines and tabs are normalized in visible text.

## Known limitations

This v0 API has no layout measurements for full Unicode grapheme clusters or emoji terminal-cell widths. No mouse/keyboard navigation, focus manager, event-loop or screen diffing exists yet. It is usable for data-driven static UI frames, dashboards and read-only command output.

## Compatibility

This is a standalone library inspired by, but not source-code equivalent to, the private Sentinel V5 engine. Do not assume a view written against unpublished Sentinel internals is compatible.
