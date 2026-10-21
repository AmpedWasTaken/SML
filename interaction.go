package sml

import (
 "errors"
 "fmt"
 "regexp"
 "strings"
 "unicode/utf8"
)

// Hotspot uses logical terminal cell offsets in the rendered SML content.
// The TUI translates these into coordinates after borders and scrolling.
type Hotspot struct {
 Line int
 Start int // inclusive, 0-based
 End int // exclusive
 Label string
 Page string
 Command string
 Kind string
 Key string
 Options []string
 PluginID string
 PluginAction string
 Arg string
 Source string
 Bindings string
 Min int
 Max int
}

var bindingName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]{0,95}$`)
var sessionArgument = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
var sessionStateSource = regexp.MustCompile(`^state\.[a-z][a-z0-9_]{0,31}$`)
var uiLabel = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _./:+-]{0,46}$`)
var safeCommands = map[string]bool{
 "help":true, "status":true,"whoami":true,"jobs":true,
 "lab":true,"lab methods":true,"lab ongoing":true,"lab history":true,
 "api list":true,"plugins list":true,"sessions":true,"version":true,
}

// button is a read-only command or page navigation. A future extension API
// may register additional typed safe actions without giving SML a shell.
func validateControl(n Node) error {
 if n.Kind!="button" && n.Kind!="link" && n.Kind!="plugin-action" {return nil}
 if len(n.Children)!=0 || strings.TrimSpace(n.Text)!="" {
  return errors.New("interactive SML elements must be empty")
 }
 allowed:=map[string]bool{"label":true,"page":true,"command":true,"enabled-when":true,"variant":true,"tone":true,"align":true,"case":true,"padding":true,"border":true}
 if n.Kind=="plugin-action"{
  allowed=map[string]bool{"label":true,"plugin":true,"action":true,"arg":true,"source":true,"bindings":true,"enabled-when":true,"style":true,"variant":true,"tone":true,"align":true,"case":true,"padding":true,"border":true}
  if !validName.MatchString(n.Attr("plugin","")) || !validName.MatchString(n.Attr("action","")){return errors.New("invalid SML plugin action identity")}
  arg,source:=n.Attr("arg",""),n.Attr("source","")
  if (arg=="")!=(source==""){return errors.New("plugin action must declare arg and source together")}
  if arg!="" && (!sessionArgument.MatchString(arg)||!sessionStateSource.MatchString(source)){return errors.New("invalid SML plugin action binding")}
  spec:=n.Attr("bindings","")
  if spec!="" {
   if arg!=""||source!="" {return errors.New("use either bindings or arg/source")}
   if _,err:=ParseActionBindings(spec);err!=nil{return err}
  }
 }
 if binding,declared:=n.Attrs["enabled-when"];declared && !bindingName.MatchString(binding){
  return errors.New("invalid interactive enabled-when binding")
 }
 for key:=range n.Attrs {if !allowed[key]{return fmt.Errorf("unsupported interactive attribute %q",key)}}
 if !uiLabel.MatchString(n.Attr("label","")) {return errors.New("invalid control label (ASCII only, max 47 chars)")}
 page:=n.Attr("page","")
 command:=n.Attr("command","")
 if n.Kind!="plugin-action" && (page=="")== (command=="") {return errors.New("button/link needs exactly one page or command action")}
 if page!="" && !validName.MatchString(page) {return errors.New("invalid page destination")}
 if command!="" && !safeCommands[command] {return fmt.Errorf("command %q is not in the read-only SML action allowlist",command)}
 return nil
}

// controlEnabled gates the visual interactivity of a control using current,
// already-authorized SML bindings. It never replaces server-side role and
// plugin command authorization.
func controlEnabled(n Node, values map[string]string) bool {
 key:=n.Attr("enabled-when","")
 if key=="" {return true}
 value:=strings.ToLower(strings.TrimSpace(values[key]))
 return value!="" && value!="0" && value!="false" && value!="off" && value!="no"
}

func renderActionControl(n Node, values map[string]string, width int) string {
 label:=n.Attr("label","")
 if !controlEnabled(n,values) {label+=" (disabled)"}
 return styleAction(n,label,width)
}

