package sml

import (
	"strings"
	"testing"
)

func TestRenderDataReusesParsedTree(t *testing.T) {
	root, err := Parse([]byte("<sml><panel title=\"{{title}}\"><metric label=\"Users\" value=\"{{users}}\"/></panel></sml>"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := RenderData(root, Bindings{"title": "First", "users": "20"}, Options{Width: 40})
	if err != nil {
		t.Fatal(err)
	}
	b, err := RenderData(root, Bindings{"title": "Second", "users": "30"}, Options{Width: 40})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a, "First") || !strings.Contains(b, "Second") || root.Children[0].Attrs["title"] != "{{title}}" {
		t.Fatal("bindings did not update or original tree was mutated")
	}
}

func TestBindingSafety(t *testing.T) {
	root, err := Parse([]byte("<sml><text>{{value}}</text></sml>"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderData(root, Bindings{"value": "<exec command=\"rm\"/>\x1b[31m"}, Options{Width: 80})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\x1b") || !strings.Contains(out, "<exec") {
		t.Fatal("expected plain text output with ANSI control removed")
	}
	if _, err := RenderData(root, Bindings{}, Options{Width: 80}); err == nil {
		t.Fatal("missing binding should fail")
	}
	if _, err := RenderData(root, Bindings{"value": strings.Repeat("a", 4097)}, Options{Width: 80}); err == nil {
		t.Fatal("oversized binding should fail")
	}
}
