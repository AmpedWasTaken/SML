package sml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStandaloneParseAndRender(t *testing.T) {
	view, err := Parse([]byte(`<screen name="hello" version="3"><derive key="derived.ready" op="coalesce" source="app.status" fallback="offline"/><text>Hello {{user.name}}</text><badge>{{derived.ready}}</badge></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(view.Render(map[string]string{"user.name": "SML", "app.status": "online"}), "\n")
	if !strings.Contains(got, "Hello SML") || !strings.Contains(got, "online") {
		t.Fatalf("unexpected render: %q", got)
	}
}

func TestStandaloneAssetsInclude(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "views"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "components"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "views", "home.sml"), []byte(`<screen name="home" version="3"><include name="header"/></screen>`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "components", "header.sml"), []byte(`<component name="header"><text>Reusable component</text></component>`), 0600); err != nil {
		t.Fatal(err)
	}
	views, err := LoadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(views["home"].Render(nil), "\n")
	if !strings.Contains(got, "Reusable component") {
		t.Fatalf("got %q", got)
	}
}
