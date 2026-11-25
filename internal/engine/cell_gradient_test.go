package engine

import (
	"strings"
	"testing"
)

func TestDeclarativeCellGradients(t *testing.T) {
	view, err := Parse([]byte(`<screen name="demo" version="3"><panel border="round" border-gradient="#ff0000,#00ff00,#ff0000"><text foreground-gradient="#0000ff,#ffffff">Gradient title</text></panel></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 40)
	if err != nil {
		t.Fatal(err)
	}
	lines, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lines[0], "\x1b[38;2;255;0;0m") || !strings.Contains(strings.Join(lines, "\n"), "\x1b[38;2;0;0;255m") {
		t.Fatal("missing declarative gradient colors")
	}
	plain, err := PaintGlowFrame(frame, GlowOptions{TrueColor: true, NoColor: true})
	if err != nil || strings.Contains(strings.Join(plain, "\n"), "\x1b") {
		t.Fatal("plain fallback emits SGR")
	}
	if _, err := Parse([]byte(`<screen name="demo" version="3"><text foreground-gradient="#fff,unsafe">Invalid</text></screen>`)); err == nil {
		t.Fatal("invalid gradient accepted")
	}
}
