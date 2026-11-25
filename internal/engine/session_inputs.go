package engine

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var inputDefault = regexp.MustCompile(`^[ -~]{0,64}$`)

func validateSessionInput(n Node) error {
	if n.Kind != "input" && n.Kind != "number" {
		return nil
	}
	if len(n.Children) > 0 || strings.TrimSpace(n.Text) != "" {
		return errors.New("SML input cannot contain children")
	}
	for name := range n.Attrs {
		switch name {
		case "label", "key", "id", "placeholder", "min", "max", "border":
		case "appearance", "foreground", "background", "border-color", "foreground-gradient", "border-gradient":
			if n.Kind != "input" {
				return fmt.Errorf("attribute %q requires input", name)
			}
		default:
			return fmt.Errorf("unsupported SML input attribute %q", name)
		}
	}
	if !uiLabel.MatchString(n.Attr("label", "")) || !sessionKey.MatchString(n.Attr("key", "")) {
		return errors.New("invalid input label or local state key")
	}
	if !inputDefault.MatchString(n.Attr("placeholder", "")) {
		return errors.New("unsafe SML input placeholder")
	}
	if n.Kind == "input" {
		if appearance := n.Attr("appearance", "inline"); appearance != "inline" && appearance != "boxed" {
			return errors.New("SML input appearance must be inline or boxed")
		}
		if border := n.Attr("border", "round"); n.Attr("appearance", "") == "boxed" && border == "none" {
			return errors.New("boxed SML input requires a visible border")
		}
		if _, ok := n.Attrs["min"]; ok {
			return errors.New("min is only supported on number")
		}
		if _, ok := n.Attrs["max"]; ok {
			return errors.New("max is only supported on number")
		}
	} else {
		lo, err := strconv.Atoi(n.Attr("min", "0"))
		if err != nil || lo < 0 || lo > 1000000 {
			return errors.New("invalid number min")
		}
		hi, err := strconv.Atoi(n.Attr("max", "100"))
		if err != nil || hi < lo || hi > 1000000 {
			return errors.New("invalid number max")
		}
	}
	return nil
}
func renderSessionInput(n Node, values map[string]string, width int) string {
	v := values[n.Attr("key", "")]
	if v == "" {
		v = n.Attr("placeholder", "")
	}
	if v == "" {
		v = "(empty)"
	}
	if utf8.RuneCountInString(v) > 64 {
		v = string([]rune(v)[:64])
	}
	if hasControlBorder(n) {
		return trim(n.Attr("label", "")+": "+safe(v), width)
	}
	return trim("[ "+n.Attr("label", "")+": "+safe(v)+" ]", width)
}

func renderBoxedInput(n Node, values map[string]string, width int) []string {
	border := styleBorder(n.Attr("border", "round"))
	value := values[n.Attr("key", "")]
	if value == "" {
		value = n.Attr("placeholder", "")
	}
	value = safe(strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(value))
	inner := max(0, width-2)
	rule := strings.Repeat(border.h, inner)
	return []string{
		border.tl + rule + border.tr,
		border.v + PadCells(value, inner) + border.v,
		border.bl + rule + border.br,
	}
}
