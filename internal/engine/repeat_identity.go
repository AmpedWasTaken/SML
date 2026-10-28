package engine

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Keyed repeats are intentionally limited to scalar CSV item values.
// Structured record keys belong to a future bounded typed data-provider API.
var repeatIndexAlias = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

func validateRepeatSyntax(n Node) error {
	if n.Kind != "repeat" {
		return nil
	}
	_, keyed := n.Attrs["key-by"]
	_, indexed := n.Attrs["index-as"]
	if keyed {
		if n.Attr("key-by", "") != "value" {
			return errors.New("repeat key-by must be value")
		}
		if !bindingName.MatchString(n.Attr("source", "")) ||
			!validName.MatchString(n.Attr("as", "item")) {
			return errors.New("keyed repeat requires a safe source and alias")
		}
	}
	if indexed {
		alias := n.Attr("index-as", "")
		if !repeatIndexAlias.MatchString(alias) || alias == n.Attr("as", "item") {
			return errors.New("repeat index-as must be a distinct safe local alias")
		}
		switch alias {
		case "user", "server", "plugin", "session", "state", "derived", "app", "repeat":
			return errors.New("repeat index-as cannot shadow a protected namespace")
		}
	}
	return nil
}

// repeatItems limits keyed item-source scanning and always respects the
// existing 0..100 repetition limit. Legacy (unkeyed) semantics are preserved.
func repeatItems(n Node, raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	limit := boundedInt(n.Attr("limit", "30"), 30, 0, 100)
	if limit == 0 {
		return nil, nil
	}
	if n.Attr("key-by", "") == "value" {
		if len(raw) > 32768 || !utf8.ValidString(raw) {
			return nil, errors.New("keyed repeat source must be valid UTF-8 and at most 32KiB")
		}
		items := strings.SplitN(raw, ",", limit+1)
		return items[:min(len(items), limit)], nil
	}
	items := strings.Split(raw, ",")
	return items[:min(len(items), limit)], nil
}

// repeatInstanceID returns a deterministic fragment used by layout paths.
// Existing unkeyed documents retain index identities. Keyed items retain
// their identity across reorders and reject ambiguous/missing values.
func repeatInstanceID(n Node, value string, index int, seen map[string]bool) (string, error) {
	if n.Attr("key-by", "") != "value" {
		return fmt.Sprintf("repeat:%d", index), nil
	}
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || !utf8.ValidString(value) {
		return "", errors.New("keyed repeat item needs a nonempty UTF-8 key of at most 128 bytes")
	}
	if seen[value] {
		return "", errors.New("duplicate keyed repeat item")
	}
	seen[value] = true
	digest := sha256.Sum256([]byte(value))
	return fmt.Sprintf("repeat:key:%x", digest[:]), nil
}
