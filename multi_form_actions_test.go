package sml

import "testing"

func TestMultiFormBindings(t *testing.T) {
	bindings, err := ParseActionBindings("region:state.region,period:state.period")
	if err != nil || bindings["region"] != "state.region" || bindings["period"] != "state.period" {
		t.Fatalf("bad bindings: %v %+v", err, bindings)
	}
	v, err := Parse([]byte(`<screen name="reports" version="2"><plugin-action label="Report" plugin="server-health" action="report" bindings="region:state.region,period:state.period"/></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	_, hotspots, _ := v.RenderInteractive(nil, 90)
	if len(hotspots) != 1 || hotspots[0].Bindings != "region:state.region,period:state.period" {
		t.Fatalf("missing form hotspot: %+v", hotspots)
	}
}
func TestMultiFormBindingsRejectInvalid(t *testing.T) {
	for _, spec := range []string{
		"region:plugin.secret", "region:state.region,region:state.second",
		"region:state.region,period:state.region",
		"a:state.a,b:state.b,c:state.c,d:state.d,e:state.e",
		"region:state.region;status",
	} {
		if _, err := ParseActionBindings(spec); err == nil {
			t.Errorf("accepted %q", spec)
		}
	}
	for _, source := range []string{
		`<screen name="reports"><plugin-action label="Report" plugin="server-health" action="report" bindings="region:state.region" arg="region" source="state.region"/></screen>`,
		`<screen name="reports"><plugin-action label="Report" plugin="server-health" action="report" bindings="ip:plugin.secret"/></screen>`,
	} {
		if _, err := Parse([]byte(source)); err == nil {
			t.Errorf("accepted %s", source)
		}
	}
}
