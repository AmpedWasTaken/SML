package sml

import (
 "strings"
 "testing"
 "unicode/utf8"
)

func TestSMLInlineStylingRendersPanelAndText(t *testing.T) {
 source:=`<screen name="demo" version="3">
  <panel title="Operator Panel" border="ascii" padding="2">
   <text align="right" case="upper" padding="1">active session</text>
   <badge tone="success" variant="outline">online</badge>
  </panel>
 </screen>`
 view,err:=Parse([]byte(source))
 if err!=nil{t.Fatal(err)}
 rows:=view.RenderWidth(nil,48)
 joined:=strings.Join(rows,"\n")
 for _,expected:=range []string{"+","Operator Panel","ACTIVE SESSION","[OK] [ online ]"} {
  if !strings.Contains(joined,expected){t.Errorf("expected %q in %q",expected,joined)}
 }
 for _,line:=range rows{
  if utf8.RuneCountInString(line)>48{t.Errorf("panel line exceeds width: %d: %q",utf8.RuneCountInString(line),line)}
 }
 if !strings.HasPrefix(rows[0],"+") || !strings.HasSuffix(rows[0],"+") {
  t.Fatalf("panel border not applied: %q",rows[0])
 }
}

func TestSMLStyleButtonVariantKeepsHotspot(t *testing.T) {
 source:=`<screen name="demo" version="3">
  <button variant="solid" case="upper" padding="1" label="Quick status" command="status"/>
 </screen>`
 view,err:=Parse([]byte(source))
 if err!=nil{t.Fatal(err)}
 lines,spots,_:=view.RenderInteractive(nil,80)
 if len(spots)!=1 || spots[0].Command!="status" {t.Fatalf("missing styled hotspot %+v",spots)}
 visible:=strings.Join(lines,"\n")
 if !strings.Contains(visible,"<< QUICK STATUS >>"){t.Fatalf("unexpected action style %q",visible)}
 if !(spots[0].End>spots[0].Start){t.Fatalf("hotspot has zero width: %+v",spots[0])}
}

func TestSMLStylePanelBorderVariants(t *testing.T) {
 borders:=map[string][2]string{
  "round":{"╭","╰"},
  "single":{"┌","└"},
  "double":{"╔","╚"},
  "heavy":{"┏","┗"},
  "ascii":{"+","+"},
 }
 for name,expect:=range borders {
  source:=`<screen name="demo" version="3"><panel title="Status" border="`+name+`" padding="0"><text>ok</text></panel></screen>`
  view,err:=Parse([]byte(source))
  if err!=nil{t.Fatal(err)}
  rows:=view.RenderWidth(nil,40)
  if len(rows)!=3 || !strings.HasPrefix(rows[0],expect[0]) || !strings.HasPrefix(rows[2],expect[1]) {
   t.Errorf("%s: bad panel output %#v",name,rows)
  }
  for _,line:=range rows{if utf8.RuneCountInString(line)!=40{t.Errorf("%s: panel width %d (want 40)",name,utf8.RuneCountInString(line))}}
 }
}
func TestSMLStyleDividerVariants(t *testing.T){
 for _,variant:=range []string{"dashed","dotted","double","heavy","ascii"}{
  view,err:=Parse([]byte(`<screen name="demo" version="3"><divider variant="`+variant+`"/></screen>`))
  if err!=nil{t.Fatal(err)}
  lines:=view.RenderWidth(nil,33)
  if len(lines)!=1||utf8.RuneCountInString(lines[0])!=33{t.Errorf("divider %q: %#v",variant,lines)}
 }
}
func TestSMLStyleRejectsDeclarationsAndUnsafeAttrs(t *testing.T){
 examples:=[]string{
  `<screen name="x" version="2"><style name="card" border="ascii"/></screen>`,
  `<screen name="x" version="3"><style name="card" border="ascii"/></screen>`,
  `<screen name="x" version="3"><panel style="card"><text>x</text></panel></screen>`,
  `<screen name="x" version="3"><panel title="Bad" border="glass"><text>x</text></panel></screen>`,
  `<screen name="x" version="3"><panel title="Bad" padding="999"><text>x</text></panel></screen>`,
  `<screen name="x" version="3"><panel title="Bad" color="javascript:alert(1)"><text>x</text></panel></screen>`,
  `<screen name="x" version="2"><panel title="Bad" border="ascii"><text>x</text></panel></screen>`,
  `<screen name="x" version="3"><button label="Open" command="status" padding="5"/></screen>`,
  `<screen name="x" version="3"><button label="Open" command="status" style="primary"/></screen>`,
 }
 for _,source:=range examples {
  if _,err:=Parse([]byte(source));err==nil{t.Errorf("unsafe SML style accepted: %s",source)}
 }
}

func TestInlineStylePanelDirectly(t *testing.T){
 view,err:=Parse([]byte(`<screen name="demo" version="3">
  <panel title="Inline styling" border="ascii" padding="0" case="upper"><text>ok</text></panel>
 </screen>`))
 if err!=nil{t.Fatal(err)}
 rows:=view.RenderWidth(nil,35)
 if !strings.HasPrefix(rows[0],"+")||!strings.HasPrefix(rows[1],"|ok"){
  t.Fatalf("inline styling not applied %#v",rows)
 }
 if !strings.Contains(rows[0],"INLINE STYLING"){
  t.Fatalf("inline text casing ignored %#v",rows)
 }
}

func TestLegacyV2VisualsRemainParsable(t *testing.T){
 source:=`<screen name="legacy" version="2"><text align="center">hello</text><panel title="Overview"><badge>Online</badge></panel></screen>`
 view,err:=Parse([]byte(source))
 if err!=nil{t.Fatal(err)}
 lines:=strings.Join(view.RenderWidth(nil,50),"\n")
 if !strings.Contains(lines,"hello")||!strings.Contains(lines,"Online"){t.Fatalf("v2 rendering broke: %q",lines)}
}
