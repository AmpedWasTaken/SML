package engine

import (
	"math"
	"strings"
)

func parseCellGradient(value string) ([]glowRGB, bool) {
	parts := strings.Split(value, ",")
	if len(parts) < 2 || len(parts) > 8 {
		return nil, false
	}
	colors := make([]glowRGB, len(parts))
	for i, part := range parts {
		color, ok := glowParseHex(strings.TrimSpace(part))
		if !ok {
			return nil, false
		}
		colors[i] = color
	}
	return colors, true
}

func sampleCellGradient(colors []glowRGB, position float64) glowRGB {
	position = max(0, min(1, position)) * float64(len(colors)-1)
	index := min(len(colors)-2, int(position))
	fraction := position - float64(index)
	a, b := colors[index], colors[index+1]
	mix := func(x, y int) int {
		return int(math.Round(float64(x) + (float64(y)-float64(x))*fraction))
	}
	return glowRGB{mix(a.r, b.r), mix(a.g, b.g), mix(a.b, b.b)}
}

func perimeterPosition(x, y, width, height int) float64 {
	right, bottom := max(0, width-1), max(0, height-1)
	offset := 0
	switch {
	case y == 0:
		offset = x
	case x == right:
		offset = right + y
	case y == bottom:
		offset = right + bottom + right - x
	default:
		offset = 2*right + bottom + bottom - y
	}
	return float64(offset) / float64(max(1, 2*(right+bottom)))
}
