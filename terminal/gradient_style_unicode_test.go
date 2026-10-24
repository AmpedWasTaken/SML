package terminal

import (
	"regexp"
	"strings"
	"testing"
)

func TestSMLGradientDoesNotSplitUnicodeGraphemes(t *testing.T) {
	style := GradientStyle{
		Colors: []RGB{{R: 255}, {B: 255}},
		Mode:   "linear",
	}
	for _, input := range []string{"👩‍💻 é 界", "🇳🇱 OK", "A\r\n👩‍💻"} {
		out := PaintGradientLine(input, style, 0, 0)
		withoutANSI := regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(out, "")
		if withoutANSI != input {
			t.Fatalf("gradient changed grapheme clusters: %q -> %q", input, withoutANSI)
		}
		if strings.Contains(out, "\x1b[0m‍") || strings.Contains(out, "‍\x1b[") {
			t.Fatalf("ANSI code inserted inside emoji cluster: %q", out)
		}
	}
	output := PaintGradientLine("👩‍💻 é 界", style, 0, 0)
	if n := strings.Count(output, "\x1b[38;2;"); n != 3 {
		t.Fatalf("expected one color per visible grapheme, got %d: %q", n, output)
	}
}
