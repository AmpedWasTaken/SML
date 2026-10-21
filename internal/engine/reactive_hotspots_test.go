package engine

import (
	"strings"
	"testing"
)

func TestReactiveHotspotsOnlyFromActiveBranch(t *testing.T) {
	src := `<screen name="demo" version="3">
		<derive key="derived.ready" op="equals" source="plugin.health" value="online"/>
		<if field="derived.ready" equals="true">
			<button label="Open Report" command="status"/>
			<otherwise><button label="Open Report" command="help"/></otherwise>
		</if>
	</screen>`
	view, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	lines, hotspots, _ := view.RenderInteractive(map[string]string{"plugin.health": "online"}, 80)
	if len(hotspots) != 1 || hotspots[0].Command != "status" || !strings.Contains(strings.Join(lines, "\n"), "Open Report") {
		t.Fatalf("wrong online hotspot: %+v; lines=%q", hotspots, lines)
	}
	_, hotspots, _ = view.RenderInteractive(map[string]string{"plugin.health": "offline"}, 80)
	if len(hotspots) != 1 || hotspots[0].Command != "help" {
		t.Fatalf("wrong fallback hotspot: %+v", hotspots)
	}
}

func TestReactiveHotspotsHiddenEmptyAndWhen(t *testing.T) {
	src := `<screen name="demo" version="3">
		<when field="service.health"><button label="Check service" command="status"/></when>
		<empty field="service.health"><button label="Check service" command="help"/></empty>
	</screen>`
	view, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	_, spots, _ := view.RenderInteractive(map[string]string{"service.health": "false"}, 80)
	if len(spots) != 0 {
		t.Fatalf("expected no hotspot for false and nonempty binding: %+v", spots)
	}
	_, spots, _ = view.RenderInteractive(map[string]string{}, 80)
	if len(spots) != 1 || spots[0].Command != "help" {
		t.Fatalf("expected only empty fallback: %+v", spots)
	}
	_, spots, _ = view.RenderInteractive(map[string]string{"service.health": "online"}, 80)
	if len(spots) != 1 || spots[0].Command != "status" {
		t.Fatalf("expected only online conditional: %+v", spots)
	}
}
