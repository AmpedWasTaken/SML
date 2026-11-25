package engine

import "testing"

func TestCSSContextualSelectors(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		button { color: #111111; }
		panel.workspace button { color: #222222; }
		panel.workspace > button { color: #333333; }
		panel.workspace > box button.primary { color: #444444; }
		#special { color: #555555; }
		button:hover { color: #666666; }
		button.primary:hover { color: #777777; }
		panel.workspace button.primary:hover { color: #888888; }
		panel.workspace > button.primary:focus { color: #999999; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<panel class="workspace">
			<button id="direct" class="primary" label="Direct" action="app.open"/>
			<box><button id="nested" class="primary" label="Nested" action="app.open"/>
			<button id="special" class="primary" label="Special" action="app.open"/></box>
		</panel>
		<button id="outside" class="primary" label="Outside" action="app.open"/>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := views["home"].RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][3]string{
		"direct":  {"#333333", "#888888", "#999999"},
		"nested":  {"#444444", "#888888", "#444444"},
		"special": {"#555555", "#888888", "#555555"},
		"outside": {"#111111", "#777777", "#111111"},
	}
	for _, element := range frame.Elements {
		if expected, ok := want[element.ID]; ok {
			got := [3]string{
				element.Foreground,
				element.withControlPaint(GlowOptions{HoveredID: element.ID}).Foreground,
				element.withControlPaint(GlowOptions{FocusedID: element.ID}).Foreground,
			}
			if got != expected {
				t.Fatalf("%s: got %v, want %v", element.ID, got, expected)
			}
			delete(want, element.ID)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing elements: %v", want)
	}
	for _, spot := range frame.Hotspots {
		got, ok := frame.HitTest(spot.Start, spot.Line)
		if !ok || got.ID != spot.ID {
			t.Fatalf("invalid hotspot: %+v", spot)
		}
	}
}

func TestCSSContextualSelectorValidation(t *testing.T) {
	for _, source := range []string{
		"panel > { color: #111111; }",
		"> button { color: #111111; }",
		"panel box row grid button { color: #111111; }",
		"button:hover button { color: #111111; }",
		"panel button:focus { action: app.open; }",
		"panel button:hover { width: 20ch; }",
		"panel.primary:hover { color: #111111; }",
	} {
		if styles, err := parseStyleCSS(source); err == nil {
			if err := validateCompoundRules(&styles); err == nil {
				t.Fatalf("invalid contextual selector accepted: %s", source)
			}
		}
	}
}
