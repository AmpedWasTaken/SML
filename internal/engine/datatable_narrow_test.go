package engine

import(
 "strings"
 "testing"
)

func TestDataTableNarrowViewportClipsFooterAndEmptyMessage(t *testing.T){
 v,err:=Parse([]byte(`<screen name="small"><table source="rows" empty="No records available in this table" page-size="1"><column key="name" label="NAME" width="6"/></table></screen>`))
 if err!=nil{t.Fatal(err)}
 for _,width:=range []int{12,13,16,20}{
  empty,err:=v.RenderChecked(map[string]string{"rows":"[]"},width)
  if err!=nil{t.Fatal(err)}
  if len(empty)!=1||CellWidth(empty[0])>width{t.Fatalf("empty state overflows width %d: %#v",width,empty)}
  filled,err:=v.RenderChecked(map[string]string{"rows":`[{"name":"alpha"},{"name":"beta"}]`},width)
  if err!=nil{t.Fatal(err)}
  for _,line:=range filled{
   if CellWidth(line)>width{t.Fatalf("table overflows width %d: %q",width,line)}
  }
  if !strings.Contains(strings.Join(filled,"\n"),"Page 1/2") && width>=16{
   t.Fatalf("page indicator unexpectedly missing at width %d: %#v",width,filled)
  }
 }
}
