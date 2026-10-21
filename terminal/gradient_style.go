package terminal

import (
	"fmt"
	"math"
	"strings"
)

// GradientStyle is presentation-only, parsed from a validated SML <gradient>
// element. It cannot perform I/O or change the terminal outside ANSI SGR.
type GradientStyle struct {
	Colors    []RGB
	Speed     float64 // cycles per second, 0 = stationary
	FPS       int     // upper bound for the TUI / banner redraw frequency
	Mode      string  // loop, linear, pingpong
	Direction string  // right, left, down, up
	Animated  bool
}

// PaintGradientLine applies the selected color interpolation to printable
// runes. RowFraction supplies vertical position for full-screen / ASCII art.
func PaintGradientLine(value string, style GradientStyle, elapsedSeconds, rowFraction float64) string {
	if len(style.Colors) < 2 {
		return value
	}
	mode := style.Mode
	if mode == "" {
		mode = "loop"
	}
	direction := style.Direction
	if direction == "" {
		direction = "right"
	}
	var phase float64
	if style.Animated {
		phase = PhaseForTime(elapsedSeconds, style.Speed)
	}
	chars := []rune(value)
	count := 0
	for _, r := range chars {
		if r != '\n' && r != '\r' {
			count++
		}
	}
	if count == 0 {
		return value
	}
	var result strings.Builder
	col := 0
	rowFraction = math.Max(0, math.Min(1, rowFraction))
	for _, r := range chars {
		if r == '\n' || r == '\r' {
			result.WriteRune(r)
			continue
		}
		fraction := 0.0
		if count > 1 {
			fraction = float64(col) / float64(count-1)
		}
		switch direction {
		case "left":
			fraction = 1 - fraction
		case "down":
			fraction = rowFraction
		case "up":
			fraction = 1 - rowFraction
		}
		var color RGB
		switch mode {
		case "linear":
			if style.Animated {
				fraction = fract(fraction - phase)
			}
			color = SampleLinear(style.Colors, fraction)
		case "pingpong":
			color = Gradient(style.Colors, fraction-phase)
		default:
			if style.Animated {
				color = Sample(style.Colors, fraction-phase)
			} else {
				color = SampleLinear(style.Colors, fraction)
			}
		}
		if r == ' ' || r == '\t' {
			result.WriteRune(r)
		} else {
			fmt.Fprintf(&result, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", color.R, color.G, color.B, r)
		}
		col++
	}
	return result.String()
}

// PaintGradientLines shares vertical alignment across a multiline logo.
func PaintGradientLines(lines []string, style GradientStyle, elapsedSeconds float64) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		row := 0.0
		if len(lines) > 1 {
			row = float64(i) / float64(len(lines)-1)
		}
		out[i] = PaintGradientLine(line, style, elapsedSeconds, row)
	}
	return out
}
