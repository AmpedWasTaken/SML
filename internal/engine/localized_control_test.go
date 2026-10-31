package engine

import (
 "strings"
 "testing"
)

func TestInteractiveLocalizationUsesHostBindingsAndPreservesAction(t *testing.T) {
 source := `<screen name="localization" version="3"><column><button id="next" label="NEXT" label-key="i18n.navigation.next" page="status"/><link id="help" label="HELP" label-key="i18n.navigation.help" command="help"/></column></screen>`
 view,err:=Parse([]byte(source))
 if err!=nil{t.Fatalf("parse localized controls: %v",err)}
 for _,tc:=range []struct{label string;values map[string]string}{
  {"下一步",map[string]string{"i18n.navigation.next":"下一步","i18n.navigation.help":"帮助"}},
  {"Далее",map[string]string{"i18n.navigation.next":"Далее","i18n.navigation.help":"Помощь"}},
  {"NEXT",nil},
 }{
  frame,e:=view.RenderFrame(tc.values,80)
  if e!=nil{t.Fatalf("render %q: %v",tc.label,e)}
  if len(frame.Hotspots)!=2{t.Fatalf("expected 2 controls, got %d",len(frame.Hotspots))}
  control:=frame.Hotspots[0]
  if control.Label!=tc.label||control.Page!="status"||control.Command!=""{
   t.Errorf("control label or action changed: %+v",control)
  }
  if !strings.Contains(frame.Lines[control.Line],tc.label){
   t.Fatalf("visible text differs from hotspot metadata: %q / %+v",frame.Lines[control.Line],control)
  }
  if control.Start>=control.End||control.End>80{t.Fatalf("invalid translated hotspot bounds: %+v",control)}
  midpoint:=(control.Start+control.End)/2
  selected,ok:=frame.HitTest(midpoint,control.Line)
  if !ok||selected.ID!=control.ID{t.Fatalf("translated label hit test failed: got %+v (ok=%v)",selected,ok)}
  }
}

func TestLocalizedControlLabelsRejectNonI18nBindings(t *testing.T){
 for _,key:=range []string{"user.name","state.secret","plugin.command","i18n.{{secret}}","i18n.test\nbad","i18n.",strings.Repeat("i18n.long.",20)}{
  source:=`<screen name="test" version="3"><button label="SAFE" label-key="`+key+`" command="help"/></screen>`
  if _,err:=Parse([]byte(source));err==nil{t.Fatalf("accepted unsafe label-key %q",key)}
 }
}

func TestInteractiveLocaleFallbackAndCellBounds(t *testing.T){
 view,err:=Parse([]byte(`<screen name="test" version="3"><button label="GO BACK" label-key="i18n.nav.back" command="help"/></screen>`))
 if err!=nil{t.Fatal(err)}
 for _,value:=range []string{"",strings.Repeat("安全",100),"\x1b[31m SECRET \x1b[0m"} {
  frame,e:=view.RenderFrame(map[string]string{"i18n.nav.back":value},24)
  if e!=nil{t.Fatalf("render bounded label: %v",e)}
  if len(frame.Lines)==0||len(frame.Hotspots)!=1{t.Fatalf("missing visible control: %+v",frame)}
  if CellWidth(frame.Lines[0])>24{t.Fatalf("localized label overflow: %q",frame.Lines[0])}
  if frame.Hotspots[0].End>24{t.Fatal("hotspot exceeds viewport")}
  if value==""&&frame.Hotspots[0].Label!="GO BACK"{t.Fatal("English fallback was not preserved")}
  if strings.Contains(frame.Lines[0],"\x1b"){t.Fatal("ANSI control sequence leaked")}
 }
}
