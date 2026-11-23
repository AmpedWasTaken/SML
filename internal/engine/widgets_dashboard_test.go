package engine

import (
  "fmt"
  "strings"
  "testing"
)

func TestDashboardWidgetsBothRenderers(t *testing.T) {
  widgets := []string{
    `<heatmap label="ACTIVITY" values="0,10,25,50,75,100" columns="3"/>`,
    `<bar-chart label="THROUGHPUT" values="2,4,7" labels="API,SSH,GUI" width="12"/>`,
    `<stacked-bar label="STORAGE" values="30,20,50" labels="Used,Cache,Free" width="12"/>`,
    `<waveform label="SIGNAL" values="0,2,5,3,1,4" width="12" height="3"/>`,
    `<kpi label="REQUESTS" value="{{total}}" delta="+12%" unit="/s"/>`,
    `<activity-feed label="EVENTS" events="12:01|ok|Login;12:02|warn|Load spike" max-items="2"/>`,
  }
  for _, widget := range widgets {
    t.Run(strings.Fields(widget)[0], func(t *testing.T) {
      source := `<screen name="test" version="3">` + widget + `</screen>`
      view, err := Parse([]byte(source))
      if err != nil { t.Fatal(err) }
      values := map[string]string{"total":"42"}
      for _, width := range []int{20, 40, 80, 120} {
        frame, err := view.RenderFrame(values, width)
        if err != nil { t.Fatalf("width %d: %v",width,err) }
        if len(frame.Lines) == 0 { t.Fatal("no widget output") }
        for i,line := range frame.Lines {
          if CellWidth(line) > width { t.Fatalf("width=%d row=%d cells=%d %q",width,i,CellWidth(line),line) }
        }
        legacy:=view.RenderWidth(values,width)
        if len(legacy)==0 { t.Fatal("legacy render empty") }
      }
    })
  }
}

func TestDashboardWidgetsRejectUnsupportedAttributes(t *testing.T) {
  widgets := []string{"heatmap","bar-chart","stacked-bar","waveform","kpi","activity-feed"}
  for _, widget:=range widgets {
    for _, attr := range []string{`onclick="run"`, `width="-2"`, `label="ok"><button label="RUN" action="app.run"/></`+widget} {
      if strings.HasPrefix(attr,"width") && (widget=="heatmap"||widget=="kpi"||widget=="activity-feed") {
        continue
      }
      src := fmt.Sprintf(`<screen name="t" version="3"><%s %s/></screen>`,widget,attr)
      if _,err:=Parse([]byte(src));err==nil {t.Fatalf("%s accepted %s",widget,attr)}
    }
  }
}

func TestDashboardWidgetsSafeUntrustedValues(t *testing.T) {
  view,err:=Parse([]byte(`<screen name="safe" version="3">
    <heatmap values="{{host.history}}"/>
    <bar-chart values="2,1" labels="{{host.labels}}"/>
    <activity-feed events="{{host.events}}"/>
    <waveform values="NaN,Inf,-Inf,4,9,0" width="12"/>
  </screen>`))
  if err!=nil {t.Fatal(err)}
  frame,err:=view.RenderFrame(map[string]string{
    "host.history":strings.Repeat("0,",300)+"99999",
    "host.labels":"\x1b[31mSECRET,READY",
    "host.events":strings.Repeat("00:01|ok|msg;",100),
  },40)
  if err!=nil {t.Fatal(err)}
  if len(frame.Lines)>120 {t.Fatalf("unbounded widget output: %d",len(frame.Lines))}
  for _,line:=range frame.Lines {
    if CellWidth(line)>40 || strings.ContainsRune(line,'\x1b'){t.Fatalf("unsafe line %q",line)}
  }
}

func TestDashboardWidgetMalformedBoundedData(t *testing.T) {
  for _,widget:=range []string{"heatmap","bar-chart","stacked-bar","waveform","kpi","activity-feed"} {
    src:=fmt.Sprintf(`<screen name="t" version="3"><%s label="Empty"/></screen>`,widget)
    view,err:=Parse([]byte(src)); if err!=nil{t.Fatal(err)}
    if _,err:=view.RenderFrame(nil,12);err!=nil{t.Fatal(err)}
  }
}
