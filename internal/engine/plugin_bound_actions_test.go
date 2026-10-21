package engine

import "testing"

func TestBoundPluginActionHotspot(t *testing.T) {
	view, err := Parse([]byte(`<screen name="demo" version="2"><choice label="Region" key="state.region" options="eu,us"/><plugin-action label="Fetch report" plugin="server-health" action="report" arg="region" source="state.region"/></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	_, spots, _ := view.RenderInteractive(map[string]string{"state.region": "us"}, 100)
	if len(spots) != 2 {
		t.Fatalf("expected two interactive controls, got %+v", spots)
	}
	last := spots[1]
	if last.PluginID != "server-health" || last.PluginAction != "report" || last.Arg != "region" || last.Source != "state.region" {
		t.Fatalf("missing bound action: %+v", last)
	}
}
func TestBoundPluginActionRejectsInvalidBinding(t *testing.T) {
	for _, src := range []string{
		`<screen name="demo"><plugin-action label="Test" plugin="server-health" action="report" arg="host" source="plugin.host"/></screen>`,
		`<screen name="demo"><plugin-action label="Test" plugin="server-health" action="report" arg="region"/></screen>`,
		`<screen name="demo"><plugin-action label="Test" plugin="server-health" action="report" arg="region" source="state.region" command="status"/></screen>`,
		`<screen name="demo"><plugin-action label="Test" plugin="server-health" action="report" arg="region;echo" source="state.region"/></screen>`,
	} {
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("accepted invalid bound plugin action: %s", src)
		}
	}
}
