package sml

import (
	"strings"
	"testing"
)

func TestDeriveReactiveViewBindings(t *testing.T) {
	source := `<screen name="demo" version="3">
		<derive key="derived.status" op="coalesce" source="plugin.netwatch.status" fallback="offline"/>
		<derive key="derived.ready" op="equals" source="derived.status" value="online"/>
		<derive key="derived.sessions" op="sum" source="sessions.online" other="jobs.total"/>
		<derive key="derived.summary" op="concat" source="derived.status" other="derived.sessions" separator=" / "/>
		<field label="Summary" key="derived.summary"/>
		<if field="derived.ready" equals="true"><text>All systems ready</text><otherwise><text>Degraded / offline</text></otherwise></if>
	</screen>`
	view, err := Parse([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	input := map[string]string{"plugin.netwatch.status": "online", "sessions.online": "2", "jobs.total": "3", "derived.status": "injected"}
	out := strings.Join(view.RenderWidth(input, 80), "\n")
	if !strings.Contains(out, "Summary: online / 5") || !strings.Contains(out, "All systems ready") {
		t.Fatalf("reactive derived rendering failed: %q", out)
	}
	if input["derived.status"] != "injected" {
		t.Fatal("view mutated host bindings")
	}
	input["plugin.netwatch.status"] = ""
	out = strings.Join(view.RenderWidth(input, 80), "\n")
	if !strings.Contains(out, "Summary: offline / 5") || !strings.Contains(out, "Degraded / offline") || strings.Contains(out, "All systems ready") {
		t.Fatalf("reactive input invalidation failed: %q", out)
	}
}

func TestDeriveBoundedAndSanitized(t *testing.T) {
	view, err := Parse([]byte(`<screen name="demo" version="3">
		<derive key="derived.output" op="concat" source="plugin.a.first" other="plugin.a.second"/>
		<field label="Result" key="derived.output"/>
	</screen>`))
	if err != nil {
		t.Fatal(err)
	}
	values := view.ResolveBindings(map[string]string{
		"plugin.a.first":  "\x1b[2J" + strings.Repeat("A", 400),
		"plugin.a.second": "\aB",
	})
	if strings.ContainsAny(values["derived.output"], "\x1b\a") || len([]rune(values["derived.output"])) > maxDerivedValue {
		t.Fatalf("unsafe or unbounded derive: %q", values["derived.output"])
	}
}

func TestDeriveRejectsMalformedDeclarations(t *testing.T) {
	cases := []string{
		`<derive key="user.role" op="coalesce" source="user.name"/>`,
		`<derive key="derived.x" op="exec" source="server.uptime"/>`,
		`<derive key="derived.x" op="sum" source="jobs.total" other="../../secret"/>`,
		`<derive key="derived.x" op="equals" source="user.name" value="a" command="shell"/>`,
		`<panel><derive key="derived.x" op="coalesce" source="user.name"/></panel>`,
		`<derive key="derived.x" op="coalesce" source="derived.x"/>`,
		`<derive key="derived.x" op="coalesce" source="user.name"/><derive key="derived.x" op="coalesce" source="user.role"/>`,
		`<otherwise><text>Unexpected</text></otherwise>`,
		`<if field="state.x"><otherwise><text>No</text></otherwise><text>Yes</text></if>`,
	}
	for _, inner := range cases {
		source := `<screen name="demo" version="3">` + inner + `</screen>`
		if _, err := Parse([]byte(source)); err == nil {
			t.Errorf("accepted unsafe derive: %s", inner)
		}
	}
}

func TestDeriveSumsOnlyBoundedIntegers(t *testing.T) {
	v, err := Parse([]byte(`<screen name="test" version="3"><derive key="derived.total" op="sum" source="a" other="b"/></screen>`))
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ a, b, expected string }{
		{"4", "5", "9"},
		{"1000000001", "2", ""},
		{"not-a-number", "2", ""},
	} {
		result := v.ResolveBindings(map[string]string{"a": tt.a, "b": tt.b})["derived.total"]
		if result != tt.expected {
			t.Errorf("sum(%q,%q)=%q, want %q", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestSMLV3VersionOptIn(t *testing.T) {
	for _, version := range []string{"", "2"} {
		view := `<screen name="test" version="` + version + `"><derive key="derived.x" op="coalesce" source="user.name"/></screen>`
		if _, err := Parse([]byte(view)); err == nil {
			t.Errorf("accepted v3 declaration under version %q", version)
		}
	}
}

func TestSMLV3SafeFormattingOps(t *testing.T) {
	source := `<screen name="demo" version="3">
		<derive key="derived.up" op="duration" source="server.seconds"/>
		<derive key="derived.upper" op="uppercase" source="plugin.name"/>
		<derive key="derived.lower" op="lowercase" source="plugin.name"/>
		<field label="Uptime" key="derived.up"/>
		<field label="Upper" key="derived.upper"/>
		<field label="Lower" key="derived.lower"/>
	</screen>`
	v, err := Parse([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Join(v.RenderWidth(map[string]string{"server.seconds": "3670", "plugin.name": "NetWatch"}, 80), "\n")
	for _, expected := range []string{"Uptime: 1h 1m", "Upper: NETWATCH", "Lower: netwatch"} {
		if !strings.Contains(rows, expected) {
			t.Errorf("missing %s from %q", expected, rows)
		}
	}
	result := v.ResolveBindings(map[string]string{"server.seconds": "-1", "plugin.name": "NetWatch"})
	if result["derived.up"] != "" {
		t.Fatalf("negative duration accepted: %q", result["derived.up"])
	}
	result = v.ResolveBindings(map[string]string{"server.seconds": "315360001"})
	if result["derived.up"] != "" {
		t.Fatalf("unbounded duration accepted: %q", result["derived.up"])
	}
}

func TestSMLV3DurationFormat(t *testing.T) {
	for _, tt := range []struct {
		seconds  int64
		expected string
	}{
		{0, "0s"}, {42, "42s"}, {60, "1m"}, {3600, "1h"}, {3670, "1h 1m"}, {86400, "1d"}, {90000, "1d 1h"},
	} {
		if got := formatUptime(tt.seconds); got != tt.expected {
			t.Errorf("formatUptime(%d)=%q, want %q", tt.seconds, got, tt.expected)
		}
	}
}
