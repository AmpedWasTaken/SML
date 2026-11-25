package engine

// ASCII returns a detached, plain-ASCII structural frame for terminals without
// reliable Unicode rendering. Cell positions, IDs and hotspot coordinates
// remain stable, even when a wide grapheme becomes a single ASCII character.
// Hosts must select this conservative mode based on terminal capabilities.
func (f Frame) ASCII() Frame {
	out := f
	out.Lines = make([]string, len(f.Cells))
	out.Cells = make([][]Cell, len(f.Cells))
	out.Elements = append([]Element(nil), f.Elements...)
	out.Hotspots = append([]Hotspot(nil), f.Hotspots...)
	for y, source := range f.Cells {
		row := append([]Cell(nil), source...)
		for x := range row {
			cell := &row[x]
			if cell.Continuation {
				cell.Continuation = false
				cell.Grapheme = " "
				continue
			}
			if cell.Grapheme != "" {
				cell.Grapheme = asciiGlyph(cell.Grapheme)
			}
		}
		out.Cells[y] = row
		out.Lines[y] = cellLineText(row)
	}
	return out
}

func asciiGlyph(g string) string {
	if len(g) == 1 && g[0] >= 32 && g[0] <= 126 {
		return g
	}
	switch g {
	case "─", "━", "═", "┄", "┈", "–", "—":
		return "-"
	case "│", "┃", "║":
		return "|"
	case "┌", "┐", "└", "┘", "╭", "╮", "╰", "╯", "┏", "┓", "┗", "┛", "╔", "╗", "╚", "╝", "┼", "├", "┤", "┬", "┴":
		return "+"
	case "█", "▓", "■", "●", "◆", "★":
		return "#"
	case "░", "▒", "·", "○", "□":
		return "."
	case "›", "→", "▶":
		return ">"
	case "‹", "←", "◀":
		return "<"
	case "    ":
		return " "
	default:
		return "?"
	}
}
