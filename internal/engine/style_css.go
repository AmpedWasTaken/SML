package engine

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// cssBlock extracts a balanced declaration or keyframe block. SMLCSS has
// no strings, URLs, nested selectors, or executable expressions.
func cssBlock(source string) (head, body, tail string, err error) {
	open := strings.IndexByte(source, '{')
	if open < 0 {
		return "", "", "", fmt.Errorf("SMLCSS rule requires opening brace")
	}
	depth := 1
	for i := open + 1; i < len(source); i++ {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(source[:open]), source[open+1 : i], source[i+1:], nil
			}
		}
		if depth > 2 {
			return "", "", "", fmt.Errorf("SMLCSS nesting limit exceeded")
		}
	}
	return "", "", "", fmt.Errorf("unclosed SMLCSS rule")
}

func cssDeclarations(body string) (map[string]string, error) {
	attrs := map[string]string{}
	if strings.ContainsAny(body, "{}") {
		return nil, fmt.Errorf("nested declarations are not supported")
	}
	for _, decl := range strings.Split(body, ";") {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}
		key, value, ok := strings.Cut(decl, ":")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !ok || !validName.MatchString(key) || value == "" || len(value) > 64 {
			return nil, fmt.Errorf("invalid SMLCSS declaration %q", decl)
		}
		if _, exists := attrs[key]; exists {
			return nil, fmt.Errorf("duplicate SMLCSS property %q", key)
		}
		if strings.HasPrefix(value, "var(--") && strings.HasSuffix(value, ")") {
			value = "@" + strings.TrimSuffix(strings.TrimPrefix(value, "var(--"), ")")
		}
		attrs[key] = value
	}
	if len(attrs) == 0 {
		return nil, fmt.Errorf("empty SMLCSS declaration block")
	}
	return attrs, nil
}

