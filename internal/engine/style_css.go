package engine

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// parseStyleCSS implements a deliberately small, static terminal styling
// grammar. It accepts component selectors, classes, and :root tokens only.
func parseStyleCSS(source string) (assetStyles, error) {
	out := assetStyles{
		Tokens: map[string]string{},
		Classes: map[string]map[string]string{},
		Defaults: map[string]map[string]string{},
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
	for rules := 0; strings.TrimSpace(source) != ""; rules++ {
		if rules >= 161 {
			return out, fmt.Errorf("too many SMLCSS rules")
		}
		open := strings.IndexByte(source, '{')
		if open < 0 {
			return out, fmt.Errorf("SMLCSS rule requires opening brace")
		}
		selector := strings.TrimSpace(source[:open])
		rest := source[open+1:]
		close := strings.IndexByte(rest, '}')
		if close < 0 || strings.ContainsRune(rest[:close], '{') {
			return out, fmt.Errorf("invalid SMLCSS rule %q", selector)
		}
		body := rest[:close]
		source = rest[close+1:]
		attrs := map[string]string{}
		for _, decl := range strings.Split(body, ";") {
			decl = strings.TrimSpace(decl)
			if decl == "" {
				continue
			}
			key, value, ok := strings.Cut(decl, ":")
			key, value = strings.TrimSpace(key), strings.TrimSpace(value)
			if !ok || !validName.MatchString(key) || value == "" || len(value) > 64 {
				return out, fmt.Errorf("invalid SMLCSS declaration %q", decl)
			}
			if _, exists := attrs[key]; exists {
				return out, fmt.Errorf("duplicate SMLCSS property %q", key)
			}
			if strings.HasPrefix(value, "var(--") && strings.HasSuffix(value, ")") {
				value = "@" + strings.TrimSuffix(strings.TrimPrefix(value, "var(--"), ")")
			}
			attrs[key] = value
		}
		if len(attrs) == 0 {
			return out, fmt.Errorf("empty SMLCSS rule %q", selector)
		}
		switch {
		case selector == ":root":
			for key, value := range attrs {
				if !strings.HasPrefix(key, "--") || !validName.MatchString(key[2:]) {
					return out, fmt.Errorf("invalid SMLCSS token %q", key)
				}
				if _, exists := out.Tokens[key[2:]]; exists {
					return out, fmt.Errorf("duplicate SMLCSS token %q", key)
				}
				out.Tokens[key[2:]] = value
			}
		case strings.HasPrefix(selector, "."):
			name := selector[1:]
			if !validName.MatchString(name) || out.Classes[name] != nil {
				return out, fmt.Errorf("invalid or duplicate SMLCSS class %q", selector)
			}
			out.Classes[name] = attrs
		default:
			if !styleTargets[selector] || out.Defaults[selector] != nil {
				return out, fmt.Errorf("invalid or duplicate SMLCSS component %q", selector)
			}
			out.Defaults[selector] = attrs
		}
	}
	return out, nil
}
