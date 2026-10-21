package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "dashboard.sml")
	if err := os.WriteFile(path, []byte(`<screen name="dashboard"><text>Hello</text></screen>`), 0600); err != nil {
		t.Fatal(err)
	}
	view, err := ParseFile(path)
	if err != nil || view.Name != "dashboard" {
		t.Fatalf("unexpected view: %+v / %v", view, err)
	}
}

func TestParseFileRejectsUnsafeInput(t *testing.T) {
	root := t.TempDir()
	good := filepath.Join(root, "good.sml")
	if err := os.WriteFile(good, []byte(`<screen name="good"/>`), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.sml")
	if err := os.Symlink(good, link); err == nil {
		if _, err := ParseFile(link); err == nil {
			t.Fatal("accepted symbolic link")
		}
	}
	big := filepath.Join(root, "oversized.sml")
	if err := os.WriteFile(big, []byte(strings.Repeat("x", 65537)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFile(big); err == nil {
		t.Fatal("accepted oversized file")
	}
}
