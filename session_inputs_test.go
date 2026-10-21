package sml

import (
 "strings"
 "testing"
)

func TestLocalInputSchemaAndGeometry(t *testing.T){
 v,err:=Parse([]byte(`<screen name="settings" version="2"><input label="Alias" key="state.alias" placeholder="Enter name"/><number label="Rows" key="state.rows" min="8" max="100" placeholder="24"/></screen>`))
 if err!=nil {t.Fatal(err)}
 lines,hits,_:=v.RenderInteractive(map[string]string{},90)
 if len(hits)!=2||hits[0].Kind!="input"||hits[1].Kind!="number"{t.Fatalf("wrong input hotspots: %+v",hits)}
 if hits[1].Min!=8||hits[1].Max!=100{t.Fatal("numeric range not propagated")}
 if !strings.Contains(strings.Join(lines,"\n"),"Enter name"){t.Fatal("placeholder missing")}
}
func TestLocalInputRejectsUnsafeSchema(t *testing.T){
 for _,src:=range []string{
  `<screen name="settings"><input label="bad" key="plugin.secret"/></screen>`,
  `<screen name="settings"><number label="Rows" key="state.rows" min="100" max="8"/></screen>`,
  `<screen name="settings"><input label="X" key="state.x" command="status"/></screen>`,
 }{
  if _,err:=Parse([]byte(src));err==nil{t.Errorf("accepted %s",src)}
 }
}
