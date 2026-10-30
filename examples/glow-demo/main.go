// Command glow-demo previews static truecolor glow on native SML panels/tables.
// Run from repository root: go run ./examples/glow-demo
package main

import (
	"fmt"
	"os"

	sml "github.com/AmpedWasTaken/SML"
)

func main(){
	source,err:=os.ReadFile("examples/glow-panels.sml")
	if err!=nil{panic(err)}
	view,err:=sml.Parse(source)
	if err!=nil{panic(err)}
	values:=map[string]string{"plugins.rows":`[{"name":"hashkit","state":"DISABLED","runtime":"code: exec"},{"name":"monitor","state":"ENABLED","runtime":"SML native"}]`}
	frame,err:=view.RenderFrame(values,78)
	if err!=nil{panic(err)}
	rows,err:=sml.PaintGlowFrame(frame,sml.GlowOptions{TrueColor:true})
	if err!=nil{panic(err)}
	for _,row:=range rows {fmt.Println(row)}
	fmt.Print("\x1b[0m")
}
