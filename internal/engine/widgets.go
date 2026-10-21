package engine

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func validateWidget(n Node) error {
	switch n.Kind {
	case "gauge":
		allowed := map[string]bool{"label": true, "value": true, "width": true}
		for key := range n.Attrs {
			if !allowed[key] {
				return fmt.Errorf("unsupported gauge attribute %q", key)
			}
		}
		if len(n.Children) > 0 {
			return errors.New("gauge cannot contain child elements")
		}
		if len(n.Attr("label", "")) > 64 {
			return errors.New("gauge label too long")
		}
		w, err := strconv.Atoi(n.Attr("width", "16"))
		if err != nil || w < 4 || w > 48 {
			return errors.New("gauge width must be 4..48")
		}
	case "sparkline":
		allowed := map[string]bool{"label": true, "values": true, "width": true}
		for key := range n.Attrs {
			if !allowed[key] {
				return fmt.Errorf("unsupported sparkline attribute %q", key)
			}
		}
		if len(n.Children) > 0 {
			return errors.New("sparkline cannot contain child elements")
		}
		if len(n.Attr("label", "")) > 64 {
			return errors.New("sparkline label too long")
		}
		w, err := strconv.Atoi(n.Attr("width", "24"))
		if err != nil || w < 4 || w > 60 {
			return errors.New("sparkline width must be 4..60")
		}
	}
	return nil
}

func renderGauge(n Node, values map[string]string, width int) []string {
	title := interpolate(n.Attr("label", ""), values)
	nWidth, _ := strconv.Atoi(n.Attr("width", "16"))
	raw := interpolate(n.Attr("value", "0"), values)
	pct, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsInf(pct, 0) || math.IsNaN(pct) {
		pct = 0
	}
	pct = math.Max(0, math.Min(100, pct))
	filled := int(math.Round(pct * float64(nWidth) / 100))
	line := fmt.Sprintf("%s [%s%s] %.0f%%", title, strings.Repeat("█", filled), strings.Repeat("░", nWidth-filled), pct)
	return []string{trim(strings.TrimSpace(line), width)}
}

func renderSparkline(n Node, values map[string]string, width int) []string {
	title := interpolate(n.Attr("label", ""), values)
	raw := interpolate(n.Attr("values", ""), values)
	limit, _ := strconv.Atoi(n.Attr("width", "24"))
	segments := strings.Split(raw, ",")
	samples := make([]float64, 0, min(limit, len(segments)))
	for i, part := range segments {
		if i >= 128 {
			break
		}
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.ParseFloat(part, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		samples = append(samples, v)
	}
	if len(samples) == 0 {
		return []string{trim(title+" (no samples)", width)}
	}
	if len(samples) > limit {
		samples = samples[len(samples)-limit:]
	}
	lower, upper := samples[0], samples[0]
	for _, v := range samples {
		lower = math.Min(lower, v)
		upper = math.Max(upper, v)
	}
	blocks := []rune("▁▂▃▄▅▆▇█")
	var graph strings.Builder
	for _, v := range samples {
		index := 0
		if upper > lower {
			index = int(math.Round(7 * (v - lower) / (upper - lower)))
		}
		graph.WriteRune(blocks[index])
	}
	return []string{trim(strings.TrimSpace(title+" "+graph.String()), width)}
}
