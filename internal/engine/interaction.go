package engine

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Hotspot uses logical terminal cell offsets in the rendered SML content.
// The TUI translates these into coordinates after borders and scrolling.
type Hotspot struct {
	Line         int
	Start        int // inclusive, 0-based
	End          int // exclusive
	Label        string
	Page         string
	Command      string
	Kind         string
	Key          string
	Options      []string
	PluginID     string
	PluginAction string
	Action       string // Generic host action; never executed by SML.
	ID           string // Explicit control ID or generated layout path.
	Arg          string
	Source       string
	Bindings     string
	Min          int
	Max          int
}

var bindingName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]{0,95}$`)
var sessionArgument = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
var sessionStateSource = regexp.MustCompile(`^state\.[a-z][a-z0-9_]{0,31}$`)
var uiLabel = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _./:+-]{0,46}$`)
var actionID = regexp.MustCompile("^[a-z][a-z0-9_.-]{0,63}$")

// button is a read-only command or page navigation. A future extension API
// may register additional typed safe actions without giving SML a shell.
func validateControl(n Node) error {
	if n.Kind != "button" && n.Kind != "link" && n.Kind != "plugin-action" {
		return nil
	}
	if len(n.Children) != 0 || strings.TrimSpace(n.Text) != "" {
		return errors.New("interactive SML elements must be empty")
	}
	allowed := map[string]bool{"label": true, "page": true, "command": true, "action": true, "id": true, "enabled-when": true, "variant": true, "tone": true, "align": true, "case": true, "padding": true, "border": true}
	if n.Kind == "plugin-action" {
		allowed = map[string]bool{"label": true, "id": true, "plugin": true, "action": true, "arg": true, "source": true, "bindings": true, "enabled-when": true, "style": true, "variant": true, "tone": true, "align": true, "case": true, "padding": true, "border": true}
		if !validName.MatchString(n.Attr("plugin", "")) || !validName.MatchString(n.Attr("action", "")) {
			return errors.New("invalid SML plugin action identity")
		}
		arg, source := n.Attr("arg", ""), n.Attr("source", "")
		if (arg == "") != (source == "") {
			return errors.New("plugin action must declare arg and source together")
		}
		if arg != "" && (!sessionArgument.MatchString(arg) || !sessionStateSource.MatchString(source)) {
			return errors.New("invalid SML plugin action binding")
		}
		spec := n.Attr("bindings", "")
		if spec != "" {
			if arg != "" || source != "" {
				return errors.New("use either bindings or arg/source")
			}
			if _, err := ParseActionBindings(spec); err != nil {
				return err
			}
		}
	}
	if binding, declared := n.Attrs["enabled-when"]; declared && !bindingName.MatchString(binding) {
		return errors.New("invalid interactive enabled-when binding")
	}
	for key := range n.Attrs {
		if !allowed[key] {
			return fmt.Errorf("unsupported interactive attribute %q", key)
		}
	}
	if !uiLabel.MatchString(n.Attr("label", "")) {
		return errors.New("invalid control label (ASCII only, max 47 chars)")
	}
	page := n.Attr("page", "")
	command := n.Attr("command", "")
	action := n.Attr("action", "")
	if id := n.Attr("id", ""); id != "" && !actionID.MatchString(id) {
		return errors.New("invalid interactive element id")
	}
	if n.Kind != "plugin-action" {
		count := 0
		for _, target := range []string{page, command, action} {
			if target != "" {
				count++
			}
		}
		if count != 1 {
			return errors.New("button/link requires exactly one of page, command or action")
		}
		if action != "" && !actionID.MatchString(action) {
			return errors.New("invalid generic action identifier")
		}
	}
	if page != "" && !validName.MatchString(page) {
		return errors.New("invalid page destination")
	}
	if command != "" && !legacyReadOnlyCommands[command] {
		return fmt.Errorf("command %q is not in the read-only SML action allowlist", command)
	}
	return nil
}

// controlEnabled gates the visual interactivity of a control using current,
// already-authorized SML bindings. It never replaces server-side role and
// plugin command authorization.
func controlEnabled(n Node, values map[string]string) bool {
	key := n.Attr("enabled-when", "")
	if key == "" {
		return true
	}
	value := strings.ToLower(strings.TrimSpace(values[key]))
	return value != "" && value != "0" && value != "false" && value != "off" && value != "no"
}

func renderActionControl(n Node, values map[string]string, width int) string {
	label := n.Attr("label", "")
	if !controlEnabled(n, values) {
		label += " (disabled)"
	}
	return styleAction(n, label, width)
}

