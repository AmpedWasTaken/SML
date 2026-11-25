package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspacePreviewAndDebug(t *testing.T) {
	dir := t.TempDir()
	views := filepath.Join(dir, "views", "core")
	if err := os.MkdirAll(views, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(views, "home.sml"), []byte(`<screen name="home"><text>Hello</text></screen>`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"preview", "debug"} {
		if got := run([]string{mode, "-assets", dir, "-ascii", "-height", "4", "home"}); got != 0 {
			t.Fatalf("%s exit=%d", mode, got)
		}
	}
	if got := run([]string{"preview", "-assets", dir, "missing"}); got == 0 {
		t.Fatal("missing required view accepted")
	}
}
