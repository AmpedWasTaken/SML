package engine

import (
	"reflect"
	"testing"
)

func diffFrame(text string, width int) Frame {
	cells := make([]Cell, width)
	for i, r := range []rune(text) {
		if i >= width {
			break
		}
		cells[i] = Cell{Grapheme: string(r)}
	}
	return Frame{Width: width, Height: 1, Cells: [][]Cell{cells}, Lines: []string{text}}
}

func TestDiffFramesMinimalCellChangeAndEmpty(t *testing.T) {
	before := diffFrame("abcde", 8)
	after := diffFrame("abZde", 8)
	diff, err := DiffFrames(before, after)
	if err != nil || diff.FullRedraw || len(diff.Runs) != 1 || diff.Runs[0].Row != 0 || diff.Runs[0].Start != 2 || len(diff.Runs[0].Cells) != 1 || diff.Runs[0].Cells[0].Grapheme != "Z" {
		t.Fatalf("unexpected run %+v err=%v", diff, err)
	}
	if len(after.Cells[0]) != 8 || before.Cells[0][2].Grapheme != "c" {
		t.Fatal("source frame changed")
	}
	same, err := DiffFrames(after, after)
	if err != nil || len(same.Runs) != 0 || same.FullRedraw {
		t.Fatalf("identical %+v %v", same, err)
	}
}

func TestDiffFramesWideGraphemeOldNewAndClears(t *testing.T) {
	old := Frame{Width: 4, Height: 1, Cells: [][]Cell{{{Grapheme: "A"}, {Grapheme: "界"}, {Continuation: true}, {Grapheme: "Z"}}}}
	next := Frame{Width: 4, Height: 1, Cells: [][]Cell{{{Grapheme: "A"}, {Grapheme: "X"}, {Grapheme: "Y"}, {Grapheme: "Z"}}}}
	diff, err := DiffFrames(old, next)
	if err != nil || len(diff.Runs) != 1 || diff.Runs[0].Start != 1 || len(diff.Runs[0].Cells) != 2 {
		t.Fatalf("wide to narrow %+v %v", diff, err)
	}
	back, err := DiffFrames(next, old)
	if err != nil || len(back.Runs) != 1 || back.Runs[0].Start != 1 || len(back.Runs[0].Cells) != 2 || !back.Runs[0].Cells[1].Continuation {
		t.Fatalf("narrow to wide %+v %v", back, err)
	}
	cleared := Frame{Width: 4, Height: 1, Cells: [][]Cell{{{Grapheme: "A"}, {}, {}, {Grapheme: "Z"}}}}
	diff, err = DiffFrames(old, cleared)
	if err != nil || len(diff.Runs) != 1 || diff.Runs[0].Start != 1 || len(diff.Runs[0].Cells) != 2 || diff.Runs[0].Cells[0].Grapheme != "" {
		t.Fatalf("clear %+v %v", diff, err)
	}
}

func TestDiffFramesResizedAndNewFrame(t *testing.T) {
	before := diffFrame("Hi", 4)
	after := diffFrame("Hi", 6)
	diff, err := DiffFrames(before, after)
	if err != nil || !diff.FullRedraw || len(diff.Runs) != 1 || len(diff.Runs[0].Cells) != 6 {
		t.Fatalf("resize %+v %v", diff, err)
	}
	blank, err := DiffFrames(Frame{}, after)
	if err != nil || !blank.FullRedraw || !reflect.DeepEqual(diff.Runs, blank.Runs) {
		t.Fatalf("first paint %+v %v", blank, err)
	}
}

func TestDiffFramesMetadataOnlyAndStyles(t *testing.T) {
	before := diffFrame("ok", 5)
	after := diffFrame("ok", 5)
	before.Hotspots = []Hotspot{{ID: "a", Action: "one", Line: 0, Start: 0, End: 2}}
	after.Hotspots = []Hotspot{{ID: "a", Action: "two", Line: 0, Start: 0, End: 2}}
	after.Elements = []Element{{ID: "a", Bounds: Rect{Width: 2, Height: 1}}}
	diff, err := DiffFrames(before, after)
	if err != nil || len(diff.Runs) != 0 || !diff.HotspotsChanged || !diff.ElementsChanged {
		t.Fatalf("metadata-only %+v %v", diff, err)
	}
	before.Lines = []string{"\x1b[31mok\x1b[0m"}
	after.Lines = []string{"\x1b[32mok\x1b[0m"}
	diff, err = DiffFrames(before, after)
	if err != nil || !diff.FullRedraw || len(diff.Runs) != 1 || diff.Runs[0].Start != 0 || len(diff.Runs[0].Cells) != 5 {
		t.Fatalf("style fallback %+v %v", diff, err)
	}
}

func TestDiffFramesRejectMalformedCurrent(t *testing.T) {
	frames := []Frame{{Width: 0, Height: 1}, {Width: 1, Height: 1}, {Width: 241}, {Width: 1, Height: 3001}}
	for _, f := range frames {
		if _, err := DiffFrames(Frame{}, f); err == nil {
			t.Fatalf("accepted invalid %+v", f)
		}
	}
}

func BenchmarkSmallFrameDiff(b *testing.B) {
	a := Frame{Width: 120, Height: 34, Cells: make([][]Cell, 34)}
	c := Frame{Width: 120, Height: 34, Cells: make([][]Cell, 34)}
	for y := 0; y < 34; y++ {
		a.Cells[y] = make([]Cell, 120)
		c.Cells[y] = make([]Cell, 120)
	}
	c.Cells[18][50] = Cell{Grapheme: "B"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diff, err := DiffFrames(a, c)
		if err != nil || len(diff.Runs) != 1 {
			b.Fatal(diff, err)
		}
	}
}
