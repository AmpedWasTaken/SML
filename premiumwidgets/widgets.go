// Package premiumwidgets renders inert, host-driven terminal widget frames.
// No terminal IO, timers, filesystem access or command execution occurs here.
package premiumwidgets

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const Height = 6
const MaxFPS = 12

type Definition struct {
	ID, Title, Category   string
	Animated, Interactive bool
}

var catalog = []Definition{
	{"gradient-meter", "Gradient Meter", "Metrics", true, false},
	{"segmented-gauge", "Segmented Gauge", "Metrics", true, false},
	{"dual-meter", "Dual Meter", "Metrics", true, false},
	{"vertical-meter", "Vertical Meter", "Metrics", true, false},
	{"radial-gauge", "Radial Gauge", "Metrics", true, false},
	{"odometer", "Odometer", "Metrics", true, false},
	{"delta-card", "Delta Card", "Metrics", true, false},
	{"quota-ring", "Quota Ring", "Metrics", true, false},
	{"wave-chart", "Wave Chart", "Charts", true, false},
	{"bar-race", "Bar Race", "Charts", true, false},
	{"sparkline-plus", "Sparkline Plus", "Charts", true, false},
	{"heatmap", "Heatmap", "Charts", true, false},
	{"histogram", "Histogram", "Charts", true, false},
	{"scatter-plot", "Scatter Plot", "Charts", true, false},
	{"stacked-bars", "Stacked Bars", "Charts", true, false},
	{"area-chart", "Area Chart", "Charts", true, false},
	{"orbit-spinner", "Orbit Spinner", "Motion", true, false},
	{"pulse-beacon", "Pulse Beacon", "Motion", true, false},
	{"skeleton-loader", "Skeleton Loader", "Motion", true, false},
	{"shimmer-text", "Shimmer Text", "Motion", true, false},
	{"equalizer", "Equalizer", "Motion", true, false},
	{"countdown", "Countdown", "Motion", true, false},
	{"step-progress", "Step Progress", "Status", true, false},
	{"timeline", "Timeline", "Status", true, false},
	{"notification-toast", "Notification Toast", "Status", true, true},
	{"activity-feed", "Activity Feed", "Status", true, false},
	{"status-matrix", "Status Matrix", "Status", true, false},
	{"carousel", "Carousel", "Layout", false, true},
	{"dial-control", "Dial Control", "Controls", false, true},
	{"slider-control", "Slider Control", "Controls", false, true},
	{"toggle-switch", "Toggle Switch", "Controls", false, true},
	{"rating-control", "Rating Control", "Controls", false, true},
}

func Catalog() []Definition { return append([]Definition(nil), catalog...) }

func Lookup(id string) (Definition, bool) {
	for _, d := range catalog {
		if d.ID == id {
			return d, true
		}
	}
	return Definition{}, false
}

type Input struct {
	Width                int
	Value, Secondary     float64
	Series               []float64
	Phase                time.Duration
	ASCII, ReducedMotion bool
	Selection            int
	Enabled, Dismissed   bool
	Label                string
}

type Frame struct{ Lines [Height]string }

func (f Frame) Bindings(prefix string) map[string]string {
	out := make(map[string]string, Height)
	for i, line := range f.Lines {
		out[fmt.Sprintf("%s.line%d", prefix, i+1)] = line
	}
	return out
}

func clamp(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	return math.Min(v, 100)
}

func glyph(a bool, ascii, unicode string) string {
	if a {
		return ascii
	}
	return unicode
}

func label(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 {
			b.WriteRune(r)
		}
		if b.Len() >= 32 {
			break
		}
	}
	return b.String()
}

func bar(v float64, w int, a bool) string {
	n := int(math.Round(clamp(v) * float64(w) / 100))
	return strings.Repeat(glyph(a, "#", "█"), n) + strings.Repeat(glyph(a, ".", "░"), w-n)
}

func pct(v float64) string { return fmt.Sprintf("%3.0f%%", clamp(v)) }

func sample(s []float64, i, w int) float64 {
	if len(s) == 0 {
		return 0
	}
	return clamp(s[min(len(s)-1, int(float64(i)*float64(len(s))/float64(w)))])
}

func chart(s []float64, w, h int, a bool, kind string) []string {
	rows := make([]string, h)
	for y := 0; y < h; y++ {
		var b strings.Builder
		for x := 0; x < w; x++ {
			v := sample(s, x, w)
			level := int(math.Round(v * float64(h-1) / 100))
			on := h-1-y == level
			if kind == "area-chart" || kind == "histogram" {
				on = h-1-y <= level && v > 0
			}
			if kind == "scatter-plot" && x%2 != 0 {
				on = false
			}
			if on {
				if kind == "scatter-plot" {
					b.WriteString(glyph(a, "o", "●"))
				} else {
					b.WriteString(glyph(a, "#", "█"))
				}
			} else {
				b.WriteByte(' ')
			}
		}
		rows[y] = b.String()
	}
	return rows
}

