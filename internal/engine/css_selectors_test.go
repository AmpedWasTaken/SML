package engine

import "testing"

func TestCSSCompoundCascade(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		button { color: #111111; }
		.primary { color: #222222; }
		.compact { color: #333333; }
		button.primary { color: #444444; }
		.primary.compact { color: #555555; }
		.primary.compact { color: #666666; }
		#save { color: #777777; }
		button#save { color: #888888; }
		button.primary { color: #999999; }
		button.marker { color: #AAAAAA; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<button id="save" class="primary compact" label="Save" action="app.save"/>
		<button id="both" class="primary compact" label="Both" action="app.open"/>
		<button id="single" class="primary" label="Single" action="app.open"/>
		<button id="inline" class="primary compact" foreground="#BBBBBB" label="Inline" action="app.open"/>
		<button id="marker" class="marker" label="Marker" action="app.open"/>
		<button id="plain" label="Plain" action="app.open"/>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := views["home"].RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"save": "#888888", "both": "#666666", "single": "#999999",
		"inline": "#BBBBBB", "marker": "#AAAAAA", "plain": "#111111",
	}
	for _, element := range frame.Elements {
		if expected, ok := want[element.ID]; ok {
			if element.Foreground != expected {
				t.Fatalf("%s: got %s, want %s", element.ID, element.Foreground, expected)
			}
			delete(want, element.ID)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing rendered elements: %v", want)
	}
}

func TestCSSCompoundRejectsInvalidRules(t *testing.T) {
	for _, source := range []string{
		"button..primary { color: #111111; }",
		"button#one#two { color: #111111; }",
		"button.primary.primary { color: #111111; }",
		"button.primary { action: app.open; }",
		"button.primary { width: 20ch; }",
		"panel > > button.primary { color: #111111; }",
		"button.primary:hover { action: app.open; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><button class="primary" label="Open" action="app.open"/></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("accepted invalid selector: %s", source)
		}
	}
}
