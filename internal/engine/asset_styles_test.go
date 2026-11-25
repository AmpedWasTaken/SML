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
