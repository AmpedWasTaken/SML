# Declarative framed layouts

The `box` element constrains its children to a width of 12–240 terminal
cells (automatically shrinking to the available width) and positions them
left, center or right. Its contents are real SML nodes, not host-rendered rows.

An `input appearance="boxed" border="round"` is a real three-row field.
Its hotspot spans the interior of the middle row; the host still owns input
editing, masking, and validation. The outer field width comes from its parent
box. A host should use the interior start plus the displayed value's cell
width for caret placement, clamped to the visible interior.

`foreground`, `background`, and `border-color` accept literal RGB hex values
on supported visual elements. The structural renderer remains plain and
the optional PaintGlowFrame pass paints trusted ANSI SGR only. NoColor
continues to return unstyled content. Conditions and design tokens can
select focus styles without changing actions.

See examples/framed-login.sml for an entirely declarative layout. The
example is an engine demonstration, not a deployed Sentinel login. Modal
layering, perimeter gradients, exact legacy snapshot parity, host caret
integration and older-client profile selection require further work.

A `foreground-gradient` or `border-gradient` takes two to eight comma-separated
hex colors. Foreground gradients follow horizontal cell position. Border
gradients follow the rectangle perimeter clockwise, allowing a continuous
closed loop when the first and last color match. See
`examples/language-dialog.sml` for a compact conditional dialog body.
The host must still composite it over the login background; that overlay
is not implemented by this example.