func validateUI(nodes []Node, screen string) error {
 seen:=map[string]bool{}
 for _,node:=range nodes {
  if node.Kind!="derive" {continue}
  key:=node.Attr("key","")
  if seen[key] {return errors.New("duplicate SML derive key")}
  if len(seen)>=maxDerivations {return errors.New("too many SML derivations")}
  seen[key]=true
 }
 return validateUIChildren(nodes,screen,"")
}
func validateUIChildren(nodes []Node, screen, parent string) error {
 for _,node:=range nodes{
  if node.Kind=="gradient" && screen!="appearance" {return errors.New("gradient metadata belongs only in appearance.sml")}
   if node.Kind=="derive" {
   if parent!="" {return errors.New("derive must be a top-level screen declaration")}
   if err:=validateDerive(node);err!=nil{return err}
  }
  if node.Kind=="otherwise" {
   if parent!="if" && parent!="when" {return errors.New("otherwise requires if/when parent")}
   if len(node.Attrs)>0 || strings.TrimSpace(node.Text)!="" {return errors.New("otherwise cannot have attributes or text")}
  }
  if node.Kind=="if" || node.Kind=="when" {
   for i,child:=range node.Children {
    if child.Kind=="otherwise" && i!=len(node.Children)-1 {return errors.New("otherwise must be the last conditional child")}
   }
  }
  if err:=validateControl(node);err!=nil{return err}
  if err:=validateSessionControl(node);err!=nil{return err}
  if err:=validateSessionInput(node);err!=nil{return err}
  if node.Kind=="section" {
   if len(node.Attr("title",""))>120 {return errors.New("SML section title too long")}
   for key:=range node.Attrs {if key!="title"&&!styleAttributes[key]{return fmt.Errorf("invalid section attribute %q",key)}}
  }
  if node.Kind=="field" {
   if len(node.Children)>0||strings.TrimSpace(node.Text)!="" {return errors.New("SML field must be empty")}
   if !bindingName.MatchString(node.Attr("key","")){return errors.New("SML field requires a safe binding key")}
   if len(node.Attr("label",""))>80||len(node.Attr("unit",""))>32||len(node.Attr("fallback",""))>256{return errors.New("SML field attribute too long")}
   for key:=range node.Attrs {if key!="key"&&key!="label"&&key!="unit"&&key!="fallback"&&key!="style"&&!styleAttributes[key]{return fmt.Errorf("invalid field attribute %q",key)}}
  }
  if node.Kind=="empty" {
   if !bindingName.MatchString(node.Attr("field","")){return errors.New("SML empty requires safe binding field")}
   for key:=range node.Attrs {if key!="field"{return fmt.Errorf("invalid empty attribute %q",key)}}
  }
  if node.Kind=="table" {if err:=validateTable(node);err!=nil{return err}}
  // Layout <column> stacks children; table <column key="..."> is schema-only.
  if node.Kind=="column"&&parent!="table"{
   if _,ok:=node.Attrs["key"];ok{return errors.New("table column elements require a table parent")}
  }
  if err:=validateWidget(node);err!=nil{return err}
  if node.Kind=="image"||node.Kind=="gif" {
   if err:=validateMediaNode(node);err!=nil{return err}
  }
  if err:=validateUIChildren(node.Children,screen,node.Kind);err!=nil{return err}
 }
 return nil
}

