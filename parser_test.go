package sml

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestParseAndRenderFrame(t *testing.T) {
	input := []byte("<sml><panel title=\"Service overview\"><text>Current status of the application, shown as a safe static TUI</text><metric label=\"Active\" value=\"24\"/></panel><grid columns=\"2\"><panel title=\"A\"><text>Ready</text></panel><panel title=\"B\"><text>Healthy</text></panel></grid></sml>")
	root, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := RenderString(root, Options{Width: 60})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(frame, "Service overview") || !strings.Contains(frame, "Healthy") {
		t.Errorf("unexpected frame:\n%s", frame)
	}
	for _, line := range strings.Split(strings.TrimSuffix(frame, "\n"), "\n") {
		if utf8.RuneCountInString(line) > 60 {
			t.Errorf("line exceeds 60 columns: %q", line)
		}
	}
}

func TestRejectsUnknownAndUnsafeDocumentForms(t *testing.T) {
	tests := []string{
		"<sml><exec command=\"ls\"/></sml>",
		"<!DOCTYPE sml><sml/>",
		"<sml><panel>",
		"<other><text>oops</text></other>",
		"<sml><grid columns=\"bad\"/></sml><sml/>",
		"<sml xmlns:x=\"https://example.test\"><x:text/></sml>",
	}
	for _, input := range tests {
		if _, err := Parse([]byte(input)); err == nil {
			t.Errorf("expected parse error: %s", input)
		}
	}
}

func TestRejectsInvalidColumnCountDuringRender(t *testing.T) {
	root, err := Parse([]byte("<sml><grid columns=\"5\"><text>x</text></grid></sml>"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := RenderString(root, Options{Width: 80}); err == nil {
		t.Fatal("expected grid columns error")
	}
}

func TestEscapesTerminalControlsInText(t *testing.T) {
	// XML explicitly rejects the ESC character; verify rendering also
	// scrubs control bytes in an application-constructed tree.
	root := &Node{Tag: "sml", Children: []*Node{{Tag: "text", Text: "trusted\x1b[31m message"}}}
	result, err := RenderString(root, Options{Width: 30})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "\x1b") {
		t.Fatal("terminal control escape leaked into output")
	}
}

func TestBounds(t *testing.T) {
	if _, err := Parse([]byte(strings.Repeat("x", MaxDocumentBytes+1))); err == nil {
		t.Fatal("expected size error")
	}
	root, err := Parse([]byte("<sml><text>hello</text></sml>"))
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{1, 241} {
		if _, err := RenderString(root, Options{Width: width}); err == nil {
			t.Fatalf("expected width validation for %d", width)
		}
	}
}
