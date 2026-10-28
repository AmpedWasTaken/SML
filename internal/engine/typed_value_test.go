package engine

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestTypedValueKindsAndAccessors(t *testing.T) {
	cases := []struct {
		value Value
		kind  ValueKind
		text  string
	}{
		{NullValue(), KindNull, ""},
		{StringValue("界 👩‍💻"), KindString, "界 👩‍💻"},
		{IntValue(-42), KindInt, "-42"},
		{FloatValue(1.25), KindFloat, "1.25"},
		{BoolValue(true), KindBool, "true"},
		{BoolValue(false), KindBool, "false"},
	}
	for _, sample := range cases {
		if sample.value.Kind() != sample.kind || sample.value.String() != sample.text {
			t.Fatalf("kind or rendering mismatch: %+v", sample)
		}
	}
	if s, ok := StringValue("hello").Text(); !ok || s != "hello" {
		t.Fatal("string accessor is not typed")
	}
	if _, ok := IntValue(1).Text(); ok {
		t.Fatal("integer exposed itself as string")
	}
	if i, ok := IntValue(-5).Int64(); !ok || i != -5 {
		t.Fatal("integer accessor failed")
	}
	if f, ok := FloatValue(2.5).Float64(); !ok || f != 2.5 {
		t.Fatal("float accessor failed")
	}
	if b, ok := BoolValue(false).Bool(); !ok || b {
		t.Fatal("boolean accessor failed")
	}
	if NullValue() != (Value{}) {
		t.Fatal("zero value is not null")
	}
}

func TestTypedFramePreservesLegacyRenderingAndDeriveProtection(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"typed\" version=\"3\">\n" +
		"<derive key=\"derived.health\" source=\"service.status\" op=\"uppercase\"/>" +
		"<text>Health {{derived.health}} / {{metrics.requests}} / {{metrics.cpu}}</text>" +
		"<button id=\"refresh\" label=\"Refresh\" action=\"app.refresh\" enabled-when=\"feature.enabled\"/>" +
		"</screen>"))
	if err != nil {
		t.Fatal(err)
	}
	typed := map[string]Value{
		"service.status": StringValue("ready"),
		"metrics.requests": IntValue(42),
		"metrics.cpu": FloatValue(0.75),
		"feature.enabled": BoolValue(true),
	}
	before := copyBindingValues(typed)
	frame, err := view.RenderTypedFrame(typed, 60)
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := view.RenderFrame(map[string]string{
		"service.status": "ready", "metrics.requests": "42",
		"metrics.cpu": "0.75", "feature.enabled": "true",
	}, 60)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(frame.Lines, legacy.Lines) || !reflect.DeepEqual(frame.Hotspots, legacy.Hotspots) {
		t.Fatalf("typed values changed v3 semantics: %q / %q", frame.Lines, legacy.Lines)
	}
	if !strings.Contains(strings.Join(frame.Lines, "|"), "READY / 42 / 0.75") {
		t.Fatalf("derived data was not rendered: %v", frame.Lines)
	}
	if !reflect.DeepEqual(typed, before) {
		t.Fatal("RenderTypedFrame modified the caller's data")
	}
	typed["derived.health"] = StringValue("spoofed")
	if _, err := view.RenderTypedFrame(typed, 60); err == nil {
		t.Fatal("untrusted derived namespace override accepted")
	}
}

func TestTypedBindingValidationAndUnsafeDisplayData(t *testing.T) {
	view, err := Parse([]byte("<screen name=\"safe\"><text>{{message}}</text></screen>"))
	if err != nil {
		t.Fatal(err)
	}
	frame, err := view.RenderTypedFrame(map[string]Value{"message": StringValue("safe\x1b[31m red")}, 40)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(frame.Lines, ""), "\x1b") {
		t.Fatal("terminal control sequence escaped through typed interpolation")
	}
	bad := []map[string]Value{
		{"derived.inject": StringValue("yes")},
		{"user name": StringValue("yes")},
		{"malformed": StringValue(string([]byte{0xff, 0xfe}))},
		{"oversize": StringValue(strings.Repeat("x", 4097))},
		{"nan": FloatValue(math.NaN())},
		{"infinite": FloatValue(math.Inf(1))},
		{"too-large": Value{kind: ValueKind(250)}},
	}
	for _, values := range bad {
		if err := ValidateBindingValues(values); err == nil {
			t.Fatalf("invalid typed values were accepted: %+v", values)
		}
		if _, err := view.RenderTypedFrame(values, 40); err == nil {
			t.Fatalf("invalid values rendered: %+v", values)
		}
	}
	many := make(map[string]Value)
	for i := 0; i < 257; i++ {
		many["x."+strconv.Itoa(i)] = IntValue(int64(i))
	}
	if err := ValidateBindingValues(many); err == nil {
		t.Fatal("accepted more than 256 presentation bindings")
	}
	large := make(map[string]Value)
	for i := 0; i < 40; i++ {
		large["bulk."+strconv.Itoa(i)] = StringValue(strings.Repeat("q", 4000))
	}
	if err := ValidateBindingValues(large); err == nil {
		t.Fatal("accepted more than 128 KiB aggregate bindings")
	}
}