func Render(id string, in Input) (Frame, error) {
	d, ok := Lookup(id)
	if !ok {
		return Frame{}, fmt.Errorf("unknown widget %q", id)
	}
	w := max(16, min(72, in.Width))
	a := in.ASCII
	v, s := clamp(in.Value), clamp(in.Secondary)
	sel := max(0, min(4, in.Selection))
	phase := int((in.Phase / (time.Second / MaxFPS)) % 120)
	if phase < 0 || in.ReducedMotion {
		phase = 0
	}
	caption := label(in.Label)
	if caption == "" {
		caption = "SAMPLE"
	}
	lines := []string{d.Title}
	switch id {
	case "gradient-meter":
		chars := []rune(glyph(a, ".-=+#", "░▒▓██"))
		n := int(math.Round(v * float64(w) / 100))
		var b strings.Builder
		for i := 0; i < w; i++ {
			if i < n {
				b.WriteRune(chars[min(4, i*5/w)])
			} else {
				b.WriteString(glyph(a, ".", "░"))
			}
		}
		lines = append(lines, caption+" "+pct(v), b.String())
	case "segmented-gauge":
		n := int(math.Ceil(v / 10))
		lines = append(lines, strings.Repeat(glyph(a, "#", "■"), n)+strings.Repeat(glyph(a, ".", "□"), 10-n)+" "+pct(v))
	case "dual-meter":
		lines = append(lines, "A "+pct(v), bar(v, w, a), "B "+pct(s), bar(s, w, a))
	case "vertical-meter":
		for y := 4; y >= 1; y-- {
			lines = append(lines, fmt.Sprintf("%3d |%s|", y*25, bar(clamp((v-float64(y-1)*25)*4), w-7, a)))
		}
		lines = append(lines, pct(v))
	case "radial-gauge", "quota-ring":
		if id == "quota-ring" {
			v = s
		}
		ring := []string{" . . . ", ".     .", ".     .", ".     .", " . . . "}
		coords := [][2]int{{0, 3}, {0, 5}, {1, 6}, {3, 6}, {4, 3}, {4, 1}, {3, 0}, {1, 0}}
		for _, xy := range coords[:int(math.Round(v*8/100))] {
			row := []rune(ring[xy[0]])
			row[xy[1]] = []rune(glyph(a, "#", "●"))[0]
			ring[xy[0]] = string(row)
		}
		ring[2] = " " + pct(v) + " "
		lines = append(lines, ring...)
	case "odometer":
		lines = append(lines, "+-------------+", fmt.Sprintf("| %09.0f   |", v*100), "+-------------+")
	case "delta-card":
		lines = append(lines, caption, fmt.Sprintf("%6.1f", v), fmt.Sprintf("%+.1f vs prior", v-s), bar(v, w, a))
	case "wave-chart", "scatter-plot", "area-chart", "histogram":
		lines = append(lines, chart(in.Series, w, 4, a, id)...)
		lines = append(lines, "0"+strings.Repeat("-", w-4)+"100")
	case "bar-race":
		for i := 0; i < 4; i++ {
			x := sample(in.Series, i, 4)
			lines = append(lines, fmt.Sprintf("%c %s %s", 'A'+i, bar(x, w-8, a), pct(x)))
		}
	case "sparkline-plus":
		chars := []rune(glyph(a, "._-=+*#@", "▁▂▃▄▅▆▇█"))
		var b strings.Builder
		for i := 0; i < w; i++ {
			b.WriteRune(chars[int(math.Round(sample(in.Series, i, w)*7/100))])
		}
		lines = append(lines, b.String(), "Latest: "+pct(v), "Reference: "+pct(s))
	case "heatmap":
		chars := []rune(glyph(a, " .:#+", " ░▒▓█"))
		for y := 0; y < 4; y++ {
			var b strings.Builder
			for x := 0; x < w; x++ {
				b.WriteRune(chars[int(sample(in.Series, y*w+x, w*4)*4/100)])
			}
			lines = append(lines, b.String())
		}
		lines = append(lines, "low .... high")
	case "stacked-bars":
		for i := 0; i < 4; i++ {
			n := int(sample(in.Series, i, 4) * float64(w) / 100)
			m := (w - n) / 2
			lines = append(lines, strings.Repeat(glyph(a, "#", "█"), n)+strings.Repeat(glyph(a, "+", "▒"), m)+strings.Repeat(glyph(a, ".", "░"), w-n-m))
		}
	case "orbit-spinner":
		ring := []string{"   .   ", " .   . ", ".     .", " .   . ", "   .   "}
		coords := [][2]int{{0, 3}, {1, 5}, {2, 6}, {3, 5}, {4, 3}, {3, 1}, {2, 0}, {1, 1}}
		xy := coords[phase%len(coords)]
		row := []rune(ring[xy[0]])
		row[xy[1]] = []rune(glyph(a, "@", "●"))[0]
		ring[xy[0]] = string(row)
		lines = append(lines, ring...)
	case "pulse-beacon":
		p := phase % 6
		p = min(p, 6-p)
		lines = append(lines, "", strings.Repeat(" ", 3-p)+strings.Repeat(".", p)+glyph(a, "*", "●")+strings.Repeat(".", p), "Active")
	case "skeleton-loader", "shimmer-text":
		for y := 0; y < 4; y++ {
			var b strings.Builder
			for x := 0; x < w-y*2; x++ {
				if (x+phase+y*3)%12 < 3 {
					b.WriteString(glyph(a, "#", "█"))
				} else {
					b.WriteString(glyph(a, ".", "░"))
				}
			}
			if id == "shimmer-text" {
				b.Reset()
				for x, ch := range []rune("Rendering preview") {
					if (x+phase+y*3)%12 < 3 {
						b.WriteString(strings.ToUpper(string(ch)))
					} else {
						b.WriteRune(ch)
					}
				}
			}
			lines = append(lines, b.String())
		}
	case "equalizer":
		vals := make([]float64, 8)
		for i := range vals {
			vals[i] = clamp(sample(in.Series, i, 8) + 20*math.Sin(float64(phase+i*4)/5))
		}
		lines = append(lines, chart(vals, w, 4, a, "area-chart")...)
	case "countdown":
		elapsed := max(0, min(100, int(in.Phase/time.Second)))
		if in.ReducedMotion {
			elapsed = 0
		}
		remaining := max(0, int(v)-elapsed)
		lines = append(lines, fmt.Sprintf("%02d:%02d", remaining/60, remaining%60), bar(float64(remaining), w, a))
	case "step-progress":
		var b strings.Builder
		for i := 0; i < 4; i++ {
			mark := "[ ]"
			if v >= float64((i+1)*25) {
				mark = "[x]"
			} else if v >= float64(i*25) {
				mark = "[>]"
			}
			b.WriteString(mark + " ")
		}
		lines = append(lines, b.String(), bar(v, w, a))
	case "timeline":
		for i, name := range []string{"Started", "Prepared", "Reviewed", "Complete"} {
			mark := "[ ]"
			if v >= float64((i+1)*25) {
				mark = "[x]"
			}
			lines = append(lines, mark+" "+name)
		}
	case "notification-toast":
		if in.Dismissed {
			lines = append(lines, "No notification")
		} else {
			lines = append(lines, "+ New notification", caption, "[ Dismiss ]", bar(float64(phase%100), w, a))
		}
	case "activity-feed":
		for i, name := range []string{"Draft saved", "Preview ready", "Theme updated", "Review complete"} {
			mark := " "
			if i == phase/12%4 {
				mark = ">"
			}
			lines = append(lines, mark+" "+name)
		}
	case "status-matrix":
		for y := 0; y < 4; y++ {
			var b strings.Builder
			for x := 0; x < 8; x++ {
				if (x+y*3+phase/4)%9 == 0 {
					b.WriteString("? ")
				} else {
					b.WriteString(glyph(a, "+ ", "● "))
				}
			}
			lines = append(lines, b.String())
		}
	case "carousel":
		lines = append(lines, fmt.Sprintf("< Preview %d of 5 >", sel+1), strings.Repeat(". ", sel)+"* "+strings.Repeat(". ", 4-sel))
	case "dial-control":
		dirs := []string{"W", "NW", "N", "NE", "E"}
		lines = append(lines, "   N", " W + E", "Position: "+dirs[sel])
	case "slider-control":
		lines = append(lines, strings.Repeat("-", sel*2)+"O"+strings.Repeat("-", 8-sel*2), fmt.Sprintf("%d%%", sel*25))
	case "toggle-switch":
		state := "[ OFF ]"
		if in.Enabled {
			state = "[ ON  ]"
		}
		lines = append(lines, state, caption)
	case "rating-control":
		lines = append(lines, strings.Repeat(glyph(a, "*", "★"), sel+1)+strings.Repeat(glyph(a, ".", "☆"), 4-sel), fmt.Sprintf("%d / 5", sel+1))
	}
	var f Frame
	for i := range f.Lines {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		runes := []rune(line)
		runes = runes[:min(w, len(runes))]
		f.Lines[i] = string(runes) + strings.Repeat(" ", w-len(runes))
	}
	return f, nil
}

// Apply modifies only detached demonstration state.
func Apply(id, action string, in Input) (Input, error) {
	d, ok := Lookup(id)
	if !ok || !d.Interactive {
		return in, fmt.Errorf("widget is not interactive")
	}
	switch id {
	case "toggle-switch":
		if action == "toggle" {
			in.Enabled = !in.Enabled
			return in, nil
		}
	case "notification-toast":
		if action == "dismiss" || action == "reset" {
			in.Dismissed = action == "dismiss"
			return in, nil
		}
	default:
		switch action {
		case "next":
			in.Selection = min(4, max(0, in.Selection)+1)
			return in, nil
		case "prev":
			in.Selection = max(0, min(4, in.Selection)-1)
			return in, nil
		}
	}
	return in, fmt.Errorf("unsupported widget action")
}
