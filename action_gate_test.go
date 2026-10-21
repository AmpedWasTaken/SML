package sml

import (
	"strings"
	"testing"
)

func TestSMLV3ActionEnabledWhenDerivedValue(t *testing.T) {
	raw := `<screen name="demo" version="3">
  <derive key="derived.ready" op="equals" source="plugin.health" value="online"/>
  <plugin-action label="Fetch report" plugin="server-health" action="report" enabled-when="derived.ready"/>
 </screen>`
	view, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	lines, spots, _ := view.RenderInteractive(map[string]string{"plugin.health": "offline"}, 80)
	if len(spots) != 0 {
		t.Fatalf("disabled action is clickable: %+v", spots)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "[ Fetch report (disabled) ]") {
		t.Fatalf("missing disabled state: %q", lines)
	}
	lines, spots, _ = view.RenderInteractive(map[string]string{"plugin.health": "online"}, 80)
	if len(spots) != 1 || spots[0].PluginID != "server-health" || spots[0].PluginAction != "report" {
		t.Fatalf("enabled action has incorrect metadata: %+v", spots)
	}
	if strings.Contains(strings.Join(lines, "\n"), "disabled") {
		t.Fatal("enabled action rendered disabled")
	}
}

func TestSMLV3EnabledWhenBoundSessionState(t *testing.T) {
	raw := `<screen name="demo" version="3">
  <toggle label="Allow report" key="state.ready"/>
  <button label="Inspect" command="status" enabled-when="state.ready"/>
 </screen>`
	view, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value string
		count int
	}{
		{"", 1}, {"off", 1}, {"false", 1}, {"0", 1}, {"on", 2}, {"true", 2},
	} {
		_, spots, _ := view.RenderInteractive(map[string]string{"state.ready": tc.value}, 100)
		if len(spots) != tc.count {
			t.Errorf("state %q: got %d hotspots, want %d", tc.value, len(spots), tc.count)
		}
	}
}

func TestSMLV3ActionGateDoesNotMatchDisabledDuplicate(t *testing.T) {
	raw := `<screen name="demo" version="3">
  <button label="Inspect" command="help" enabled-when="state.denied"/>
  <button label="Inspect" command="status"/>
 </screen>`
	view, err := Parse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	lines, spots, _ := view.RenderInteractive(map[string]string{}, 80)
	if len(spots) != 1 || spots[0].Command != "status" {
		t.Fatalf("disabled duplicate became actionable: %+v (%q)", spots, lines)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Inspect (disabled)") {
		t.Fatal("disabled label missing")
	}
}

func TestSMLV3RejectsInvalidEnabledWhen(t *testing.T) {
	cases := []string{
		`<screen name="demo" version="2"><button label="Inspect" command="status" enabled-when="state.ready"/></screen>`,
		`<screen name="demo" version="2"><button label="Inspect" command="status" enabled-when=""/></screen>`,
		`<screen name="demo" version="3"><button label="Inspect" command="status" enabled-when=""/></screen>`,
		`<screen name="demo" version="3"><button label="Inspect" command="status" enabled-when="../admin"/></screen>`,
		`<screen name="demo" version="3"><toggle label="Flag" key="state.flag" enabled-when="state.ok"/></screen>`,
		`<screen name="demo" version="3"><plugin-action label="Test" plugin="demo" action="status" enabled-when="${cmd}"/></screen>`,
	}
	for _, input := range cases {
		if _, err := Parse([]byte(input)); err == nil {
			t.Errorf("accepted invalid action gate: %s", input)
		}
	}
}
