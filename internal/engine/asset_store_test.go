package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSnapshotAsset(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestAssetStoreRetainsLastGoodComponent(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "views/core/home.sml", `<screen name="home" version="3"><include name="header"/></screen>`)
	writeSnapshotAsset(t, dir, "components/header.sml", `<component name="header"><text>Before</text></component>`)
	store, err := NewAssetStore(dir, "home")
	if err != nil {
		t.Fatal(err)
	}
	before, rev, ok := store.View("home")
	if !ok || rev != 1 {
		t.Fatal("missing initial revision")
	}
	writeSnapshotAsset(t, dir, "components/header.sml", `<component name="header"><unknown/></component>`)
	if rev, err := store.Reload(); err == nil || rev != 1 {
		t.Fatalf("invalid reload: %d %v", rev, err)
	}
	current, _, _ := store.View("home")
	if got := strings.Join(current.Render(nil), "\n"); !strings.Contains(got, "Before") {
		t.Fatalf("snapshot changed: %q", got)
	}
	writeSnapshotAsset(t, dir, "components/header.sml", `<component name="header"><text>After</text></component>`)
	if rev, err := store.Reload(); err != nil || rev != 2 {
		t.Fatalf("valid reload: %d %v", rev, err)
	}
	current, _, _ = store.View("home")
	if got := strings.Join(current.Render(nil), "\n"); !strings.Contains(got, "After") {
		t.Fatalf("new revision missing: %q", got)
	}
	if got := strings.Join(before.Render(nil), "\n"); !strings.Contains(got, "Before") {
		t.Fatalf("old revision changed: %q", got)
	}
	current.Nodes[0].Attrs["name"] = "missing"
	current, _, _ = store.View("home")
	if current.Nodes[0].Attr("name", "") != "header" {
		t.Fatal("caller mutated stored view")
	}
	if err := os.Remove(filepath.Join(dir, "views/core/home.sml")); err != nil {
		t.Fatal(err)
	}
	if rev, err := store.Reload(); err == nil || rev != 2 {
		t.Fatalf("required view removal accepted: %d %v", rev, err)
	}
}

func TestNestedViewDuplicateRejected(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{"views/core/home.sml", "views/other/home.sml"} {
		writeSnapshotAsset(t, dir, path, `<screen name="home"><text>Home</text></screen>`)
	}
	if _, err := LoadDir(dir); err == nil || !strings.Contains(err.Error(), "duplicate SML view") {
		t.Fatalf("duplicate not reported: %v", err)
	}
}

func TestInvalidComponentPropsFailAtLoad(t *testing.T) {
	dir := t.TempDir()
	writeSnapshotAsset(t, dir, "views/core/home.sml", `<screen name="home" version="3"><use name="header"/></screen>`)
	writeSnapshotAsset(t, dir, "components/header.sml", `<component name="header" props="title"><text>{{props.title}}</text></component>`)
	if _, err := LoadDir(dir); err == nil || !strings.Contains(err.Error(), "missing prop") {
		t.Fatalf("missing prop not reported: %v", err)
	}
}
