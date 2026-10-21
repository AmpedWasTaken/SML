package sml

import "testing"

func TestHostActionIDsAndMetadata(t *testing.T) {
	view, err := Parse([]byte(`<screen name="actions">
      <button id="refresh" label="Refresh" action="metrics.refresh"/>
      <link id="open" label="Open" action="workspace.open"/>
      <button label="Legacy" command="status"/>
    </screen>`))
	if err != nil {
		t.Fatal(err)
	}
	_, spots, _ := view.RenderInteractive(nil, 80)
	if len(spots) != 3 {
		t.Fatalf("spots=%+v", spots)
	}
	if spots[0].ID != "refresh" || spots[0].Action != "metrics.refresh" || spots[0].Command != "" {
		t.Fatalf("generic action %+v", spots[0])
	}
	if spots[1].Action != "workspace.open" {
		t.Fatal(spots[1])
	}
	if spots[2].Command != "status" || spots[2].Action != "" {
		t.Fatal(spots[2])
	}
}
func TestHostActionRejectsUnsafeNames(t *testing.T) {
	cases := []string{
		`<screen name="a"><button label="Run" action=";rm -rf /"/></screen>`,
		`<screen name="a"><button label="Run" id="../bad" action="app.run"/></screen>`,
		`<screen name="a"><button label="Run" action="app.run" command="status"/></screen>`,
		`<screen name="a"><toggle label="Toggle" id="../bad" key="state.foo"/></screen>`,
	}
	for _, src := range cases {
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("accepted invalid markup %q", src)
		}
	}
}
