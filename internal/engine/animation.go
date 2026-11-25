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
	curve, hasCurve := n.Attrs["animation-curve"]
	repeat, hasRepeat := n.Attrs["animation-repeat"]
	delay, hasDelay := n.Attrs["animation-delay-ms"]
	direction, hasDirection := n.Attrs["animation-direction"]
	if !hasEffect && !hasDuration && !hasEasing && !hasCurve && !hasRepeat && !hasDelay && !hasDirection {
		return nil
	}
	if n.Kind != "panel" && n.Kind != "table" {
		return fmt.Errorf("animation is supported only on panels and tables")
	}
	switch effect {
	case "pulse", "sweep", "fade", "keyframes", "color-keyframes":
	default:
		return fmt.Errorf("unsupported animation %q", effect)
	}
	if effect == "keyframes" {
		if _, err := parseBrightnessCurve(curve); err != nil {
			return err
		}
	} else if effect == "color-keyframes" {
		if _, _, err := parseColorCurve(curve); err != nil {
			return err
		}
	} else if hasCurve {
		return fmt.Errorf("animation-curve requires keyframes")
	}
	if hasRepeat && !validAnimationRepeat(repeat) {
		return fmt.Errorf("animation-repeat must be 1..100 or infinite")
	}
	if hasDelay && !validAnimationDelay(delay) {
		return fmt.Errorf("animation-delay-ms must be 0..10000")
	}
	if hasDirection && !validAnimationDirection(direction) {
		return fmt.Errorf("unsupported animation direction %q", direction)
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

func validAnimationRepeat(value string) bool {
	n, err := strconv.Atoi(value)
	return value == "infinite" || (err == nil && n >= 1 && n <= 100)
}

func validAnimationDelay(value string) bool {
	n, err := strconv.Atoi(value)
	return err == nil && n >= 0 && n <= 10000
}

func validAnimationDirection(value string) bool {
	switch value {
	case "normal", "reverse", "alternate", "alternate-reverse":
		return true
	}
	return false
}

// animationProgress keeps exact cycle boundaries and finite endpoints,
// while quantizing intermediate frames to at most twelve updates per second.
func animationProgress(e Element, phase time.Duration) (float64, bool) {
	duration := time.Duration(max(250, min(10000, e.DurationMS))) * time.Millisecond
	delay := time.Duration(max(0, min(10000, e.AnimationDelayMS))) * time.Millisecond
	if phase < delay {
		return 0, false
	}
	phase -= delay
	cycles, _ := strconv.Atoi(e.AnimationRepeat)
	if e.AnimationRepeat == "" && e.Animation == "fade" {
		cycles = 1
	}
	cycle := int64(phase / duration)
	remainder := phase % duration
	progress := float64(remainder/(time.Second/12)*(time.Second/12)) / float64(duration)
	if cycles > 0 && cycle >= int64(cycles) {
		cycle, progress = int64(cycles-1), 1
	}
	reverse := e.AnimationDirection == "reverse" ||
		(e.AnimationDirection == "alternate" && cycle%2 == 1) ||
		(e.AnimationDirection == "alternate-reverse" && cycle%2 == 0)
	if reverse {
		progress = 1 - progress
	}
	return animationEase(progress, e.Easing), true
}

// animationBrightness is pure and never changes glyphs or geometry.
func animationBrightness(e Element, phase time.Duration, base int, position float64) int {
	progress, active := animationProgress(e, phase)
	if !active {
		return base
	}
	var strength float64
	switch e.Animation {
	case "keyframes":
		strength = keyframeStrength(e.AnimationCurve, progress)
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
