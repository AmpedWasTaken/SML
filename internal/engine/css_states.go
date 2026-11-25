package engine

import (
	"fmt"
	"strings"
)

// Only paint attributes may change on hover, focus or disabled state.
// Control identity, frame geometry and enabled state are never changed.
type cssStatePaint struct {
	Foreground  string
	Background  string
	BorderColor string
}

type cssControlPaint struct {
	Hover    cssStatePaint
	Focus    cssStatePaint
	Disabled cssStatePaint
}

func cssStateSelector(selector string) (kind, state string, err error) {
	kind, state, _ = strings.Cut(selector, ":")
	switch kind {
	case "button", "link", "plugin-action", "input", "number", "choice", "toggle":
	default:
		return "", "", fmt.Errorf("state selectors require a control, got %q", kind)
	}
	switch state {
	case "hover", "focus", "disabled":
		return kind, state, nil
	default:
		return "", "", fmt.Errorf("unsupported control state %q", state)
	}
}

func validateCSSStates(styles *assetStyles) error {
	for selector, attrs := range styles.States {
		if _, _, err := cssStateSelector(selector); err != nil {
			return err
		}
		for key, value := range attrs {
			switch key {
			case "foreground", "background", "border-color":
			default:
				return fmt.Errorf("state selector %s cannot change %s", selector, key)
			}
			resolved, err := styles.resolve(key, value)
			if err != nil {
				return err
			}
			if resolved == "inherit" {
				return fmt.Errorf("state selectors require explicit colors")
			}
			attrs[key] = resolved
		}
	}
	return nil
}

func (v View) controlPaint(kind string) cssControlPaint {
	if v.styles == nil {
		return cssControlPaint{}
	}
	paint := func(state string) cssStatePaint {
		attrs := v.styles.States[kind+":"+state]
		return cssStatePaint{
			Foreground:  attrs["foreground"],
			Background:  attrs["background"],
			BorderColor: attrs["border-color"],
		}
	}
	return cssControlPaint{Hover: paint("hover"), Focus: paint("focus"), Disabled: paint("disabled")}
}

func (e Element) withControlPaint(opts GlowOptions) Element {
	if e.disabled {
		return e.withPaint(e.controlPaint.Disabled)
	}
	if e.ID == opts.HoveredID && opts.HoveredID != "" {
		e = e.withPaint(e.controlPaint.Hover)
	}
	if e.ID == opts.FocusedID && opts.FocusedID != "" {
		e = e.withPaint(e.controlPaint.Focus)
	}
	return e
}

func (e Element) withPaint(p cssStatePaint) Element {
	if p.Foreground != "" {
		e.Foreground = p.Foreground
		e.FGGradient = ""
	}
	if p.Background != "" {
		e.Background = p.Background
	}
	if p.BorderColor != "" {
		e.BorderColor = p.BorderColor
		e.BorderGradient = ""
	}
	return e
}