// collectControls walks only the branch that the renderer can display.
// A hidden control must never inherit a hotspot from visible text with the
// same label: a copied label is not permission to invoke an invisible action.
func (v View) collectControls(nodes []Node, values map[string]string, depth int, dst *[]Node) error {
	if depth > maxDepth { return errors.New("SML control include depth exceeded") }
	for _, n := range nodes {
		switch n.Kind {
		case "button", "link", "plugin-action":
			if controlEnabled(n, values) { *dst = append(*dst, n) }
		case "toggle", "choice", "input", "number":
			*dst = append(*dst, n)
		case "derive", "otherwise":
			continue
		case "if", "when":
			field := n.Attr("field", n.Attr("value", ""))
			value := values[field]
			matched := false
			if n.Kind == "if" {
				matched = value == n.Attr("equals", "true")
			} else {
				matched = value != "" && value != "0" && value != "false" && value != "no"
			}
			primary := n.Children
			var alternative []Node
			for index, child := range n.Children {
				if child.Kind == "otherwise" {
					primary, alternative = n.Children[:index], child.Children
					break
				}
			}
			visible := primary
			if !matched { visible = alternative }
			if err := v.collectControls(visible, values, depth+1, dst); err != nil { return err }
		case "empty":
			if values[n.Attr("field", "")] == "" {
				if err := v.collectControls(n.Children, values, depth+1, dst); err != nil {return err}
			}
		case "include":
			part, err := v.component(n.Attr("name", ""))
			if err != nil { return err }
			if err := v.collectControls(part.Children, values, depth+1, dst); err != nil { return err }
		case "repeat":
			key := n.Attr("source", "")
			if values[key] == "" { continue }
			alias := n.Attr("as", "item")
			items := strings.Split(values[key], ",")
			limit := min(len(items), boundedInt(n.Attr("limit", "30"), 30, 0, 100))
			for i := 0; i < limit; i++ {
				local := make(map[string]string, len(values)+1)
				for k, val := range values { local[k] = val }
				local[alias] = strings.TrimSpace(items[i])
				if err := v.collectControls(n.Children, local, depth+1, dst); err != nil {return err}
			}
		default:
			if err := v.collectControls(n.Children, values, depth+1, dst); err != nil {return err}
		}
	}
	return nil
}

// RenderInteractive is a side-effect-free render plus a geometry index. Labels
// are matched against the final rendered lines (including nested grid/panel).
// Only controls still visible after truncation get a hotspot.
func (v View) RenderInteractive(values map[string]string, width int) ([]string,[]Hotspot,bool) {
 values = v.ResolveBindings(values)
 lines:=v.RenderWidth(values,width)
 controls:=[]Node{}
 if err:=v.collectControls(v.Nodes,values,0,&controls);err!=nil{return []string{"SML render error: "+safe(err.Error())},nil,false}
 hotspots:=[]Hotspot{}
 occupied:=map[int]map[int]bool{}
 for _,raw:=range controls{
  control:=raw
  text:=renderActionControl(control,values,width)
  if control.Kind=="toggle"||control.Kind=="choice" {text=renderSessionControl(control,values,width)}
  if control.Kind=="input"||control.Kind=="number" {text=renderSessionInput(control,values,width)}
  for row,line:=range lines{
   searchFrom:=0
   for searchFrom<len(line) {
    offset:=strings.Index(line[searchFrom:],text)
    if offset<0{break}
    offset+=searchFrom
    begin:=utf8.RuneCountInString(line[:offset])
    end:=begin+utf8.RuneCountInString(text)
    taken:=false
    for col:=begin;col<end;col++{if occupied[row][col]{taken=true;break}}
    if !taken {
     if occupied[row]==nil{occupied[row]=map[int]bool{}}
     for col:=begin;col<end;col++{occupied[row][col]=true}
     hotspots=append(hotspots,Hotspot{Line:row,Start:begin,End:end,Label:control.Attr("label",""),Page:control.Attr("page",""),Command:control.Attr("command",""),Kind:control.Kind,Key:control.Attr("key",""),Options:controlOptions(control),PluginID:control.Attr("plugin",""),PluginAction:control.Attr("action",""),Arg:control.Attr("arg",""),Source:control.Attr("source",""),Bindings:control.Attr("bindings",""),Min:boundedInt(control.Attr("min","0"),0,0,1000000),Max:boundedInt(control.Attr("max","100"),100,0,1000000)})
    }
    searchFrom=offset+len(text)
   }
  }
 }
 return lines,hotspots,v.HasAnimation()
}

func (v View) HasAnimation() bool {
 var visit func([]Node,int) bool
 visit=func(nodes []Node,depth int)bool{
  if depth>maxDepth{return false}
  for _,node:=range nodes{
   if node.Kind=="gif"{return true}
   if node.Kind=="include"{
    if part,err:=v.component(node.Attr("name",""));err==nil&&visit(part.Children,depth+1){return true}
   }
   if visit(node.Children,depth+1){return true}
  }
  return false
 }
 return visit(v.Nodes,0)
}
