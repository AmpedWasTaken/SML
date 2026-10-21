package sml

import (
	"strings"
	"testing"
)

func TestUnicodeDisplayCells(t *testing.T) {
	for _, tc := range []struct {
		text  string
		width int
	}{
		{"A", 1}, {"界", 2}, {"e\u0301", 1}, {"👩‍💻", 2}, {"🇳🇱", 2},
	} {
		if got := CellWidth(tc.text); got != tc.width {
			t.Errorf("%q: width %d, want %d", tc.text, got, tc.width)
		}
	}
	if got := ClipCells("界AB", 3); got != "界A" {
		t.Fatalf("clip=%q", got)
	}
	if got := ClipCells("界AB", 1); got != "" {
		t.Fatalf("split wide grapheme: %q", got)
	}
	if got := PadCells("界", 4); CellWidth(got) != 4 {
		t.Fatalf("pad=%q", got)
	}
}

func TestWrapTextCombiningAndEmoji(t *testing.T) {
	lines := WrapText("Hello wide 界界 and 👩‍💻 users", 11)
	for _, line := range lines {
		if CellWidth(line) > 11 {
			t.Fatalf("overflow: %q", line)
		}
	}
	if !strings.Contains(strings.Join(lines, ""), "👩‍💻") {
		t.Fatal(lines)
	}
}
