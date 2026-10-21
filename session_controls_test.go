package sml

import (
	"strings"
	"testing"
)

func TestSessionWidgetsRenderAndValidate(t *testing.T) {
	source := `<screen name="demo" version="2"><toggle label="Local preview" key="state.preview"/><choice label="Density" key="state.density" options="comfortable,compact,dense"/></screen>`
	v, err := Parse([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	lines, hotspots, _ := v.RenderInteractive(map[string]string{}, 90)
	output := strings.Join(lines, "\n")
	if !strings.Contains(output, "Local preview: off") || !strings.Contains(output, "Density: comfortable") {
		t.Fatalf("default widget output: %s", output)
	}
	if len(hotspots) != 2 || hotspots[0].Kind != "toggle" || hotspots[1].Kind != "choice" {
		t.Fatalf("widget hotspots: %+v", hotspots)
	}
	lines, _, _ = v.RenderInteractive(map[string]string{"state.preview": "on", "state.density": "compact"}, 90)
	output = strings.Join(lines, "\n")
	if !strings.Contains(output, "Local preview: on") || !strings.Contains(output, "Density: compact") {
		t.Fatalf("live widget output: %s", output)
	}
}
func TestSessionWidgetsRejectUnsafeOptions(t *testing.T) {
	cases := []string{
		`<screen name="demo"><toggle label="test" key="plugin.admin"/></screen>`,
		`<screen name="demo"><toggle label="test" key="state.test" command="status"/></screen>`,
		`<screen name="demo"><choice label="test" key="state.test" options="a,a"/></screen>`,
		`<screen name="demo"><choice label="test" key="state.test" options="one,\x1b[2J"/></screen>`,
	}
	for _, v := range cases {
		if _, err := Parse([]byte(v)); err == nil {
			t.Fatalf("unsafe control accepted: %s", v)
		}
	}
}