func parseStyleCSS(source string) (assetStyles, error) {
	out := assetStyles{
		Tokens:   map[string]string{},
		Classes:  map[string]map[string]string{},
		Defaults: map[string]map[string]string{},
		States:   map[string]map[string]string{},
	}
	if !utf8.ValidString(source) || len(source) > maxFileSize {
		return out, fmt.Errorf("invalid SMLCSS size or UTF-8")
	}
	for {
		start := strings.Index(source, "/*")
		if start < 0 {
			break
		}
		end := strings.Index(source[start+2:], "*/")
		if end < 0 {
			return out, fmt.Errorf("unterminated SMLCSS comment")
		}
		source = source[:start] + " " + source[start+end+4:]
	}
	keyframes := map[string]string{}
	for rules := 0; strings.TrimSpace(source) != ""; rules++ {
		if rules >= 256 {
			return out, fmt.Errorf("too many SMLCSS rules")
		}
		selector, body, tail, err := cssBlock(source)
		if err != nil {
			return out, err
		}
		source = tail
		if strings.HasPrefix(selector, "@media ") {
			media, err := parseCSSMedia(selector, body)
			if err != nil {
				return out, err
			}
			out.Media = append(out.Media, media...)
			if len(out.Media) > 128 {
				return out, fmt.Errorf("too many SMLCSS media rules")
			}
			continue
		}
		if strings.HasPrefix(selector, "@keyframes ") {
			name := strings.TrimSpace(strings.TrimPrefix(selector, "@keyframes "))
			if !validName.MatchString(name) || keyframes[name] != "" || len(keyframes) >= 32 || name == "pulse" || name == "sweep" || name == "fade" {
				return out, fmt.Errorf("invalid or duplicate SMLCSS animation %q", name)
			}
			curve, err := cssKeyframes(body)
			if err != nil {
				return out, fmt.Errorf("keyframes %s: %w", name, err)
			}
			keyframes[name] = curve
			continue
		}
		attrs, err := cssDeclarations(body)
		if err != nil {
			return out, err
		}
		if selector == ":root" {
			for key, value := range attrs {
				if !strings.HasPrefix(key, "--") || !validName.MatchString(key[2:]) {
					return out, fmt.Errorf("invalid SMLCSS token %q", key)
				}
				out.Tokens[key[2:]] = value
			}
			continue
		}
		selectors := strings.Split(selector, ",")
		if len(selectors) > 16 {
			return out, fmt.Errorf("too many grouped SMLCSS selectors")
		}
		for _, part := range selectors {
			part = strings.TrimSpace(part)
			var target map[string]string
			if strings.ContainsAny(part, " >\t\n\r") || (strings.Contains(part, ":") && strings.ContainsAny(part, ".#")) {
				rule, err := parseContextSelector(part)
				if err != nil {
					return out, err
				}
				rule.attr = make(map[string]string, len(attrs))
				out.Compound = append(out.Compound, rule)
				if len(out.Compound) > 128 {
					return out, fmt.Errorf("too many compound selectors")
				}
				target = rule.attr
			} else if strings.Contains(part, ":") {
				if _, _, err := cssStateSelector(part); err != nil {
					return out, err
				}
				if out.States[part] == nil {
					out.States[part] = map[string]string{}
				}
				target = out.States[part]
			} else if strings.Contains(part, "#") || strings.Count(part, ".") > 1 || (strings.Contains(part, ".") && !strings.HasPrefix(part, ".")) {
				rule, err := parseCompoundSelector(part)
				if err != nil {
					return out, err
				}
				rule.attr = make(map[string]string, len(attrs))
				out.Compound = append(out.Compound, rule)
				if len(out.Compound) > 128 {
					return out, fmt.Errorf("too many compound selectors")
				}
				target = rule.attr
			} else if strings.HasPrefix(part, ".") {
				name := part[1:]
				if !validName.MatchString(name) {
					return out, fmt.Errorf("invalid SMLCSS class %q", part)
				}
				if out.Classes[name] == nil {
					out.Classes[name] = map[string]string{}
				}
				target = out.Classes[name]
			} else {
				if !styleTargets[part] {
					return out, fmt.Errorf("invalid SMLCSS component %q", part)
				}
				if out.Defaults[part] == nil {
					out.Defaults[part] = map[string]string{}
				}
				target = out.Defaults[part]
			}
			for key, value := range attrs {
				target[key] = value
			}
		}
	}
	out.Transitions = make(map[string]cssTransition)
	for kind, attrs := range out.Defaults {
		if value, ok := attrs["transition"]; ok {
			if _, _, err := cssStateSelector(kind + ":hover"); err != nil {
				return out, err
			}
			if strings.HasPrefix(value, "@") {
				value = out.Tokens[value[1:]]
			}
			spec, err := parseCSSTransition(value)
			if err != nil {
				return out, err
			}
			out.Transitions[kind] = spec
			delete(attrs, "transition")
			if len(attrs) == 0 {
				delete(out.Defaults, kind)
			}
		}
	}
	for _, group := range []map[string]map[string]string{out.Classes, out.Defaults, out.States} {
		for _, attrs := range group {
			if err := normalizeCSS(attrs, out.Tokens, keyframes); err != nil {
				return out, err
			}
		}
	}
	for _, rule := range out.Compound {
		if err := normalizeCSS(rule.attr, out.Tokens, keyframes); err != nil {
			return out, err
		}
	}
	for _, rule := range out.Media {
		if err := normalizeCSS(rule.attr, out.Tokens, keyframes); err != nil {
			return out, err
		}
	}
	return out, nil
}

