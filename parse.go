package sml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

var validName = regexp.MustCompile("^[A-Za-z0-9_-]{1,48}$")
var allowed = map[string]bool{
	"screen": true, "component": true, "column": true, "row": true, "grid": true,
	"text": true, "banner": true, "stat": true, "metric": true, "divider": true,
	"spacer": true, "panel": true, "badge": true, "progress": true, "include": true,
	"if": true, "when": true, "repeat": true, "table": true, "prompt": true, "alert": true,
	"gradient": true, "button": true, "link": true, "image": true, "gif": true,
	"gauge": true, "sparkline": true, "plugin-action": true,
	"toggle": true, "choice": true, "input": true, "number": true,
	"section": true, "field": true, "empty": true, "derive": true, "otherwise": true,
	"use": true, "slot": true, "fill": true,
}

// Parse rejects invalid XML and unknown components before publishing a view.
func Parse(data []byte) (View, error) {
	if len(data) > maxFileSize || !utf8.Valid(data) {
		return View{}, errors.New("invalid SML size or UTF-8")
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	var root *Node
	stack := []*Node{}
	count := 0
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return View{}, fmt.Errorf("invalid SML XML: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			count++
			if count > maxNodes || len(stack) >= maxDepth {
				return View{}, errors.New("SML node/depth limit exceeded")
			}
			if !allowed[t.Name.Local] || t.Name.Space != "" {
				return View{}, fmt.Errorf("unsupported SML component <%s>", t.Name.Local)
			}
			n := &Node{Kind: t.Name.Local, Attrs: map[string]string{}}
			for _, a := range t.Attr {
				if a.Name.Space != "" {
					return View{}, errors.New("SML namespaces are not supported")
				}
				if _, exists := n.Attrs[a.Name.Local]; exists {
					return View{}, errors.New("duplicate SML attribute")
				}
				n.Attrs[a.Name.Local] = a.Value
			}
			if len(stack) == 0 {
				if root != nil {
					return View{}, errors.New("SML requires exactly one root")
				}
				root = n
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, *n)
				n = &parent.Children[len(parent.Children)-1]
			}
			stack = append(stack, n)
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(t)) != "" {
					return View{}, errors.New("unexpected text outside SML root")
				}
			} else {
				stack[len(stack)-1].Text += string(t)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return View{}, errors.New("unbalanced SML XML")
			}
			stack = stack[:len(stack)-1]
		case xml.ProcInst, xml.Directive:
			return View{}, errors.New("SML directives and processing instructions are not supported")
		}
	}
	if root == nil || root.Kind != "screen" {
		return View{}, errors.New("SML root must be <screen name=\"...\">")
	}
	version := root.Attr("version", "")
	if version != "" && version != "1" && version != "2" && version != "3" {
		return View{}, errors.New("unsupported SML version")
	}
	name := root.Attr("name", "")
	if !validName.MatchString(name) {
		return View{}, errors.New("invalid SML view name")
	}
	if len(stack) != 0 {
		return View{}, errors.New("incomplete SML XML")
	}
	view := View{Name: name, Nodes: root.Children}
	if version != "3" && requiresSML3(view.Nodes) {
		return View{}, errors.New("reactive SML components require version 3")
	}
	if err := validateComponentSyntax(view.Nodes); err != nil {
		return View{}, err
	}
	if err := validateUI(view.Nodes, name); err != nil {
		return View{}, err
	}
	if err := validateStyleSystem(view.Nodes); err != nil {
		return View{}, err
	}
	if name == "appearance" {
		if _, err := view.GradientStyles(); err != nil {
			return View{}, err
		}
	} else {
		for _, node := range view.Nodes {
			if node.Kind == "gradient" {
				return View{}, errors.New("gradient metadata belongs in appearance.sml")
			}
		}
	}
	return view, nil
}
