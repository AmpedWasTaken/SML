package terminal

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type RGB struct{ R, G, B uint8 }

func ParseColor(s string) (RGB, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return RGB{}, fmt.Errorf("invalid color %q", s)
	}
	n, err := strconv.ParseUint(s, 16, 24)
	if err != nil {
		return RGB{}, err
	}
	return RGB{uint8(n >> 16), uint8(n >> 8), uint8(n)}, nil
}

func Palette(hex []string) ([]RGB, error) {
	colors := make([]RGB, 0, len(hex))
	for _, s := range hex {
		c, err := ParseColor(s)
		if err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}
	if len(colors) < 2 {
		return nil, fmt.Errorf("gradient needs at least 2 colors")
	}
	return colors, nil
}

func mix(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a)*(1-t) + float64(b)*t))
}

func fract(t float64) float64 {
	t = math.Mod(t, 1)
	if t < 0 {
		t += 1
	}
	return t
}

// Sample interpolates along the palette with seamless wrap (cycling gradient).
func Sample(palette []RGB, t float64) RGB {
	if len(palette) == 0 {
		return RGB{}
	}
	if len(palette) == 1 {
		return palette[0]
	}
	t = fract(t)
	scaled := t * float64(len(palette))
	idx := int(math.Floor(scaled)) % len(palette)
	next := (idx + 1) % len(palette)
	f := scaled - math.Floor(scaled)
	return RGB{mix(palette[idx].R, palette[next].R, f), mix(palette[idx].G, palette[next].G, f), mix(palette[idx].B, palette[next].B, f)}
}

// SampleLinear is the static left→right fade (no last→first wrap).
func SampleLinear(palette []RGB, t float64) RGB {
	if len(palette) == 0 {
		return RGB{}
	}
	if len(palette) == 1 {
		return palette[0]
	}
	if t <= 0 {
		return palette[0]
	}
	if t >= 1 {
		return palette[len(palette)-1]
	}
	scaled := t * float64(len(palette)-1)
	i := int(math.Floor(scaled))
	if i >= len(palette)-1 {
		return palette[len(palette)-1]
	}
	f := scaled - float64(i)
	return RGB{mix(palette[i].R, palette[i+1].R, f), mix(palette[i].G, palette[i+1].G, f), mix(palette[i].B, palette[i+1].B, f)}
}

// Gradient keeps the mirrored ping-pong used by older call sites / tests.
func Gradient(palette []RGB, position float64) RGB {
	if len(palette) == 0 {
		return RGB{}
	}
	if len(palette) == 1 {
		return palette[0]
	}
	x := math.Mod(position, 2)
	if x < 0 {
		x += 2
	}
	if x > 1 {
		x = 2 - x
	}
	x *= float64(len(palette) - 1)
	i := int(math.Floor(x))
	if i >= len(palette)-1 {
		return palette[len(palette)-1]
	}
	t := x - float64(i)
	return RGB{mix(palette[i].R, palette[i+1].R, t), mix(palette[i].G, palette[i+1].G, t), mix(palette[i].B, palette[i+1].B, t)}
}

// PhaseForTime matches animation phase: phase = elapsedSeconds * speed.
func PhaseForTime(elapsedSeconds, speed float64) float64 {
	if speed == 0 {
		return 0
	}
	return elapsedSeconds * speed
}

// PaintLine colours non-space glyphs. smoothLoop scrolls the palette (CNC style);
// otherwise phase 0 is a classic left→right fade.
func PaintLine(text string, palette []RGB, phase float64, smoothLoop bool) string {
	chars := []rune(text)
	if len(chars) == 0 || len(palette) == 0 {
		return text
	}
	width := 0
	for _, r := range chars {
		if r != '\n' && r != '\r' {
			width++
		}
	}
	if width <= 0 {
		return text
	}
	var b strings.Builder
	col := 0
	for _, r := range chars {
		if r == '\n' || r == '\r' {
			b.WriteRune(r)
			continue
		}
		t := 0.0
		if width > 1 {
			t = float64(col) / float64(width-1)
		}
		var c RGB
		if smoothLoop || phase != 0 {
			c = Sample(palette, t-phase)
		} else {
			c = SampleLinear(palette, t)
		}
		if r == ' ' || r == '\t' {
			b.WriteRune(r)
		} else {
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", c.R, c.G, c.B, r)
		}
		col++
	}
	return b.String()
}

// PaintLines paints every line with a shared width so multi-line ASCII stays aligned.
func PaintLines(lines []string, palette []RGB, phase float64, smoothLoop bool) []string {
	width := 0
	for _, line := range lines {
		if n := len([]rune(strings.TrimRight(line, "\r\n"))); n > width {
			width = n
		}
	}
	out := make([]string, len(lines))
	for i, line := range lines {
		chars := []rune(line)
		if len(chars) == 0 || len(palette) == 0 {
			out[i] = line
			continue
		}
		var b strings.Builder
		for col, r := range chars {
			t := 0.0
			if width > 1 {
				t = float64(col) / float64(width-1)
			}
			var c RGB
			if smoothLoop || phase != 0 {
				c = Sample(palette, t-phase)
			} else {
				c = SampleLinear(palette, t)
			}
			if r == ' ' || r == '\t' {
				b.WriteRune(r)
			} else {
				fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm%c\x1b[0m", c.R, c.G, c.B, r)
			}
		}
		out[i] = b.String()
	}
	return out
}

// Paint colours a short string (prompt / TUI title) with a travelling sample.
func Paint(text string, palette []RGB, phase float64) string {
	return PaintLine(text, palette, phase, phase != 0)
}

// Banner animates only before the interactive prompt starts, avoiding competing
// terminal writes while users are typing. The full TUI will own its own tick loop.
func Banner(w io.Writer, text string, palette []RGB, frames, delayMS int, animate bool) error {
	if !animate || frames == 0 {
		_, err := io.WriteString(w, PaintLine(text, palette, 0, false)+"\r\n")
		return err
	}
	if _, err := io.WriteString(w, "\x1b[?25l"); err != nil {
		return err
	}
	defer io.WriteString(w, "\x1b[?25h")
	started := time.Now()
	speed := 1.0
	if frames > 0 && delayMS > 0 {
		speed = 1.0 / (float64(frames) * float64(delayMS) / 1000.0)
	}
	for i := 0; i < frames; i++ {
		phase := PhaseForTime(time.Since(started).Seconds(), speed)
		if _, err := io.WriteString(w, "\r\x1b[2K"+PaintLine(text, palette, phase, true)); err != nil {
			return err
		}
		time.Sleep(time.Duration(delayMS) * time.Millisecond)
	}
	_, err := io.WriteString(w, "\r\x1b[2K"+PaintLine(text, palette, 0, false)+"\r\n")
	return err
}
