// Package sml provides a tiny, dependency-free prototype for declarative terminal UIs.
//
// This independent PoC intentionally does not claim feature parity with the
// private Sentinel V5 engine. It does not run commands or evaluate expressions.
package sml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Node represents a validated declarative SML element.
type Node struct {
	Tag      string
	Attrs    map[string]string
	Text     string
	Children []*Node
}

const MaxDocumentBytes = 1 << 20
const MaxNestingDepth = 32

var tags = map[string]bool{"sml": true, "panel": true, "grid": true, "text": true, "metric": true}

// Parse turns a UTF-8 XML-compatible SML document into a UI tree.
//
// Supported elements: sml, panel, grid, text, metric.
// Unknown elements, directives (including DOCTYPE), namespaces, and deep trees
// are rejected. There is no external entity fetching or code evaluation.
func Parse(data []byte) (*Node, error) {
	if len(data) > MaxDocumentBytes {
		return nil, fmt.Errorf("SML document exceeds %d bytes", MaxDocumentBytes)
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true
	var root *Node
	var stack []*Node
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid SML: %w", err)
		}
		switch tok := token.(type) {
		case xml.StartElement:
			if tok.Name.Space != "" || !tags[tok.Name.Local] {
				return nil, fmt.Errorf("unsupported SML element: %q", tok.Name.Local)
			}
			if len(stack) >= MaxNestingDepth {
				return nil, fmt.Errorf("SML tree exceeds nesting limit")
			}
			if root != nil && len(stack) == 0 {
				return nil, fmt.Errorf("multiple root elements")
			}
			node := &Node{Tag: tok.Name.Local, Attrs: make(map[string]string)}
			for _, attr := range tok.Attr {
				if attr.Name.Space != "" {
					return nil, fmt.Errorf("namespaced attributes are unsupported")
				}
				if _, exists := node.Attrs[attr.Name.Local]; exists {
					return nil, fmt.Errorf("duplicate attribute %q", attr.Name.Local)
				}
				node.Attrs[attr.Name.Local] = attr.Value
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(tok)) != "" {
					return nil, fmt.Errorf("text outside the SML root")
				}
			} else {
				stack[len(stack)-1].Text += string(tok)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("unexpected closing tag")
			}
			stack = stack[:len(stack)-1]
		case xml.Comment:
			// Comments are documentation only.
		default:
			return nil, fmt.Errorf("XML directives/instructions are not supported")
		}
	}
	if root == nil || root.Tag != "sml" {
		return nil, fmt.Errorf("root element must be <sml>")
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("incomplete SML tree")
	}
	return root, nil
}
