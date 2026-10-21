package sml

import "testing"

// FuzzParseNeverPanics checks that arbitrary markup does not panic the
// parser or the bounded terminal renderer.
func FuzzParseNeverPanics(f *testing.F) {
    f.Add([]byte(`<screen name="home"><text>hello</text></screen>`))
    f.Add([]byte(`<screen name="test" version="3"><derive key="derived.x" op="coalesce" source="x" fallback="off"/></screen>`))
    f.Add([]byte(`<!DOCTYPE s><screen name="bad"/>`))
    f.Fuzz(func(t *testing.T, input []byte) {
        if len(input)>65536 { t.Skip() }
        view,err:=Parse(input)
        if err==nil { _,_=view.RenderChecked(nil,80) }
    })
}

func TestRenderCheckedSurfacesMissingAsset(t *testing.T) {
    view,err:=Parse([]byte(`<screen name="home"><include name="missing"/></screen>`))
    if err!=nil { t.Fatal(err) }
    if _,err:=view.RenderChecked(nil,80);err==nil { t.Fatal("expected render-time missing asset error") }
}
