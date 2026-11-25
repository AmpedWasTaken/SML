package premiumwidgets

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	sml "github.com/AmpedWasTaken/SML"
)

type Route struct {
	ID, Title string
}

var routes = []Route{
	{"metrics", "Metrics"},
	{"charts", "Charts"},
	{"motion", "Motion"},
	{"status", "Status"},
	{"controls", "Controls"},
	{"layout", "Layout"},
}

func Routes() []Route { return append([]Route(nil), routes...) }

// The motion route's widget selection must not share the motion mode key.
func widgetSelectionKey(route string) string {
	if route == "motion" {
		return "state.pw_widget_motion"
	}
	return "state.pw_" + route
}

type galleryPage struct {
	view sml.View
	ids  []string
	err  error
}

var galleryOnce sync.Once
var galleryPages map[string]galleryPage

func prepareGallery() {
	galleryPages = make(map[string]galleryPage, len(routes))
	for _, route := range routes {
		var ids []string
		for _, d := range Catalog() {
			if strings.EqualFold(d.Category, route.ID) {
				ids = append(ids, d.ID)
			}
		}
		var src strings.Builder
		src.WriteString(`<screen name="premium-gallery" version="3"><column><text>PREMIUM WIDGETS / SYNTHETIC PREVIEW</text><divider variant="dotted"/>`)
		if len(ids) > 1 {
			fmt.Fprintf(&src, `<choice label="Widget" key="%s" options="%s" id="pw-widget"/>`, widgetSelectionKey(route.ID), strings.Join(ids, ","))
		}
		src.WriteString(`<choice label="Motion" key="state.pw_motion" options="Animated,Static" id="pw-motion"/><choice label="Glyphs" key="state.pw_glyphs" options="ASCII,Unicode" id="pw-glyphs"/><panel title="LIVE PREVIEW" border="single" padding="1">`)
		for i := 1; i <= Height; i++ {
			fmt.Fprintf(&src, "<text>{{preview.line%d}}</text>", i)
		}
		src.WriteString(`</panel><when field="preview.control" equals="selection"><choice label="Position" key="state.pw_position" options="0,1,2,3,4" id="pw-position"/></when><when field="preview.control" equals="toggle"><toggle label="Enabled" key="state.pw_enabled" id="pw-enabled"/></when><when field="preview.control" equals="dismiss"><toggle label="Dismiss notification" key="state.pw_dismiss" id="pw-dismiss"/></when><text>Local demo only. Tab / Enter or mouse to change controls.</text></column></screen>`)
		view, err := sml.Parse([]byte(src.String()))
		galleryPages[route.ID] = galleryPage{view: view, ids: ids, err: err}
	}
}

// Gallery returns a real SML view and detached presentation bindings. The host
// controls redraw frequency, terminal capabilities and entitlement checks.
// No real host metrics or operational actions are involved.
func Gallery(route string, width int, phase time.Duration, state map[string]string) (sml.View, map[string]string, bool, error) {
	galleryOnce.Do(prepareGallery)
	p, ok := galleryPages[route]
	if !ok {
		return sml.View{}, nil, false, fmt.Errorf("unknown gallery route %q", route)
	}
	if p.err != nil {
		return sml.View{}, nil, false, p.err
	}
	id := p.ids[0]
	key := widgetSelectionKey(route)
	for _, candidate := range p.ids {
		if state[key] == candidate {
			id = candidate
			break
		}
	}
	d, _ := Lookup(id)
	reduced := state["state.pw_motion"] == "Static"
	ascii := state["state.pw_glyphs"] != "Unicode"
	if reduced {
		phase = 0
	}
	position, _ := strconv.Atoi(state["state.pw_position"])
	position = max(0, min(4, position))
	t := phase.Seconds()
	series := make([]float64, 32)
	for i := range series {
		series[i] = 50 + 40*math.Sin(float64(i)*.57+t*1.8)
	}
	input := Input{
		Width: max(16, min(72, width-6)), Value: 60 + 25*math.Sin(t), Secondary: 47,
		Series: series, Phase: phase, ASCII: ascii, ReducedMotion: reduced,
		Selection: position, Enabled: state["state.pw_enabled"] == "on",
		Dismissed: state["state.pw_dismiss"] == "on", Label: "Sample only",
	}
	frame, err := Render(id, input)
	if err != nil {
		return sml.View{}, nil, false, err
	}
	values := frame.Bindings("preview")
	values[key] = id
	values["state.pw_position"] = strconv.Itoa(position)
	values["state.pw_motion"] = "Animated"
	if reduced {
		values["state.pw_motion"] = "Static"
	}
	values["state.pw_glyphs"] = "ASCII"
	if !ascii {
		values["state.pw_glyphs"] = "Unicode"
	}
	values["state.pw_enabled"] = "off"
	if input.Enabled {
		values["state.pw_enabled"] = "on"
	}
	values["state.pw_dismiss"] = "off"
	if input.Dismissed {
		values["state.pw_dismiss"] = "on"
	}
	switch id {
	case "toggle-switch":
		values["preview.control"] = "toggle"
	case "notification-toast":
		values["preview.control"] = "dismiss"
	default:
		if d.Interactive {
			values["preview.control"] = "selection"
		}
	}
	return p.view, values, d.Animated && !reduced, nil
}
