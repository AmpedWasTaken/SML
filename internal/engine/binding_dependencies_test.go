package engine

import (
	"reflect"
	"strings"
	"testing"
)

func TestViewBindingDependenciesDerivedAndConditionalSources(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"dependencies\" version=\"3\">" +
		"<derive key=\"derived.health\" source=\"service.status\" op=\"uppercase\"/>" +
		"<derive key=\"derived.unused\" source=\"unused.source\" op=\"uppercase\"/>" +
		"<text>Health {{derived.health}}</text>" +
		"<field key=\"metrics.requests\" label=\"Requests\"/>" +
		"<if field=\"feature.ready\" equals=\"true\">" +
		"<button id=\"retry\" label=\"Retry\" action=\"service.retry\" enabled-when=\"feature.enabled\"/>" +
		"<otherwise><text>{{fallback.message}}</text></otherwise>" +
		"</if>" +
		"<repeat source=\"items\" as=\"item\" limit=\"3\"><text>{{item}}</text></repeat>" +
		"</screen>"))
	if err != nil {
		t.Fatal(err)
	}
	usage := view.BindingDependencies()
	want := []string{
		"fallback.message", "feature.enabled", "feature.ready", "item",
		"items", "metrics.requests", "service.status",
	}
	if usage.Conservative || !reflect.DeepEqual(usage.Keys, want) {
		t.Fatalf("unexpected binding graph: %+v", usage)
	}
	for _, key := range want {
		if !usage.AffectedBy(key) {
			t.Fatalf("known source %q did not trigger redraw", key)
		}
	}
	if usage.AffectedBy("unused.source") || usage.AffectedBy("other.key") || usage.AffectedBy() {
		t.Fatal("unrelated value invalidated static frame")
	}
	// Recomputing the read-only usage descriptor must remain deterministic.
	if !reflect.DeepEqual(usage, view.BindingDependencies()) {
		t.Fatal("dependency extraction changed across identical reads")
	}
}

func TestViewBindingDependenciesTransitiveDerivedSources(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"chain\" version=\"3\">" +
		"<derive key=\"derived.first\" source=\"server.health\" op=\"uppercase\"/>" +
		"<derive key=\"derived.summary\" source=\"derived.first\" other=\"server.count\" op=\"concat\" separator=\" \"/>" +
		"<text>{{derived.summary}}</text></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	usage := view.BindingDependencies()
	if usage.Conservative || !reflect.DeepEqual(usage.Keys, []string{"server.count", "server.health"}) {
		t.Fatalf("transitive derived sources missing: %+v", usage)
	}
	if !usage.AffectedBy("server.health") || !usage.AffectedBy("server.count") || usage.AffectedBy("server.unused") {
		t.Fatalf("derived invalidation incorrect: %+v", usage)
	}
}

func TestBindingDependenciesConservativeForExternalComponents(t *testing.T) {
	for _, source := range []string{
		"<screen name=\"included\"><include name=\"header\"/></screen>",
		"<screen name=\"used\" version=\"3\"><use name=\"header\"/></screen>",
	} {
		view, err := Parse([]byte(source))
		if err != nil {
			t.Fatal(err)
		}
		usage := view.BindingDependencies()
		if !usage.Conservative || !usage.AffectedBy("any.binding") || usage.AffectedBy() {
			t.Fatalf("external component was not conservative: %+v", usage)
		}
	}
	view := View{Nodes: []Node{{Kind: "unvalidated-extension", Text: "{{host.value}}"}}}
	if !view.BindingDependencies().Conservative {
		t.Fatal("unknown manually constructed component must be conservative")
	}
}

func TestBindingDependenciesAgreeWithRenderedChanges(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"monitor\" version=\"3\">" +
		"<derive key=\"derived.display\" source=\"host.status\" op=\"uppercase\"/>" +
		"<text>Host: {{derived.display}}</text>" +
		"<button label=\"Restart\" action=\"host.restart\" enabled-when=\"host.enabled\"/>" +
		"</screen>"))
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]string{"host.status": "ready", "host.enabled": "true", "unrelated.counter": "1"}
	after := map[string]string{"host.status": "ready", "host.enabled": "true", "unrelated.counter": "2"}
	first, err := view.RenderFrame(before, 40)
	if err != nil {
		t.Fatal(err)
	}
	second, err := view.RenderFrame(after, 40)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Lines, second.Lines) || !reflect.DeepEqual(first.Hotspots, second.Hotspots) {
		t.Fatal("the supposedly unrelated binding changed visible content")
	}
	usage := view.BindingDependencies()
	if usage.AffectedBy("unrelated.counter") || !usage.AffectedBy("host.status") ||
		!usage.AffectedBy("host.enabled") {
		t.Fatalf("wrong conservative dependency decision: %+v", usage)
	}
	after["host.status"] = "down"
	third, err := view.RenderFrame(after, 40)
	if err != nil || !strings.Contains(strings.Join(third.Lines, ""), "DOWN") {
		t.Fatalf("derived binding failed to update: %v %v", third.Lines, err)
	}
	after["host.enabled"] = "false"
	fourth, err := view.RenderFrame(after, 40)
	if err != nil || len(fourth.Hotspots) != 0 {
		t.Fatalf("focus gate failed to update: %+v %v", fourth.Hotspots, err)
	}
}
