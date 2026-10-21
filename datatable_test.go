package sml

import (
 "fmt"
 "strings"
 "testing"
)

func TestDataTableRenderSortFilterPage(t *testing.T){
 markup:=[]byte(`<screen name="list"><table source="data.rows" style="rounded" sort="user" order="asc" filter="{{data.filter}}" page="1" page-size="2"><column key="user" label="USER" width="12"/><column key="status" label="STATUS" width="10"/></table></screen>`)
 v,err:=Parse(markup);if err!=nil{t.Fatal(err)}
 values:=map[string]string{"data.rows":`[{"user":"zara","status":"offline"},{"user":"alice","status":"online"},{"user":"adam","status":"online"}]`,"data.filter":"on"}
 out:=strings.Join(v.RenderWidth(values,70),"\n")
 if !strings.Contains(out,"adam")||!strings.Contains(out,"alice")||strings.Contains(out,"zara"){t.Fatalf("incorrect filtered output: %s",out)}
 if strings.Index(out,"adam")>strings.Index(out,"alice"){t.Fatalf("sort order incorrect: %s",out)}
}
func TestDataTableRejectsUnsafeSchema(t *testing.T){
 cases:=[]string{
  `<screen name="x"><table source="rows" style="nope"><column key="user"/></table></screen>`,
  `<screen name="x"><table source="rows"><column key="user"/><column key="user"/></table></screen>`,
  `<screen name="x"><column key="user"/></screen>`,
  `<screen name="x"><table source="rows"><button label="bad" command="help"/></table></screen>`,
 }
 for _,c:=range cases{if _,err:=Parse([]byte(c));err==nil{t.Errorf("accepted invalid schema: %s",c)}}
}
func TestDataTableEscapesControlAndRespectsBounds(t *testing.T){
 v,err:=Parse([]byte(`<screen name="x"><table source="rows" style="ascii"><column key="name" label="NAME" width="8"/></table></screen>`));if err!=nil{t.Fatal(err)}
 out:=strings.Join(v.RenderWidth(map[string]string{"rows":`[{"name":"very-long-header-name"}]`},32),"\n")
 if !strings.Contains(out,"…") {t.Fatal(out)}
 if len(v.RenderWidth(map[string]string{"rows":"not-json"},32))==0{t.Fatal("expected visible error")}
}

func TestLegacyStyleFramesAndBoundedPaging(t *testing.T) {
 expected:=map[string]string{
  "rounded":"╭","default":"┌","unicode":"┏",
  "markdown":"| ","ascii":"NAME",
  "compact":"NAME","compact-lite":"NAME","compact-classic":"NAME",
 }
 for style,prefix:=range expected {
  markup:=fmt.Sprintf(`<screen name="sample"><table source="rows" style="%s" page="99" page-size="1"><column key="name" label="NAME" width="14"/><column key="status" label="STATUS" width="8"/></table></screen>`,style)
  v,err:=Parse([]byte(markup));if err!=nil{t.Fatalf("%s parse: %v",style,err)}
  lines:=v.RenderWidth(map[string]string{"rows":`[{"name":"alpha","status":"up"},{"name":"beta","status":"down"}]`},65)
  joined:=strings.Join(lines,"\n")
  if !strings.HasPrefix(joined,prefix){t.Errorf("%s frame starts incorrectly: %q",style,joined)}
  if !strings.Contains(joined,"beta")||strings.Contains(joined,"alpha"){t.Errorf("%s page bound: %q",style,joined)}
  if !strings.Contains(joined,"Page 2/2"){t.Errorf("%s page number: %q",style,joined)}
  for _,line:=range lines{
   if len([]rune(line))>65{t.Errorf("%s line overflow: %q",style,line)}
  }
 }
}

func TestLegacyTableRejectsInvalidColumnWidth(t *testing.T) {
 for _,width:=range []string{"0","-1","61","garbage"} {
  markup:=fmt.Sprintf(`<screen name="bad"><table source="rows"><column key="name" width="%s"/></table></screen>`,width)
  if _,err:=Parse([]byte(markup));err==nil{t.Fatalf("accepted invalid width: %s",width)}
 }
}
