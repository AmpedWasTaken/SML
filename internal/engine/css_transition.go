package engine

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type cssTransition struct {
	Property string
	Duration time.Duration
	Easing   string
}

func parseCSSTransition(value string) (cssTransition, error) {
	parts := strings.Fields(value)
	if len(parts) < 2 || len(parts) > 3 {
		return cssTransition{}, fmt.Errorf("transition requires property duration [easing]")
	}
	property := parts[0]
	switch property {
	case "color":
		property = "foreground"
	case "background-color":
		property = "background"
	case "foreground", "background", "border-color", "all":
	default:
		return cssTransition{}, fmt.Errorf("unsupported transition property %q", property)
	}
	duration, err := time.ParseDuration(parts[1])
	if err != nil || duration < 16*time.Millisecond || duration > 10*time.Second || duration%time.Millisecond != 0 {
		return cssTransition{}, fmt.Errorf("transition duration must be 16ms..10s")
	}
	easing := "ease-in-out"
	if len(parts) == 3 {
		easing = parts[2]
		switch easing {
		case "linear", "ease-in", "ease-out", "ease-in-out":
		default:
			return cssTransition{}, fmt.Errorf("unsupported transition easing %q", easing)
		}
	}
	return cssTransition{Property: property, Duration: duration, Easing: easing}, nil
}

func (e Element) withTransitionPaint(opts GlowOptions) Element {
	target := e.withControlPaint(opts)
	spec := e.transition
	if !opts.Transition || opts.ReducedMotion || spec.Duration <= 0 || e.disabled {
		return target
	}
	previous := opts
	previous.FocusedID = opts.PreviousFocusedID
	previous.HoveredID = opts.PreviousHoveredID
	source := e.withControlPaint(previous)
	phase := max(0, min(spec.Duration, opts.TransitionElapsed))
	t := animationEase(float64(phase)/float64(spec.Duration), spec.Easing)
	// Terminal default colors and gradients have no known RGB endpoints.
	// Such properties switch directly rather than inventing a color.
	if (spec.Property == "all" || spec.Property == "foreground") && source.FGGradient == "" && target.FGGradient == "" {
		target.Foreground = interpolateCSSColor(source.Foreground, target.Foreground, t)
	}
	if spec.Property == "all" || spec.Property == "background" {
		target.Background = interpolateCSSColor(source.Background, target.Background, t)
	}
	if (spec.Property == "all" || spec.Property == "border-color") && source.BorderGradient == "" && target.BorderGradient == "" {
		target.BorderColor = interpolateCSSColor(source.BorderColor, target.BorderColor, t)
	}
	return target
}

func interpolateCSSColor(from, to string, t float64) string {
	a, okA := glowParseHex(from)
	b, okB := glowParseHex(to)
	if !okA || !okB || t >= 1 || from == to {
		return to
	}
	mix := func(x, y int) int {
		return int(math.Round(float64(x) + (float64(y)-float64(x))*t))
	}
	return fmt.Sprintf("#%02x%02x%02x", mix(a.r, b.r), mix(a.g, b.g), mix(a.b, b.b))
}
