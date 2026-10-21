package sml

import (
	"regexp"
	"strconv"
	"strings"
)

// Derived values are presentation-only. They cannot shadow server, user,
// plugin or session state, and never execute code or mutate the host.
var derivedKey = regexp.MustCompile(`^derived\.[a-z][a-z0-9_]{0,31}$`)

const maxDerivations = 24
const maxDerivedValue = 256

func validateDerive(n Node) error {
	if len(n.Children) > 0 || strings.TrimSpace(n.Text) != "" {
		return errDerived("derive cannot contain content")
	}
	key := n.Attr("key", "")
	if !derivedKey.MatchString(key) || !bindingName.MatchString(n.Attr("source", "")) {
		return errDerived("derive requires a derived.* key and safe source binding")
	}
	op := n.Attr("op", "")
	allowedAttrs := map[string]bool{"key": true, "op": true, "source": true}
	switch op {
	case "coalesce":
		allowedAttrs["fallback"] = true
		if len(n.Attr("fallback", "")) > maxDerivedValue {
			return errDerived("derive fallback too long")
		}
	case "equals":
		allowedAttrs["value"] = true
		if len(n.Attr("value", "")) > maxDerivedValue {
			return errDerived("derive comparison too long")
		}
	case "concat":
		allowedAttrs["other"] = true
		allowedAttrs["separator"] = true
		if !bindingName.MatchString(n.Attr("other", "")) || len(n.Attr("separator", "")) > 48 {
			return errDerived("invalid derive concat binding or separator")
		}
	case "duration", "uppercase", "lowercase":
		// These are pure view formatting operations and have no extra fields.
	case "sum":
		allowedAttrs["other"] = true
		if !bindingName.MatchString(n.Attr("other", "")) {
			return errDerived("invalid derive sum binding")
		}
	default:
		return errDerived("unsupported derive operator")
	}
	for attr := range n.Attrs {
		if !allowedAttrs[attr] {
			return errDerived("unsupported derive attribute")
		}
	}
	if n.Attr("source", "") == key || n.Attr("other", "") == key {
		return errDerived("derive cannot depend on itself")
	}
	return nil
}

type derivedError string

func (e derivedError) Error() string { return "SML: " + string(e) }
func errDerived(s string) error      { return derivedError(s) }

// ResolveBindings calculates bounded, validated view-local values from
// already-authorized input bindings. Declarations evaluate in screen order:
// a derive may consume an earlier derive, but never an untrusted expression.
// The input map is not mutated.
func (v View) ResolveBindings(input map[string]string) map[string]string {
	out := make(map[string]string, len(input)+maxDerivations)
	for k, val := range input {
		if !derivedKey.MatchString(k) {
			out[k] = val
		}
	}
	for _, n := range v.Nodes {
		if n.Kind != "derive" {
			continue
		}
		key, op := n.Attr("key", ""), n.Attr("op", "")
		// Defense in depth for views constructed programmatically without Parse.
		if !derivedKey.MatchString(key) || validateDerive(n) != nil {
			continue
		}
		source := out[n.Attr("source", "")]
		var result string
		switch op {
		case "coalesce":
			result = source
			if strings.TrimSpace(result) == "" {
				result = n.Attr("fallback", "")
			}
		case "equals":
			if source == n.Attr("value", "") {
				result = "true"
			} else {
				result = "false"
			}
		case "concat":
			result = source + n.Attr("separator", "") + out[n.Attr("other", "")]
		case "uppercase":
			result = strings.ToUpper(source)
		case "lowercase":
			result = strings.ToLower(source)
		case "duration":
			seconds, err := strconv.ParseInt(strings.TrimSpace(source), 10, 64)
			if err == nil && seconds >= 0 && seconds <= 315360000 {
				result = formatUptime(seconds)
			}
		case "sum":
			a, ea := strconv.ParseInt(strings.TrimSpace(source), 10, 64)
			b, eb := strconv.ParseInt(strings.TrimSpace(out[n.Attr("other", "")]), 10, 64)
			if ea == nil && eb == nil && a >= -1000000000 && a <= 1000000000 && b >= -1000000000 && b <= 1000000000 {
				result = strconv.FormatInt(a+b, 10)
			}
		}
		// Derived values are single-line display data, even when a trusted
		// binding contains tabs or line breaks.
		clean := strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(safe(result))
		runes := []rune(clean)
		if len(runes) > maxDerivedValue {
			runes = runes[:maxDerivedValue]
		}
		out[key] = string(runes)
	}
	return out
}

func requiresSML3(nodes []Node) bool {
	for _, n := range nodes {
		_, hasGate := n.Attrs["enabled-when"]
		if n.Kind == "derive" || n.Kind == "otherwise" || hasStyleAttributes(n) || hasGate || requiresSML3(n.Children) {
			return true
		}
	}
	return false
}

// formatUptime emits a compact, terminal-friendly unit string for a bounded
// non-negative number of seconds. A second precision value is shown only
// when the duration is shorter than a minute.
func formatUptime(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	remain := seconds % 60
	switch {
	case days > 0:
		if hours > 0 {
			return strconv.FormatInt(days, 10) + "d " + strconv.FormatInt(hours, 10) + "h"
		}
		return strconv.FormatInt(days, 10) + "d"
	case hours > 0:
		if minutes > 0 {
			return strconv.FormatInt(hours, 10) + "h " + strconv.FormatInt(minutes, 10) + "m"
		}
		return strconv.FormatInt(hours, 10) + "h"
	case minutes > 0:
		return strconv.FormatInt(minutes, 10) + "m"
	default:
		return strconv.FormatInt(remain, 10) + "s"
	}
}
