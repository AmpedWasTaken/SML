package engine

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

// validateAnimation accepts only bounded, inert visual effects on panels
// and tables. Animation metadata cannot change layout or dispatch actions.
func validateAnimation(n Node) error {
	effect, hasEffect := n.Attrs["animate"]
	duration, hasDuration := n.Attrs["duration-ms"]
	easing, hasEasing := n.Attrs["easing"]
	if !hasEffect && !hasDuration && !hasEasing {
		return nil
	}
	if n.Kind != "panel" && n.Kind != "table" {
		return fmt.Errorf("animation is supported only on panels and tables")
	}
	switch effect {
	case "pulse", "sweep", "fade":
	default:
		return fmt.Errorf("unsupported animation %q", effect)
	}
	if hasDuration {
		ms, err := strconv.Atoi(duration)
		if err != nil || ms < 250 || ms > 10000 {
			return fmt.Errorf("animation duration-ms must be 250..10000")
		}
	}
	if hasEasing {
		switch easing {
		case "linear", "ease-in", "ease-out", "ease-in-out":
		default:
			return fmt.Errorf("unsupported animation easing %q", easing)
		}
	}
	return nil
}

func animationEase(t float64, easing string) float64 {
	t = max(0, min(1, t))
	switch easing {
	case "ease-in":
		return t * t
	case "ease-out":
		return 1 - (1-t)*(1-t)
	case "ease-in-out":
		return t * t * (3 - 2*t)
	default:
		return t
	}
}

// animationBrightness is pure and quantized to 12 frames per second.
// A fade completes once; pulse and sweep loop. No glyph is ever hidden.
func animationBrightness(e Element, phase time.Duration, base int, position float64) int {
	duration := time.Duration(max(250, min(10000, e.DurationMS))) * time.Millisecond
	phase = max(0, phase)
	phase = phase / (time.Second / 12) * (time.Second / 12)
	progress := float64(phase%duration) / float64(duration)
	if e.Animation == "fade" {
		progress = min(1, float64(phase)/float64(duration))
	}
	progress = animationEase(progress, e.Easing)
	var strength float64
	switch e.Animation {
	case "fade":
		strength = .45 + .55*progress
	case "pulse":
		strength = .55 + .45*(.5-.5*math.Cos(2*math.Pi*progress))
	case "sweep":
		distance := math.Abs(position - progress)
		distance = min(distance, 1-distance)
		strength = .55 + .45*max(0, 1-distance*5)
	default:
		return base
	}
	return max(30, min(100, int(math.Round(float64(base)*strength))))
}
