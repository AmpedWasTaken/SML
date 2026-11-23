package engine

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// These widgets are presentation-only. All data is supplied by the host as
// bounded string bindings; no SML component opens files or starts jobs.
func validateDashboardWidget(n Node) error {
	attributes := map[string]map[string]bool{
		"heatmap":       {"label": true, "values": true, "columns": true},
		"bar-chart":     {"label": true, "values": true, "labels": true, "width": true},
		"stacked-bar":   {"label": true, "values": true, "labels": true, "width": true},
		"waveform":      {"label": true, "values": true, "width": true, "height": true},
		"kpi":           {"label": true, "value": true, "unit": true, "delta": true},
		"activity-feed": {"label": true, "events": true, "max-items": true},
	}
	allowed, known := attributes[n.Kind]
	if !known {
		return nil
	}
	if len(n.Children) != 0 || strings.TrimSpace(n.Text) != "" {
		return fmt.Errorf("%s cannot contain child elements or text", n.Kind)
	}
	for key, val := range n.Attrs {
		if !allowed[key] {
			return fmt.Errorf("unsupported %s attribute %q", n.Kind, key)
		}
		if len(val) > 4096 {
			return fmt.Errorf("%s attribute %q exceeds size limit", n.Kind, key)
		}
	}
	if len(n.Attr("label", "")) > 64 {
		return errors.New("widget label too long")
	}
	if len(n.Attr("unit", "")) > 24 {
		return errors.New("kpi unit too long")
	}
	for _, rule := range []struct {
		name string
		def, min, max int
	}{
		{"columns", 12, 2, 48},
		{"width", 24, 4, 60},
		{"height", 4, 2, 8},
		{"max-items", 6, 1, 12},
	} {
		raw, ok := n.Attrs[rule.name]
		if !ok {
			continue
		}
		num, err := strconv.Atoi(raw)
		if err != nil || num < rule.min || num > rule.max {
			return fmt.Errorf("%s %s must be %d..%d", n.Kind, rule.name, rule.min, rule.max)
		}
		if n.Kind == "waveform" && rule.name == "width" && num < 8 {
			return errors.New("waveform width must be 8..60")
		}
	}
	return nil
}

func widgetLiteral(s string, width int) string {
	t := strings.NewReplacer("\n", " ", "\t", " ", "\r", " ").Replace(safe(s))
	return trim(strings.TrimSpace(t), max(0, width))
}

func widgetText(s string, values map[string]string, width int) string {
	return widgetLiteral(interpolate(s, values), width)
}

// parseWidgetNumbers ignores invalid/non-finite values and caps both the token
// count and numeric magnitude. A negative reading never inflates chart geometry.
func parseWidgetNumbers(raw string, limit int) []float64 {
	parts := strings.SplitN(raw, ",", limit+1)
	out := make([]float64, 0, min(limit, len(parts)))
	for i, part := range parts {
		if i >= limit {
			break
		}
		num, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
			continue
		}
		out = append(out, math.Max(-1e9, math.Min(1e9, num)))
	}
	return out
}

func widgetHeader(label string, values map[string]string, width int) []string {
	if title := widgetText(label, values, width); title != "" {
		return []string{title}
	}
	return nil
}

func renderDashboardWidget(n Node, values map[string]string, width int) []string {
	switch n.Kind {
	case "heatmap":
		return renderHeatmap(n, values, width)
	case "bar-chart":
		return renderBarChart(n, values, width)
	case "stacked-bar":
		return renderStackedBar(n, values, width)
	case "waveform":
		return renderWaveform(n, values, width)
	case "kpi":
		return renderKPI(n, values, width)
	case "activity-feed":
		return renderActivityFeed(n, values, width)
	}
	return nil
}

func renderHeatmap(n Node, values map[string]string, width int) []string {
	out := widgetHeader(n.Attr("label", ""), values, width)
	samples := parseWidgetNumbers(interpolate(n.Attr("values", ""), values), 96)
	if len(samples) == 0 {
		return append(out, "(no samples)")
	}
	cols := min(boundedInt(n.Attr("columns", "12"), 12, 2, 48), max(2, width/2))
	shades := []string{"·", "░", "▒", "▓", "█"}
	for start := 0; start < len(samples); start += cols {
		var row []string
		for _, v := range samples[start:min(start+cols, len(samples))] {
			pct := math.Max(0, math.Min(100, v))
			level := int(math.Floor(pct / 25))
			row = append(row, shades[min(level, 4)])
		}
		out = append(out, trim(strings.Join(row, " "), width))
	}
	return out
}

func renderBarChart(n Node, values map[string]string, width int) []string {
	out := widgetHeader(n.Attr("label", ""), values, width)
	samples := parseWidgetNumbers(interpolate(n.Attr("values", ""), values), 12)
	if len(samples) == 0 {
		return append(out, "(no samples)")
	}
	labels := strings.SplitN(interpolate(n.Attr("labels", ""), values), ",", 13)
	top := 0.0
	for _, v := range samples {
		top = math.Max(top, math.Max(0, v))
	}
	labelWidth := min(12, max(3, width/5))
	barWidth := min(boundedInt(n.Attr("width", "24"), 24, 4, 60), max(2, width-labelWidth-15))
	if width < 28 {
		barWidth = max(2, width-labelWidth-2)
	}
	for i, sample := range samples {
		name := fmt.Sprintf("%d", i+1)
		if i < len(labels) && strings.TrimSpace(labels[i]) != "" {
			name = widgetLiteral(labels[i], labelWidth)
		}
		ratio := 0.0
		if top > 0 {
			ratio = math.Max(0, sample) / top
		}
		filled := max(0, min(barWidth, int(math.Round(ratio*float64(barWidth)))))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		line := pad(name, labelWidth) + " " + bar
		if width >= 28 {
			line += fmt.Sprintf(" %g", sample)
		}
		out = append(out, trim(line, width))
	}
	return out
}

