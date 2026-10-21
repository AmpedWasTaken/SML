package sml

import (
	"strings"
	"testing"
)

func TestSMLV2WidgetsAndFallbacks(t *testing.T) {
	source := `<screen name="demo" version="2"><section title="LIVE METRICS"><field label="Connections" key="plugin.netwatch.connections" fallback="offline"/><field label="Load" key="server.load" unit="%"/><empty field="missing"><text>Not configured</text></empty></section></screen>`
	view, err := Parse([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	rendered := strings.Join(view.RenderWidth(map[string]string{"server.load": "42"}, 80), "\n")
	for _, want := range []string{"LIVE METRICS", "Connections: offline", "Load: 42 %", "Not configured"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q from %s", want, rendered)
		}
	}
}
func TestSMLV2RejectsInvalidTags(t *testing.T) {
	cases := []string{
		`<screen name="test" version="9"><text>invalid</text></screen>`,
		`<screen name="test" version="2"><field label="x" key="../secret"/></screen>`,
		`<screen name="test" version="2"><field label="x" key="valid" command="shell"/></screen>`,
		`<screen name="test" version="2"><empty field="bad-key"><text>x</text></empty></screen>`,
	}
	for _, source := range cases {
		if _, err := Parse([]byte(source)); err == nil {
			t.Errorf("accepted invalid SML: %s", source)
		}
	}
}
func TestSMLV2SanitizesPluginData(t *testing.T) {
	view, err := Parse([]byte(`<screen name="test" version="2"><field label="Status" key="plugin.a.status"/></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	rendered := strings.Join(view.RenderWidth(map[string]string{"plugin.a.status": "ok\x1b[2J\a"}, 80), "\n")
	if strings.Contains(rendered, "\x1b") || strings.Contains(rendered, "\a") {
		t.Fatalf("terminal escape leaked: %q", rendered)
	}
}
