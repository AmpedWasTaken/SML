package engine

import (
	"fmt"
	"strconv"
	"strings"
)

type colorStop struct {
	position int
	color    string
}

func colorKeyframeProperty(property string) bool {
	switch property {
	case "foreground", "background", "border-color", "glow-color":
		return true
	}
	return false
}

func parseColorCurve(curve string) (string, []colorStop, error) {
	property, raw, ok := strings.Cut(curve, "|")
	if !ok || !colorKeyframeProperty(property) {
		return "", nil, fmt.Errorf("invalid color keyframe property")
	}
	parts := strings.Split(raw, ",")
	if len(parts) < 2 || len(parts) > 8 {
		return "", nil, fmt.Errorf("color keyframes require 2..8 stops")
	}
	stops := make([]colorStop, 0, len(parts))
	for _, part := range parts {
		a, color, ok := strings.Cut(part, ":")
		position, err := strconv.Atoi(a)
		if !ok || err != nil || position < 0 || position > 100 || !validHexColor(color) ||
			(len(stops) > 0 && position <= stops[len(stops)-1].position) {
			return "", nil, fmt.Errorf("invalid color keyframe %q", part)
		}
		stops = append(stops, colorStop{position, color})
	}
	if stops[0].position != 0 || stops[len(stops)-1].position != 100 {
		return "", nil, fmt.Errorf("color keyframes require 0%% and 100%% endpoints")
	}
	return property, stops, nil
}

func sampleColorCurve(stops []colorStop, progress float64) string {
	p := max(0, min(1, progress)) * 100
	for i := 1; i < len(stops); i++ {
		a, b := stops[i-1], stops[i]
		if p <= float64(b.position) {
			t := (p - float64(a.position)) / float64(b.position-a.position)
			return interpolateCSSColor(a.color, b.color, t)
		}
	}
	return stops[len(stops)-1].color
}

func (e Element) withColorKeyframes(opts GlowOptions) Element {
	if e.Animation != "color-keyframes" || !opts.Animate || opts.ReducedMotion {
		return e
	}
	property, stops, err := parseColorCurve(e.AnimationCurve)
	if err != nil {
		return e
	}
	progress, active := animationProgress(e, opts.Phase)
	if !active {
		return e
	}
	color := sampleColorCurve(stops, progress)
	switch property {
	case "foreground":
		e.Foreground, e.FGGradient = color, ""
	case "background":
		e.Background = color
	case "border-color":
		e.BorderColor, e.BorderGradient = color, ""
	case "glow-color":
		e.GlowColor = color
		if e.Glow == "none" {
			e.Glow = "soft"
		}
	}
	return e
}
