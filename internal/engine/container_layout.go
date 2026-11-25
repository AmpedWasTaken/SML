package engine

import (
	"fmt"
	"strconv"
	"strings"
)

func layoutStyleAttribute(key string) bool {
	switch key {
	case "width", "min-width", "max-width", "columns", "weights", "gap", "row-gap", "collapse-below", "align-items":
		return true
	}
	return false
}

func layoutStyleTarget(kind, key string) bool {
	switch key {
	case "width", "min-width", "max-width":
		return kind == "box"
	case "columns", "weights":
		return kind == "grid"
	case "gap", "row-gap", "collapse-below", "align-items":
		return kind == "row" || kind == "grid"
	}
	return false
}

func validateLayoutStyleValue(key, value string) error {
	if key == "align-items" {
		switch value {
		case "start", "center", "end":
			return nil
		}
		return fmt.Errorf("align-items requires start, center or end")
	}
	if key == "weights" {
		parts := strings.Split(value, ",")
		if len(parts) < 1 || len(parts) > 8 {
			return fmt.Errorf("grid weights require 1..8 entries")
		}
		for _, part := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || n < 1 || n > 16 {
				return fmt.Errorf("grid weights require integers 1..16")
			}
		}
		return nil
	}
	if key == "width" || key == "min-width" || key == "max-width" {
		if value == "auto" {
			return nil
		}
		if strings.HasSuffix(value, "%") {
			n, err := strconv.Atoi(strings.TrimSuffix(value, "%"))
			if err == nil && n >= 1 && n <= 100 {
				return nil
			}
			return fmt.Errorf("%s must be 1%%..100%%", key)
		}
	}
	lo, hi := 0, 8
	switch key {
	case "width", "min-width", "max-width":
		lo, hi = 12, 240
	case "columns":
		lo, hi = 1, 8
	case "collapse-below":
		lo, hi = 20, 240
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < lo || n > hi {
		return fmt.Errorf("%s must be %d..%d", key, lo, hi)
	}
	return nil
}

func validateContainerLayout(n Node) error {
	if n.Kind != "box" && n.Kind != "row" && n.Kind != "grid" {
		return nil
	}
	for key, value := range n.Attrs {
		if !layoutStyleAttribute(key) {
			continue
		}
		if !layoutStyleTarget(n.Kind, key) {
			return fmt.Errorf("%s is not supported on %s", key, n.Kind)
		}
		if err := validateLayoutStyleValue(key, value); err != nil {
			return err
		}
	}
	return nil
}

func resolveBoxLength(value string, available, fallback int) int {
	if value == "" || value == "auto" {
		return fallback
	}
	if strings.HasSuffix(value, "%") {
		p, err := strconv.Atoi(strings.TrimSuffix(value, "%"))
		if err == nil {
			return available * p / 100
		}
		return fallback
	}
	return boundedInt(value, fallback, 12, 240)
}
