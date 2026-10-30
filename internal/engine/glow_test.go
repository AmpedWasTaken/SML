package engine

import (
	"regexp"
	"strings"
	"testing"
)

var stripGlowSGR=regexp.MustCompile("\x1b\\[[0-9;]*m")

func TestPanelAndTableNativeGlowPreserveCellsAndHotspots(t *testing.T){
	source:=`<screen name="glow" version="3">
	  <panel title="Plugin management" border="round" glow="soft" glow-color="#70C7EF" accent="#A38DEF">
	    <button label="Refresh" command="status"/>
	    <table source="plugin.rows" style="rounded" glow="medium" glow-color="#71D6E6" header-style="accent" row-style="striped">
	      <column key="name" label="PLUGIN" width="12"/>
	      <column key="state" label="STATE" width="12"/>
	    </table>
	  </panel>
	</screen>`
	view,err:=Parse([]byte(source));if err!=nil{t.Fatal(err)}
	data:=map[string]string{"plugin.rows":`[{"name":"hashkit","state":"DISABLED"},{"name":"utils","state":"ENABLED"}]`}
	frame,err:=view.RenderFrame(data,64);if err!=nil{t.Fatal(err)}
	if len(frame.Hotspots)!=1||frame.Hotspots[0].Command!="status"{t.Fatal("glow lost button hotspots")}
	plain:=append([]string(nil),frame.Lines...)
	color,err:=PaintGlowFrame(frame,GlowOptions{TrueColor:true});if err!=nil{t.Fatal(err)}
	if len(color)!=len(plain){t.Fatal("glow changed frame height")}
	if strings.Join(color,"\n")==strings.Join(plain,"\n"){t.Fatal("glow painter did not apply colors")}
	var hasColor bool
	for i,line:=range color {
		if strings.Contains(line,"\x1b[38;2;"){hasColor=true}
		if got:=stripGlowSGR.ReplaceAllString(line,"");got!=plain[i]{
			t.Fatalf("glow changed line %d geometry/content: %q != %q",i,got,plain[i])
		}
		if strings.Contains(line,"\x1b[2J")||strings.Contains(line,"\x1b[?25")||strings.Contains(line,"\x1b[H"){
			t.Fatalf("glow emitted cursor or screen control: %q",line)
		}
	}
	if !hasColor{t.Fatal("missing truecolor stroke")}
	legacy,err:=PaintGlowFrame(frame,GlowOptions{TrueColor:true,NoColor:true});if err!=nil{t.Fatal(err)}
	if strings.Join(legacy,"\n")!=strings.Join(plain,"\n"){t.Fatal("legacy mode emitted color")}
	off,err:=PaintGlowFrame(frame,GlowOptions{});if err!=nil{t.Fatal(err)}
	if strings.Join(off,"\n")!=strings.Join(plain,"\n"){t.Fatal("plain renderer changed")}
}

func TestGlowValidationAndNeutralDefaults(t *testing.T){
	for _,source:=range []string{
		`<screen name="x" version="3"><panel glow="laser"><text>x</text></panel></screen>`,
		`<screen name="x" version="3"><panel accent="#f00"><text>x</text></panel></screen>`,
		`<screen name="x" version="3"><table source="rows" glow="super"><column key="id"/></table></screen>`,
		`<screen name="x" version="3"><table source="rows" header-style="rainbow"><column key="id"/></table></screen>`,
		`<screen name="x" version="3"><table source="rows" row-style="blink"><column key="id"/></table></screen>`,
		`<screen name="x" version="3"><table source="rows" glow-color="javascript:bad"><column key="id"/></table></screen>`,
	}{
		if _,err:=Parse([]byte(source));err==nil{t.Errorf("invalid glow option accepted: %s",source)}
	}
	v,err:=Parse([]byte(`<screen name="x" version="3"><panel border="round"><text>ok</text></panel></screen>`))
	if err!=nil{t.Fatal(err)}
	frame,err:=v.RenderFrame(nil,40);if err!=nil{t.Fatal(err)}
	out,err:=PaintGlowFrame(frame,GlowOptions{TrueColor:true});if err!=nil{t.Fatal(err)}
	if strings.Join(out,"\n")!=strings.Join(frame.Lines,"\n"){t.Fatal("unstyled view changed by opt-in painter")}
}

func TestGlowDoesNotFillTablePadding(t *testing.T) {
	view,err:=Parse([]byte(`<screen name="x" version="3">
	 <table source="rows" style="rounded" glow="strong" header-style="solid">
	  <column key="name" label="NAME" width="10"/>
	 </table></screen>`))
	if err!=nil{t.Fatal(err)}
	frame,err:=view.RenderFrame(map[string]string{"rows":`[{"name":"one"}]`},100)
	if err!=nil{t.Fatal(err)}
	lines,err:=PaintGlowFrame(frame,GlowOptions{TrueColor:true});if err!=nil{t.Fatal(err)}
	var cells int
	for i,line:=range lines {
		if strings.Contains(line,"\x1b[48;2;"){cells++}
		if !strings.HasSuffix(stripGlowSGR.ReplaceAllString(line,""),frame.Lines[i]){
			t.Errorf("colored line lost content: %q",line)
		}
		if strings.HasSuffix(line,"\x1b[48;2;"){t.Fatal("incomplete background escape")}
	}
	if cells!=1{t.Fatalf("only table header should receive subtle background, found %d lines",cells)}
}
