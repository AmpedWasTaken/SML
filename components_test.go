package sml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComponentPropsAndSlots(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "components"), 0700); err != nil {
		t.Fatal(err)
	}
	component := `<component name="card" props="title:string,value:string">
        <panel title="{{props.title}}">
            <text>{{props.value}}</text>
            <slot name="footer"><text>Fallback footer</text></slot>
        </panel>
    </component>`
	if err := os.WriteFile(filepath.Join(root, "components", "card.sml"), []byte(component), 0600); err != nil {
		t.Fatal(err)
	}
	view, err := Parse([]byte(`<screen name="home">
        <use name="card" title="Dashboard" value="{{stats.count}}">
            <fill slot="footer"><badge>Custom footer</badge></fill>
        </use>
        <use name="card" title="Second" value="7"/>
    </screen>`))
	if err != nil {
		t.Fatal(err)
	}
	view = view.WithAssets(root)
	lines, err := view.RenderChecked(map[string]string{"stats.count": "42"}, 80)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Dashboard", "42", "Custom footer", "Second", "Fallback footer"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %q", want, joined)
		}
	}
}
func TestComponentInvalidPropsAndCycles(t *testing.T) {
	cases := []string{
		`<screen name="home"><use name="../secret"/></screen>`,
		`<screen name="home"><use name="card"><fill slot="one"/><fill slot="one"/></use></screen>`,
		`<screen name="home"><use name="card"><text>Nope</text></use></screen>`,
		`<screen name="home"><component name="x" props="a:int"/></screen>`,
	}
	for _, src := range cases {
		if _, err := Parse([]byte(src)); err == nil {
			t.Fatalf("accepted %s", src)
		}
	}
}
