package engine

import (
	"strings"
	"testing"
)

func keyedRepeatFixture(t *testing.T) View {
	t.Helper()
	view, err := Parse([]byte(`<screen name="keyed" version="3">
  <repeat source="items" as="item" key-by="value" index-as="idx" limit="10">
    <text>{{idx}}:{{item}}</text>
    <button id="inspect" label="Inspect" action="app.inspect"/>
  </repeat>
</screen>`))
	if err != nil {
		t.Fatal(err)
	}
	return view
}

func TestKeyedRepeatRetainsIdentitiesAfterReordering(t *testing.T) {
	view := keyedRepeatFixture(t)
	first, err := view.RenderFrame(map[string]string{"items": "alpha,beta,gamma"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	second, err := view.RenderFrame(map[string]string{"items": "gamma,alpha,beta"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Hotspots) != 3 || len(second.Hotspots) != 3 {
		t.Fatalf("missing repeated controls: %+v %+v", first.Hotspots, second.Hotspots)
	}
	before := map[string]string{}
	after := map[string]string{}
	for i, value := range []string{"alpha", "beta", "gamma"} {
		before[value] = first.Hotspots[i].ID
	}
	for i, value := range []string{"gamma", "alpha", "beta"} {
		after[value] = second.Hotspots[i].ID
	}
	for _, key := range []string{"alpha", "beta", "gamma"} {
		if before[key] == "" || before[key] != after[key] {
			t.Fatalf("reordered item %q lost stable identity: %q -> %q", key, before[key], after[key])
		}
	}
	if before["alpha"] == before["beta"] || before["alpha"] == before["gamma"] {
		t.Fatal("distinct rows share an action ID")
	}
	if !strings.Contains(strings.Join(second.Lines, "|"), "0:gamma") ||
		!strings.Contains(strings.Join(second.Lines, "|"), "1:alpha") {
		t.Fatalf("index-as didn't follow the visible order: %v", second.Lines)
	}
	legacy, err := view.RenderChecked(map[string]string{"items": "gamma,alpha"}, 50)
	if err != nil || !strings.Contains(strings.Join(legacy, "|"), "1:alpha") {
		t.Fatalf("legacy renderer lost index binding: %v (%v)", legacy, err)
	}
}

func TestKeyedRepeatRejectsMissingDuplicateAndOversizedKeys(t *testing.T) {
	view := keyedRepeatFixture(t)
	for _, raw := range []string{
		"alpha,alpha",
		"alpha, alpha ",
		"alpha,",
		",beta",
		"alpha," + strings.Repeat("Z", 129),
		strings.Repeat("x", 32769),
	} {
		if _, err := view.RenderFrame(map[string]string{"items": raw}, 50); err == nil {
			t.Errorf("RenderFrame accepted invalid keyed repeat input, length=%d", len(raw))
		}
		if _, err := view.RenderChecked(map[string]string{"items": raw}, 50); err == nil {
			t.Errorf("RenderChecked accepted invalid keyed repeat input, length=%d", len(raw))
		}
	}
}

func TestKeyedRepeatV3SyntaxAndLegacyRendering(t *testing.T) {
	for _, sample := range []string{
		`<screen name="bad"><repeat source="items" key-by="value"><text>{{item}}</text></repeat></screen>`,
		`<screen name="bad" version="2"><repeat source="items" key-by="value"><text>{{item}}</text></repeat></screen>`,
		`<screen name="bad" version="3"><repeat source="items" key-by="index"><text>{{item}}</text></repeat></screen>`,
		`<screen name="bad" version="3"><repeat source="items" index-as="item"><text>{{item}}</text></repeat></screen>`,
		`<screen name="bad" version="3"><repeat source="items" index-as="state"><text>{{item}}</text></repeat></screen>`,
		`<screen name="bad" version="3"><repeat source="items" index-as="1bad"><text>{{item}}</text></repeat></screen>`,
		`<screen name="bad" version="3"><repeat source="../evil" key-by="value"><text>x</text></repeat></screen>`,
		`<screen name="bad" version="3"><column key-by="value"><text>x</text></column></screen>`,
		`<screen name="bad" version="3"><text index-as="idx">x</text></screen>`,
	} {
		if _, err := Parse([]byte(sample)); err == nil {
			t.Errorf("accepted invalid repeat syntax: %s", sample)
		}
	}
	// The v1-v3 legacy unkeyed markup must keep positional behavior.
	view, err := Parse([]byte(`<screen name="legacy"><repeat source="items" as="item"><text>{{item}}</text><button id="view" label="View" action="app.view"/></repeat></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	first, err := view.RenderFrame(map[string]string{"items": "alpha,beta"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	second, err := view.RenderFrame(map[string]string{"items": "beta,alpha"}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if first.Hotspots[0].ID != second.Hotspots[0].ID ||
		first.Hotspots[1].ID != second.Hotspots[1].ID {
		t.Fatal("legacy repeated control IDs changed from their positional semantics")
	}
}
