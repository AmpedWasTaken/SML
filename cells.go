package sml

import (
	"strings"

	"github.com/rivo/uniseg"
)

// CellWidth returns the number of terminal display cells required by a string.
// It accounts for combining characters, wide runes, flags, and emoji sequences.
func CellWidth(s string) int { return uniseg.StringWidth(s) }

// ClipCells clips text without splitting an extended grapheme cluster.
func ClipCells(s string, width int) string {
	if width <= 0 {
		return ""
	}
	var out strings.Builder
	used := 0
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		w := g.Width()
		if used+w > width {
			break
		}
		out.WriteString(g.Str())
		used += w
	}
	return out.String()
}

// PadCells returns exactly width terminal display cells.
func PadCells(s string, width int) string {
	s = ClipCells(s, width)
	return s + strings.Repeat(" ", max(0, width-CellWidth(s)))
}

// WrapText wraps at grapheme boundaries, preferring word boundaries.
// A glyph wider than the available width cannot be displayed and is omitted.
func WrapText(s string, width int) []string {
	if width < 1 {
		return []string{""}
	}
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		var current strings.Builder
		used := 0
		flush := func() { lines = append(lines, current.String()); current.Reset(); used = 0 }
		var token strings.Builder
		tokenWidth := 0
		spaces := false
		emit := func() {
			if token.Len() == 0 {
				return
			}
			if !spaces && used > 0 && used+tokenWidth > width {
				flush()
			}
			if spaces && used == 0 {
				token.Reset()
				tokenWidth = 0
				return
			}
			g := uniseg.NewGraphemes(token.String())
			for g.Next() {
				segment, w := g.Str(), g.Width()
				if w > width {
					continue
				}
				if used+w > width {
					flush()
				}
				current.WriteString(segment)
				used += w
			}
			token.Reset()
			tokenWidth = 0
		}
		g := uniseg.NewGraphemes(paragraph)
		for g.Next() {
			segment := g.Str()
			isSpace := segment == " " || segment == "\t"
			if isSpace != spaces && token.Len() > 0 {
				emit()
			}
			spaces = isSpace
			if segment == "\t" {
				segment = "    "
			}
			token.WriteString(segment)
			tokenWidth += CellWidth(segment)
		}
		emit()
		flush()
	}
	return lines
}
