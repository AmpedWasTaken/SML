package engine

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"
)

func TestSameGIFStaticAndAnimatedCacheDoNotCrossContaminate(t *testing.T) {
	dir := t.TempDir()
	media := filepath.Join(dir, "media")
	if err := os.MkdirAll(media, 0700); err != nil {
		t.Fatal(err)
	}
	palette := color.Palette{color.Black, color.White}
	first := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)
	second := image.NewPaletted(image.Rect(0, 0, 2, 2), palette)
	for i := range second.Pix {
		second.Pix[i] = 1
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &gif.GIF{
		Image: []*image.Paletted{first, second}, Delay: []int{5, 5},
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(media, "animation.gif"), buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	v := View{Name: "home"}.WithAssets(dir)
	// A static decode must not freeze subsequent animated rendering.
	staticFirst, err := v.mediaFrames("animation.gif", false)
	if err != nil {
		t.Fatal(err)
	}
	animatedSecond, err := v.mediaFrames("animation.gif", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(staticFirst.frames) != 1 || len(animatedSecond.frames) != 2 {
		t.Fatalf("still-first cache contamination: still=%d animated=%d", len(staticFirst.frames), len(animatedSecond.frames))
	}
	animatedAgain, err := v.mediaFrames("animation.gif", true)
	if err != nil {
		t.Fatal(err)
	}
	staticAgain, err := v.mediaFrames("animation.gif", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(animatedAgain.frames) != 2 || len(staticAgain.frames) != 1 {
		t.Fatalf("animated-first cache contamination: still=%d animated=%d", len(staticAgain.frames), len(animatedAgain.frames))
	}
}
