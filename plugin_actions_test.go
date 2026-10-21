package sml

import (
 "strings"
 "testing"
)

func TestPluginActionIsTypedAndNotACommand(t *testing.T) {
 view,err:=Parse([]byte(`<screen name="plugin" version="2"><plugin-action label="View health" plugin="server-health" action="report"/></screen>`))
 if err!=nil{t.Fatal(err)}
 lines,spots,_:=view.RenderInteractive(nil,90)
 if len(spots)!=1 {t.Fatalf("expected one hotspot, got %+v (%v)",spots,lines)}
 if spots[0].PluginID!="server-health"||spots[0].PluginAction!="report"||spots[0].Command!=""{t.Fatalf("bad typed action: %+v",spots[0])}
 if !strings.Contains(strings.Join(lines,"\n"),"[ View health ]"){t.Fatal("button missing")}
}
func TestPluginActionRejectsInjectionAndShellAttributes(t *testing.T) {
 for _,src:=range []string{
  `<screen name="test"><plugin-action label="Bad" plugin="server-health" action="report" command="status"/></screen>`,
  `<screen name="test"><plugin-action label="Bad" plugin="../server" action="report"/></screen>`,
  `<screen name="test"><plugin-action label="Bad" plugin="health" action="report;rm"/></screen>`,
  `<screen name="test"><plugin-action label="Bad" plugin="health" action=""/></screen>`,
 } {
  if _,err:=Parse([]byte(src));err==nil{t.Errorf("invalid plugin action accepted: %s",src)}
 }
}
