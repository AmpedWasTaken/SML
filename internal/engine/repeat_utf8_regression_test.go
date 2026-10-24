package engine

import (
    "strings"
    "testing"
    "unicode/utf8"
)

// Explicit IDs in a repeated template are namespaced by the repetition
// index; they should not make an otherwise valid document unrenderable.
func TestRepeatExplicitControlIDsAreUniqueStableAndHitTestable(t *testing.T) {
    view,err:=Parse([]byte(`<screen name="home" version="3">
        <repeat source="items" as="item">
            <button id="inspect" label="Inspect" page="home"/>
        </repeat>
    </screen>`))
    if err!=nil{t.Fatal(err)}
    values:=map[string]string{"items":"first,second,third"}
    frame,err:=view.RenderFrame(values,80)
    if err!=nil{t.Fatalf("repeat should not fail with duplicate visible control id: %v",err)}
    if len(frame.Hotspots)!=3 {t.Fatalf("expected three hotspots, got %d",len(frame.Hotspots))}
    distinct:=map[string]bool{}
    for _,spot:=range frame.Hotspots {
        if distinct[spot.ID]{t.Fatalf("repeated ID %q",spot.ID)}
        distinct[spot.ID]=true
        if spot.Line<0 || spot.Line>=len(frame.Cells) {t.Fatalf("bad hotspot row: %+v",spot)}
        for x:=spot.Start;x<spot.End;x++{
            if frame.Cells[spot.Line][x].ElementID!=spot.ID {
                t.Fatalf("cell %d,%d is not interactive for %q",x,spot.Line,spot.ID)
            }
        }
    }
    again,err:=view.RenderFrame(values,80)
    if err!=nil{t.Fatal(err)}
    for i:=range frame.Hotspots {
        if frame.Hotspots[i].ID!=again.Hotspots[i].ID {
            t.Fatal("unstable hotspot identity across redraws")
        }
    }
}

func TestLongInterpolationKeepsUTF8RuneBoundary(t *testing.T) {
    value:=strings.Repeat("界",1366) // 4098 bytes; the 4096 cutoff bisects a rune
    rendered:=interpolate("{{item}}",map[string]string{"item":value})
    if !utf8.ValidString(rendered) {
        t.Fatalf("interpolation produced malformed UTF-8")
    }
    if strings.ContainsRune(rendered,utf8.RuneError) {
        t.Fatalf("interpolation introduced a replacement glyph at its cutoff")
    }
    if got:=len(rendered);got!=4095 {
        t.Fatalf("safe truncation should stop at a complete rune: got %d bytes",got)
    }
    short:=interpolate("{{item}}",map[string]string{"item":"🛠️ Ready"})
    if short!="🛠️ Ready" {
        t.Fatalf("short Unicode binding changed unexpectedly: %q",short)
    }
}
