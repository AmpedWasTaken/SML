package engine

import "testing"

func TestCSSAttributeSelectorsAndInheritance(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		panel.workspace { color: #123456; background-color: #111111; }
		button { color: #222222; }
		button[variant="outline"] { color: inherit; background-color: inherit; }
		button[variant='solid'] { color: #333333; }
		button[variant=outline]:hover { color: #ABCDEF; }
		panel.workspace button[border][variant=outline] { color: #444444; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<panel class="workspace">
			<box>
				<button id="inherited" variant="outline" label="Inherited" action="app.open"/>
				<button id="solid" variant="solid" label="Solid" action="app.open"/>
				<button id="plain" label="Plain" action="app.open"/>
				<button id="specific" variant="outline" border="round" label="Specific" action="app.open"/>
				<button id="inline" variant="outline" foreground="#555555" label="Inline" action="app.open"/>
			</box>
		</panel>
		<button id="root" variant="outline" label="Root" action="app.open"/>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := views["home"].RenderFrame(nil, 60)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][2]string{
		"inherited": {"#123456", "#111111"},
		"solid":     {"#333333", ""},
		"plain":     {"#222222", ""},
		"specific":  {"#444444", "#111111"},
		"inline":    {"#555555", "#111111"},
		"root":      {"", ""},
	}
	for _, e := range frame.Elements {
		if expected, ok := want[e.ID]; ok {
			if got := [2]string{e.Foreground, e.Background}; got != expected {
				t.Fatalf("%s: got %v, want %v", e.ID, got, expected)
			}
			if e.ID == "inherited" && e.withControlPaint(GlowOptions{HoveredID: e.ID}).Foreground != "#ABCDEF" {
				t.Fatal("attribute hover selector did not match")
			}
			delete(want, e.ID)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing elements: %v", want)
	}
}

func TestCSSAttributeSelectorValidation(t *testing.T) {
	for _, source := range []string{
		"button[action] { color: #111111; }",
		"button[enabled-when] { color: #111111; }",
		"button[variant=missing] { color: #111111; }",
		"button[variant~=outline] { color: #111111; }",
		"button[variant=outline { color: #111111; }",
		"button[variant=outline]:hover { color: inherit; }",
		"button:hover { color: inherit; }",
		"box { width: inherit; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><box><button variant="outline" label="Open" action="app.open"/></box></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("accepted invalid selector or inherited property: %s", source)
		}
	}
}