func renderStackedBar(n Node, values map[string]string, width int) []string {
	out := widgetHeader(n.Attr("label", ""), values, width)
	samples := parseWidgetNumbers(interpolate(n.Attr("values", ""), values), 8)
	labels := strings.SplitN(interpolate(n.Attr("labels", ""), values), ",", 9)
	total := 0.0
	for i := range samples {
		samples[i] = math.Max(0, samples[i])
		total += samples[i]
	}
	if total <= 0 {
		return append(out, "(no positive samples)")
	}
	barWidth := min(boundedInt(n.Attr("width", "24"), 24, 4, 60), max(4, width))
	styles := []string{"█", "▓", "▒", "░", "▰", "▤", "▥", "▦"}
	var graph strings.Builder
	used := 0
	cumulative := 0.0
	for i, sample := range samples {
		cumulative += sample
		next := int(math.Round(cumulative / total * float64(barWidth)))
		next = max(used, min(barWidth, next))
		graph.WriteString(strings.Repeat(styles[i%len(styles)], next-used))
		used = next
	}
	out = append(out, trim(graph.String(), width))
	for i, sample := range samples {
		name := fmt.Sprintf("GROUP %d", i+1)
		if i < len(labels) && strings.TrimSpace(labels[i]) != "" {
			name = widgetLiteral(labels[i], 30)
		}
		out = append(out, trim(fmt.Sprintf("%s %s  %.1f%%", styles[i%len(styles)], name, 100*sample/total), width))
	}
	return out
}

func renderWaveform(n Node, values map[string]string, width int) []string {
	out := widgetHeader(n.Attr("label", ""), values, width)
	samples := parseWidgetNumbers(interpolate(n.Attr("values", ""), values), 128)
	if len(samples) == 0 {
		return append(out, "(no samples)")
	}
	graphWidth := min(boundedInt(n.Attr("width", "32"), 32, 8, 60), width)
	height := boundedInt(n.Attr("height", "4"), 4, 2, 8)
	if len(samples) > graphWidth {
		samples = samples[len(samples)-graphWidth:]
	}
	lo, hi := samples[0], samples[0]
	for _, v := range samples {
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	for row := height; row >= 1; row-- {
		var b strings.Builder
		for range graphWidth - len(samples) {
			b.WriteByte(' ')
		}
		for _, v := range samples {
			level := height / 2
			if hi > lo {
				level = int(math.Round((v - lo) * float64(height-1) / (hi - lo))) + 1
			}
			if level >= row {
				b.WriteRune('█')
			} else {
				b.WriteRune(' ')
			}
		}
		out = append(out, trim(b.String(), width))
	}
	return out
}

func renderKPI(n Node, values map[string]string, width int) []string {
	label := widgetText(n.Attr("label", ""), values, width)
	value := widgetText(n.Attr("value", ""), values, 96)
	if value == "" {
		value = "—"
	}
	unit := widgetText(n.Attr("unit", ""), values, 24)
	line := strings.TrimSpace(label + "  " + value)
	if unit != "" {
		line += " " + unit
	}
	if delta := widgetText(n.Attr("delta", ""), values, 32); delta != "" {
		num, err := strconv.ParseFloat(strings.TrimSuffix(delta, "%"), 64)
		if err == nil && !math.IsNaN(num) && !math.IsInf(num, 0) {
			arrow := "→"
			if num > 0 {
				arrow = "▲"
			} else if num < 0 {
				arrow = "▼"
			}
			line += "  " + arrow + " " + delta
		}
	}
	return []string{trim(line, width)}
}

func renderActivityFeed(n Node, values map[string]string, width int) []string {
	out := widgetHeader(n.Attr("label", ""), values, width)
	raw := interpolate(n.Attr("events", ""), values)
	events := strings.SplitN(raw, ";", 65)
	valid := make([]string, 0, min(12, len(events)))
	for i, entry := range events {
		if i >= 64 {
			break
		}
		parts := strings.SplitN(entry, "|", 3)
		if len(parts) != 3 {
			continue
		}
		stamp := widgetLiteral(parts[0], 12)
		state := strings.ToLower(strings.TrimSpace(parts[1]))
		message := widgetLiteral(parts[2], width)
		if message == "" {
			continue
		}
		icon := "·"
		switch state {
		case "ok", "success", "online":
			icon = "●"
		case "warn", "warning":
			icon = "!"
		case "error", "failed", "offline":
			icon = "×"
		case "info":
			icon = "○"
		}
		valid = append(valid, trim(icon+" "+stamp+" "+message, width))
	}
	limit := boundedInt(n.Attr("max-items", "6"), 6, 1, 12)
	if len(valid) > limit {
		valid = valid[len(valid)-limit:]
	}
	if len(valid) == 0 {
		return append(out, "(no events)")
	}
	return append(out, valid...)
}
