package sml

import "github.com/AmpedWasTaken/SML/internal/textcell"

// CellWidth reports a string's terminal display width in cells.
func CellWidth(s string) int { return textcell.CellWidth(s) }

// ClipCells clips to a cell width without breaking grapheme clusters.
func ClipCells(s string, width int) string { return textcell.ClipCells(s, width) }

// PadCells clips and pads a string to the requested display width.
func PadCells(s string, width int) string { return textcell.PadCells(s, width) }

// WrapText wraps text at grapheme boundaries, preferring word boundaries.
func WrapText(s string, width int) []string { return textcell.WrapText(s, width) }
