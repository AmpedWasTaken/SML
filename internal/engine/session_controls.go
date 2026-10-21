package engine

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var sessionKey = regexp.MustCompile(`^state\.[a-z][a-z0-9_]{0,31}$`)
var sessionOption = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 _-]{0,23}$`)

func controlOptions(n Node) []string {
	if n.Kind != "choice" {
		return nil
	}
	return strings.Split(n.Attr("options", ""), ",")
}

func validateSessionControl(n Node) error {
	if n.Kind != "toggle" && n.Kind != "choice" {
		return nil
	}
	if len(n.Children) > 0 || strings.TrimSpace(n.Text) != "" {
		return errors.New("session control must be empty")
	}
	for key := range n.Attrs {
		if key != "label" && key != "key" && key != "id" && !(n.Kind == "choice" && key == "options") {
			return fmt.Errorf("unsupported session control attribute %q", key)
		}
	}
	if !uiLabel.MatchString(n.Attr("label", "")) || !sessionKey.MatchString(n.Attr("key", "")) {
		return errors.New("invalid session control label or state key")
	}
	if n.Kind == "choice" {
		options := controlOptions(n)
		if len(options) < 2 || len(options) > 8 {
			return errors.New("choice requires 2..8 options")
		}
		seen := map[string]bool{}
		for _, v := range options {
			if !sessionOption.MatchString(v) || seen[v] {
				return errors.New("invalid or repeated choice option")
			}
			seen[v] = true
		}
	}
	return nil
}

func renderSessionControl(n Node, values map[string]string, width int) string {
	label := n.Attr("label", "")
	current := values[n.Attr("key", "")]
	if n.Kind == "toggle" {
		if current != "on" {
			current = "off"
		}
	} else {
		options := controlOptions(n)
		valid := false
		for _, option := range options {
			if option == current {
				valid = true
				break
			}
		}
		if !valid {
			current = options[0]
		}
	}
	return trim("[ "+label+": "+current+" ]", width)
}
