package textcell

import "testing"

func TestGraphemeWidthAndClipping(t *testing.T) {
	if got := CellWidth("A👩‍💻界"); got != 5 {
		t.Fatalf("display width = %d, want 5", got)
	}
	if got := ClipCells("A👩‍💻界", 2); got != "A" {
		t.Fatalf("clip split wide grapheme: %q", got)
	}
	if got := ClipCells("A👩‍💻界", 3); got != "A👩‍💻" {
		t.Fatalf("clipped string = %q", got)
	}
	if got := PadCells("界", 4); got != "界  " {
		t.Fatalf("padded string = %q", got)
	}
}

func TestWrappedLinesRespectCellWidth(t *testing.T) {
	for _, line := range WrapText("Hello 界 👩‍💻 world", 6) {
		if got := CellWidth(line); got > 6 {
			t.Fatalf("line %q exceeds requested cell width: %d", line, got)
		}
	}
}
