package engine

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The shapes the fuzz target treats as untrusted input. Each one must come back
// as an error: a silent acceptance is a parser bug, and a panic is what
// FuzzParseNeverPanics exists to catch.
func TestParseRejectsHostileInput(t *testing.T) {
	deep := func(n int) string {
		return `<screen name="s">` + strings.Repeat("<column>", n) + "<text>deep</text>" +
			strings.Repeat("</column>", n) + `</screen>`
	}
	many := func(n int) string {
		return `<screen name="s">` + strings.Repeat("<text>x</text>", n) + `</screen>`
	}

	cases := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"empty document", "", "SML root must be"},
		{"duplicate attribute", `<screen name="a" name="b"/>`, "duplicate SML attribute"},
		{"duplicate root", `<screen name="a"/><screen name="b"/>`, "exactly one root"},
		{"unknown component", `<screen name="a"><frobnicate/></screen>`, "unsupported SML component"},
		{"namespace prefix", `<s:screen name="a"/>`, "unsupported SML component"},
		{"processing instruction", `<?xml version="1.0"?><screen name="a"/>`, "processing instructions are not supported"},
		{"invalid named entity", `<screen name="a"><text>&bogus;</text></screen>`, "invalid character entity"},
		{"out of range character reference", `<screen name="a"><text>&#x110000;</text></screen>`, "invalid character entity"},
		{"nul byte", `<screen name="a"><text>` + "\x00" + `</text></screen>`, "illegal character"},
		{"invalid utf8", `<screen name="a"><text>` + "\xff\xfe" + `</text></screen>`, "invalid SML size or UTF-8"},
		{"unclosed element", `<screen name="a"><text>x</text>`, "unexpected EOF"},
		{"text outside the root", `hello<screen name="a"/>`, "unexpected text outside SML root"},
		{"nesting past the depth limit", deep(maxDepth + 1), "node/depth limit exceeded"},
		{"more than the node limit", many(maxNodes + 1), "node/depth limit exceeded"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.input))
			if err == nil {
				t.Fatalf("Parse accepted %q; want an error containing %q", tc.input, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Parse error = %q; want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// A corpus seed the fuzzer skips is dead weight: FuzzParseNeverPanics returns
// early for input longer than 64 KiB, so a seed above that never reaches the
// parser. Keep the committed corpus inside the limit the parser enforces.
func TestFuzzCorpusIsWithinTheParserLimit(t *testing.T) {
	dir := filepath.Join("testdata", "fuzz", "FuzzParseNeverPanics")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read the fuzz corpus: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("no corpus seeds in %s", dir)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		lines := strings.SplitN(string(raw), "\n", 3)
		if len(lines) < 2 || lines[0] != "go test fuzz v1" {
			t.Fatalf("%s: not a corpus file (missing the %q header)", entry.Name(), "go test fuzz v1")
		}
		literal := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(lines[1]), "[]byte("), ")")
		input, err := strconv.Unquote(literal)
		if err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		if len(input) > maxFileSize {
			t.Errorf("%s: %d bytes, over the %d-byte limit the target and the parser both apply",
				entry.Name(), len(input), maxFileSize)
		}
	}
}
