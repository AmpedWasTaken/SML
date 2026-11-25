package engine

import (
	"strings"
	"testing"
)

func TestSharedStylesAndTokens(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.json", `{"tokens":{"space":"1","good":"success"},"classes":{"notice":{"tone":"@good","variant":"outline"},"wide":{"padding":"@space"}}}`)
	writeSnapshotAsset(t, dir, "views/core/home.sml", `<screen name="home" version="3"><include name="header"/><badge class="notice wide" tone="warning">Local</badge></screen>`)
	writeSnapshotAsset(t, dir, "components/header.sml", `<component name="header"><badge class="notice">Ready</badge></component>`)
	store, err := NewAssetStore(dir, "home")
	if err != nil {
		t.Fatal(err)
	}
	view, _, _ := store.View("home")
	got := strings.Join(view.Render(nil), "\n")
	if !strings.Contains(got, "[OK] [ Ready ]") || !strings.Contains(got, "[!] [ Local ]") {
		t.Fatalf("unexpected style rendering: %q", got)
	}
	writeSnapshotAsset(t, dir, "styles.json", `{"classes":{"notice":{"action":"unsafe"}}}`)
	if rev, err := store.Reload(); err == nil || rev != 1 {
		t.Fatalf("unsafe stylesheet accepted: %d %v", rev, err)
	}
}

func TestUnknownStyleTokenRejected(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><text padding="@missing">Text</text></screen>`)
	if _, err := LoadDir(dir); err == nil || !strings.Contains(err.Error(), "unknown SML design token") {
		t.Fatalf("unknown token not rejected: %v", err)
	}
}

func TestSMLCSSDefaultsClassesAndInline(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "styles.smlcss", `
		/* terminal theme */
		:root { --edge: round; }
		button { border: var(--edge); }
		input { border: round; }
		.important { border: double; }
		table { style: double; }
	`)
	writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3">
		<button label="First" command="status"/>
		<button class="important" label="Second" command="status"/>
		<button class="important" border="ascii" label="Third" command="status"/>
		<input label="Name" key="state.name"/>
		<table source="rows"><column key="name"/></table>
	</screen>`)
	views, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	frame, err := views["home"].RenderFrame(map[string]string{"state.name": "Tim", "rows": `[{"name":"demo"}]`}, 40)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		line   int
		prefix string
	}{{0, "╭"}, {3, "╔"}, {6, "+"}, {9, "╭"}, {12, "╔"}} {
		if !strings.HasPrefix(frame.Lines[item.line], item.prefix) {
			t.Fatalf("incorrect frame at %d: %q", item.line, frame.Lines)
		}
	}
	if len(frame.Hotspots) != 4 {
		t.Fatalf("missing controls: %+v", frame.Hotspots)
	}
	for _, spot := range frame.Hotspots {
		if spot.Line%3 != 1 || spot.Start != 1 || spot.End != CellWidth(frame.Lines[spot.Line])-1 {
			t.Fatalf("incorrect framed control hotspot: %+v", spot)
		}
	}
}

func TestSMLCSSRejectsUnsupportedSyntax(t *testing.T) {
	for _, source := range []string{
		"button:hover { border: round; }",
		"@import url(test);",
		"button { action: unsafe; }",
		"button { border: var(--missing); }",
		"panel { border: round; border: double; }",
	} {
		dir := t.TempDir()
		writeSnapshotAsset(t, dir, "styles.smlcss", source)
		writeSnapshotAsset(t, dir, "views/home.sml", `<screen name="home" version="3"><button label="Test" command="status"/></screen>`)
		if _, err := LoadDir(dir); err == nil {
			t.Fatalf("invalid stylesheet accepted: %s", source)
		}
	}
}
