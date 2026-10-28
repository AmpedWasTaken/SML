package sml_test

import (
	"reflect"
	"testing"

	sml "github.com/AmpedWasTaken/SML"
)

func TestPublicBindingDependenciesCanSkipUnrelatedUpdates(t *testing.T) {
	view, err := sml.Parse([]byte("<screen name=\"deps\" version=\"3\">" +
		"<derive key=\"derived.status\" op=\"uppercase\" source=\"host.status\"/>" +
		"<text>{{derived.status}}</text>" +
		"<button label=\"Refresh\" action=\"app.refresh\" enabled-when=\"host.enabled\"/>" +
		"</screen>"))
	if err != nil {
		t.Fatal(err)
	}
	usage := view.BindingDependencies()
	var _ sml.BindingDependencies = usage
	if usage.Conservative || !reflect.DeepEqual(usage.Keys, []string{"host.enabled", "host.status"}) {
		t.Fatalf("incorrect public dependency metadata: %+v", usage)
	}
	if !usage.AffectedBy("host.status") || !usage.AffectedBy("host.enabled") {
		t.Fatal("known display or action gate did not invalidate")
	}
	if usage.AffectedBy("host.unrelated") || usage.AffectedBy() {
		t.Fatal("unrelated or empty batch was treated as dirty")
	}
}
