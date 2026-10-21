package engine

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInteractiveSMLGeometryAndSafety(t *testing.T) {
	data := []byte(`<screen name="home"><column><button label="Open Jobs" page="jobs"/><link label="Check Status" command="status"/></column></screen>`)
	view, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	lines, spots, animated := view.RenderInteractive(nil, 80)
	if animated || len(spots) != 2 || len(lines) != 2 {
		t.Fatalf("interactive output: %+v %v", spots, lines)
	}
	if spots[0].Page != "jobs" || spots[0].Line != 0 || spots[1].Command != "status" || spots[1].Line != 1 {
		t.Fatalf("unexpected geometry: %+v", spots)
	}
	for _, sample := range []string{
		`<screen name="home"><button label="Danger" command="jobs start foo"/></screen>`,
		`<screen name="home"><button label="Shell" command="sh"/></screen>`,
		`<screen name="home"><button label="Invalid" page="../root"/></screen>`,
		`<screen name="home"><button label="Unknown"/></screen>`,
		`<screen name="home"><button label="No" page="jobs" command="status"/></screen>`,
	} {
		if _, err := Parse([]byte(sample)); err == nil {
			t.Fatalf("unsafe control accepted: %s", sample)
		}
	}
}

func TestLocalImageAndGIFRenderInTextCells(t *testing.T) {
	root := t.TempDir()
	media := filepath.Join(root, "media")
	if err := os.MkdirAll(media, 0700); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 64), uint8(y * 64), 0, 255})
		}
	}
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(media, "example.png"), pngData.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	pal := color.Palette{color.Black, color.White}
	first := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	second := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			first.SetColorIndex(x, y, 0)
			second.SetColorIndex(x, y, 1)
		}
	}
	var gifData bytes.Buffer
	if err := gif.EncodeAll(&gifData, &gif.GIF{Image: []*image.Paletted{first, second}, Delay: []int{5, 5}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(media, "example.gif"), gifData.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	for _, sample := range []string{
		`<screen name="home"><image src="example.png" width="8" height="4"/></screen>`,
		`<screen name="home"><gif src="example.gif" width="8" height="4" fps="8"/></screen>`,
	} {
		view, err := Parse([]byte(sample))
		if err != nil {
			t.Fatal(err)
		}
		view = view.WithAssets(root)
		lines := view.RenderWidth(nil, 80)
		if len(lines) != 4 || strings.Contains(strings.Join(lines, ""), "SML render error") {
			t.Fatalf("media render failed: %v", lines)
		}
	}
	for _, sample := range []string{
		`<screen name="home"><image src="../etc/passwd"/></screen>`,
		`<screen name="home"><gif src="example.png"/></screen>`,
		`<screen name="home"><image src="x.png" width="1000000"/></screen>`,
	} {
		if _, err := Parse([]byte(sample)); err == nil {
			t.Fatalf("accepted unsafe media: %s", sample)
		}
	}
}
