package engine

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNativeAnimationKeepsFrameGeometry(t *testing.T) {
	for _, effect := range []string{"pulse", "sweep", "fade"} {
		v, err := Parse([]byte(`<screen name="animation" version="3"><panel title="Preview" border="round" glow="medium" animate="` + effect + `" duration-ms="1200" easing="linear"><button id="preview" label="Preview" page="home"/></panel></screen>`))
		if err != nil {
			t.Fatal(err)
		}
		f, err := v.RenderFrame(nil, 60)
		if err != nil || !v.HasAnimation() {
			t.Fatalf("%s: animation not available: %v", effect, err)
		}
		a, spots, _ := v.RenderGlowInteractive(nil, 60, GlowOptions{TrueColor: true, Animate: true})
		b, spots2, _ := v.RenderGlowInteractive(nil, 60, GlowOptions{TrueColor: true, Animate: true, Phase: 600 * time.Millisecond})
		if reflect.DeepEqual(a, b) || !reflect.DeepEqual(spots, spots2) || !reflect.DeepEqual(spots, f.Hotspots) {
			t.Fatalf("%s: animation or hotspot regression", effect)
		}
		static, err := PaintGlowFrame(f, GlowOptions{TrueColor: true})
		if err != nil {
			t.Fatal(err)
		}
		reduced, err := PaintGlowFrame(f, GlowOptions{TrueColor: true, Animate: true, ReducedMotion: true, Phase: time.Second})
		if err != nil || !reflect.DeepEqual(static, reduced) {
			t.Fatal("reduced motion changed static frame")
		}
		plain, err := PaintGlowFrame(f, GlowOptions{NoColor: true, TrueColor: true, Animate: true})
		if err != nil || !reflect.DeepEqual(plain, f.Lines) || strings.Contains(strings.Join(plain, ""), "\x1b") {
			t.Fatal("legacy output changed")
		}
	}
}

func TestNativeAnimationValidation(t *testing.T) {
	for _, source := range []string{
		`<panel animate="run"/>`,
		`<panel animate="pulse" duration-ms="0"/>`,
		`<panel animate="pulse" easing="invalid"/>`,
		`<panel duration-ms="1000"/>`,
		`<text animate="pulse">no</text>`,
	} {
		if _, err := Parse([]byte(`<screen name="bad" version="3">` + source + `</screen>`)); err == nil {
			t.Fatalf("invalid animation accepted: %s", source)
		}
	}
}
