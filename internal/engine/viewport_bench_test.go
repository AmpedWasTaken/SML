package engine

import (
	"strconv"
	"testing"
)

func benchmarkWindowFrame() Frame {
	const width, height = 120, 600
	frame := Frame{Width: width, Height: height, Cells: make([][]Cell, height)}
	for y := range frame.Cells {
		row := make([]Cell, width)
		label := strconv.Itoa(y)
		for x := range row {
			glyph := "."
			if x < len(label) {
				glyph = string(label[x])
			}
			row[x] = Cell{Grapheme: glyph}
		}
		frame.Cells[y] = row
	}
	return frame
}

// BenchmarkFrameWindow tracks post-render allocation cost independently of
// the parser and host terminal I/O. Do not use the result as a lazy DataGrid
// throughput claim: the complete source Frame is already resident in memory.
func BenchmarkFrameWindow(b *testing.B) {
	frame := benchmarkWindowFrame()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := frame.Window(10, i%500, 80, 24); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFrameHitTest(b *testing.B) {
	frame := Frame{
		Width: 80, Height: 1, Cells: [][]Cell{make([]Cell, 80)},
		Hotspots: []Hotspot{{ID: "inspect", Line: 0, Start: 5, End: 15}},
	}
	for x := 5; x < 15; x++ {
		frame.Cells[0][x].ElementID = "inspect"
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := frame.HitTest(7, 0); !ok {
			b.Fatal("hit testing regressed")
		}
	}
}
