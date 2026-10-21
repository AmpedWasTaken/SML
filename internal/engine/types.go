// Package engine implements parsing, rendering, and validated terminal interactions.
// SML has no embedded code execution.
package engine

const maxFileSize = 64 * 1024
const maxDepth = 16
const maxNodes = 800

// Node is a checked presentation element, not an executable instruction.
type Node struct {
	Kind     string
	Attrs    map[string]string
	Text     string
	Children []Node
}

func (n Node) Attr(key, fallback string) string {
	if v, ok := n.Attrs[key]; ok {
		return v
	}
	return fallback
}

type View struct {
	Name   string
	Nodes  []Node
	assets string
}

func (v View) WithAssets(dir string) View { v.assets = dir; return v }