// normalizeCSS maps familiar CSS spellings onto validated engine attributes.
func normalizeCSS(attrs, tokens, keyframes map[string]string) error {
	alias := map[string]string{"color": "foreground", "background-color": "background", "text-align": "align", "text-transform": "case", "column-gap": "gap"}
	for old, next := range alias {
		if v, ok := attrs[old]; ok {
			if _, exists := attrs[next]; exists {
				return fmt.Errorf("conflicting SMLCSS properties %s and %s", old, next)
			}
			delete(attrs, old)
			if old == "text-transform" {
				switch v {
				case "uppercase":
					v = "upper"
				case "lowercase":
					v = "lower"
				}
			}
			attrs[next] = v
		}
	}
	if value, ok := attrs["grid-template-columns"]; ok {
		if _, exists := attrs["columns"]; exists {
			return fmt.Errorf("grid-template-columns conflicts with columns")
		}
		if _, exists := attrs["weights"]; exists {
			return fmt.Errorf("grid-template-columns conflicts with weights")
		}
		if strings.HasPrefix(value, "@") {
			value = tokens[value[1:]]
		}
		parts := strings.Fields(value)
		if len(parts) < 1 || len(parts) > 8 {
			return fmt.Errorf("grid-template-columns requires 1..8 fr tracks")
		}
		weights := make([]string, len(parts))
		for i, part := range parts {
			weight, err := strconv.Atoi(strings.TrimSuffix(part, "fr"))
			if err != nil || !strings.HasSuffix(part, "fr") || weight < 1 || weight > 16 {
				return fmt.Errorf("grid tracks must be 1fr..16fr")
			}
			weights[i] = strconv.Itoa(weight)
		}
		delete(attrs, "grid-template-columns")
		attrs["columns"] = strconv.Itoa(len(weights))
		attrs["weights"] = strings.Join(weights, ",")
	}
	if value, ok := attrs["animation"]; ok {
		if strings.HasPrefix(value, "@") {
			value = tokens[value[1:]]
		}
		parts := strings.Fields(value)
		if len(parts) < 2 || len(parts) > 4 {
			return fmt.Errorf("animation requires name duration [easing] [iteration]")
		}
		for _, key := range []string{"animate", "duration-ms", "easing", "animation-curve", "animation-repeat"} {
			if _, exists := attrs[key]; exists {
				return fmt.Errorf("animation conflicts with %s", key)
			}
		}
		ms, err := cssMilliseconds(parts[1])
		if err != nil {
			return err
		}
		attrs["animate"], attrs["duration-ms"] = parts[0], strconv.Itoa(ms)
		if curve := keyframes[parts[0]]; curve != "" {
			attrs["animate"], attrs["animation-curve"] = "keyframes", curve
			if strings.Contains(curve, "|") {
				attrs["animate"] = "color-keyframes"
			}
		} else if parts[0] != "pulse" && parts[0] != "sweep" && parts[0] != "fade" {
			return fmt.Errorf("unknown SMLCSS animation %q", parts[0])
		}
		if len(parts) >= 3 {
			attrs["easing"] = parts[2]
		}
		if len(parts) == 4 {
			attrs["animation-repeat"] = parts[3]
		}
		delete(attrs, "animation")
	}
	if value, ok := attrs["animation-delay"]; ok {
		if _, exists := attrs["animation-delay-ms"]; exists {
			return fmt.Errorf("conflicting animation delay properties")
		}
		if strings.HasPrefix(value, "@") {
			value = tokens[value[1:]]
		}
		delay, err := cssDelayMilliseconds(value)
		if err != nil {
			return err
		}
		attrs["animation-delay-ms"] = strconv.Itoa(delay)
		delete(attrs, "animation-delay")
	}
	if value, ok := attrs["animation-iteration-count"]; ok {
		if _, exists := attrs["animation-repeat"]; exists {
			return fmt.Errorf("conflicting animation iteration properties")
		}
		attrs["animation-repeat"] = value
		delete(attrs, "animation-iteration-count")
	}
	for _, key := range []string{"width", "min-width", "max-width", "gap", "row-gap", "collapse-below", "padding", "flex-basis", "left"} {
		if v, ok := attrs[key]; ok {
			attrs[key] = strings.TrimSuffix(v, "ch")
		}
	}
	return nil
}
