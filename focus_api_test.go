package sml_test

import (
	"testing"

	sml "github.com/AmpedWasTaken/SML"
)

func TestPublicFocusNavigationMetadata(t *testing.T) {
	view, err := sml.Parse([]byte("<screen name=\"focus\"><button id=\"a\" label=\"A\" action=\"app.a\"/><button id=\"b\" label=\"B\" action=\"app.b\"/></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderFrame(nil, 40)
	if err != nil {
		t.Fatal(err)
	}
	ordered := frame.FocusOrder()
	if len(ordered) != 2 || ordered[0].ID != "a" || ordered[1].ID != "b" {
		t.Fatalf("public focus order is unstable: %+v", ordered)
	}
	var _ sml.Hotspot = ordered[0]
	if selected, ok := frame.NextFocus("a", false); !ok || selected.ID != "b" {
		t.Fatalf("public NextFocus failed: %+v, %v", selected, ok)
	}
}