func validateUI(nodes []Node, screen string) error {
	seen := map[string]bool{}
	for _, node := range nodes {
		if node.Kind != "derive" {
			continue
		}
		key := node.Attr("key", "")
		if seen[key] {
			return errors.New("duplicate SML derive key")
		}
		if len(seen) >= maxDerivations {
			return errors.New("too many SML derivations")
		}
		seen[key] = true
	}
	return validateUIChildren(nodes, screen, "")
}
func validateUIChildren(nodes []Node, screen, parent string) error {
	for _, node := range nodes {
		if node.Kind == "gradient" && screen != "appearance" {
			return errors.New("gradient metadata belongs only in appearance.sml")
		}
		if node.Kind == "derive" {
			if parent != "" {
				return errors.New("derive must be a top-level screen declaration")
			}
			if err := validateDerive(node); err != nil {
				return err
			}
		}
		if node.Kind == "otherwise" {
			if parent != "if" && parent != "when" {
				return errors.New("otherwise requires if/when parent")
			}
			if len(node.Attrs) > 0 || strings.TrimSpace(node.Text) != "" {
				return errors.New("otherwise cannot have attributes or text")
			}
		}
		if node.Kind == "if" || node.Kind == "when" {
			for i, child := range node.Children {
				if child.Kind == "otherwise" && i != len(node.Children)-1 {
					return errors.New("otherwise must be the last conditional child")
				}
			}
		}
		if err := validateControl(node); err != nil {
			return err
		}
		if err := validateRepeatSyntax(node); err != nil {
			return err
		}
		if err := validateSessionControl(node); err != nil {
			return err
		}
		if err := validateSessionInput(node); err != nil {
			return err
		}
		if node.Kind == "section" {
			if len(node.Attr("title", "")) > 120 {
				return errors.New("SML section title too long")
			}
			for key := range node.Attrs {
				if key != "title" && !styleAttributes[key] {
					return fmt.Errorf("invalid section attribute %q", key)
				}
			}
		}
		if node.Kind == "field" {
			if len(node.Children) > 0 || strings.TrimSpace(node.Text) != "" {
				return errors.New("SML field must be empty")
			}
			if !bindingName.MatchString(node.Attr("key", "")) {
				return errors.New("SML field requires a safe binding key")
			}
			if len(node.Attr("label", "")) > 80 || len(node.Attr("unit", "")) > 32 || len(node.Attr("fallback", "")) > 256 {
				return errors.New("SML field attribute too long")
			}
			for key := range node.Attrs {
				if key != "key" && key != "label" && key != "unit" && key != "fallback" && key != "style" && !styleAttributes[key] {
					return fmt.Errorf("invalid field attribute %q", key)
				}
			}
		}
		if node.Kind == "empty" {
			if !bindingName.MatchString(node.Attr("field", "")) {
				return errors.New("SML empty requires safe binding field")
			}
			for key := range node.Attrs {
				if key != "field" {
					return fmt.Errorf("invalid empty attribute %q", key)
				}
			}
		}
		if node.Kind == "table" {
			if err := validateTable(node); err != nil {
				return err
			}
		}
		// Layout <column> stacks children; table <column key="..."> is schema-only.
		if node.Kind == "column" && parent != "table" {
			if _, ok := node.Attrs["key"]; ok {
				return errors.New("table column elements require a table parent")
			}
		}
		if err := validateWidget(node); err != nil {
			return err
		}
		if node.Kind == "image" || node.Kind == "gif" {
			if err := validateMediaNode(node); err != nil {
				return err
			}
		}
		if err := validateUIChildren(node.Children, screen, node.Kind); err != nil {
			return err
		}
	}
	return nil
}

// RenderInteractive preserves the legacy signature while delegating all layout
// and hit testing to the structural cell renderer. No label matching is used.
func (v View) RenderInteractive(values map[string]string, width int) ([]string, []Hotspot, bool) {
	frame, err := v.RenderFrame(values, width)
	if err != nil {
		return []string{"SML render error: " + safe(err.Error())}, nil, false
	}
	return frame.Lines, frame.Hotspots, v.HasAnimation()
}

// RenderGlowInteractive keeps SML input/hotspot geometry identical to
// RenderInteractive, but opt-in hosts may receive static, truecolor ANSI SGR
// for native panel and table accents. No control actions or cursor movement.
func (v View) RenderGlowInteractive(values map[string]string,width int,opts GlowOptions)([]string,[]Hotspot,bool){
	frame,err:=v.RenderFrame(values,width)
	if err!=nil{return []string{"SML render error: "+safe(err.Error())},nil,false}
	painted,err:=PaintGlowFrame(frame,opts)
	if err!=nil{return []string{"SML glow error: "+safe(err.Error())},nil,false}
	return painted,frame.Hotspots,v.HasAnimation()
}

func (v View) HasAnimation() bool {
	var visit func([]Node, int) bool
	visit = func(nodes []Node, depth int) bool {
		if depth > maxDepth {
			return false
		}
		for _, node := range nodes {
			if node.Kind == "gif" {
				return true
			}
			if node.Kind == "include" || node.Kind == "use" {
				if part, err := v.component(node.Attr("name", "")); err == nil && visit(part.Children, depth+1) {
					return true
				}
			}
			if visit(node.Children, depth+1) {
				return true
			}
		}
		return false
	}
	return visit(v.Nodes, 0)
}
