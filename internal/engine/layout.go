package engine

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
)

// Cell represents one terminal display column. A multi-column grapheme
// occupies one leading cell and one or more continuation cells.
type Cell struct {
	Grapheme     string
	Continuation bool
	ElementID    string
}

// Rect is expressed in zero-based terminal display cells.
type Rect struct{ X, Y, Width, Height int }

// Element records the layout bounds of a named view node.
type Element struct {
	ID     string
	Kind   string
	Bounds Rect
	// Presentation-only SML metadata. The base frame stays plain text;
	// clients opt into ANSI colors through PaintGlowFrame.
	Glow        string
	GlowColor   string
	Accent      string
	HeaderStyle string
	RowStyle    string
}

// Frame is a validated, positioned terminal snapshot. No I/O is performed.
type Frame struct {
	Width    int
	Height   int
	Lines    []string
	Cells    [][]Cell
	Elements []Element
	Hotspots []Hotspot
}

type layoutBlock struct {
	lines    []string
	elements []Element
	spots    []Hotspot
}
type locatedBlock struct {
	block layoutBlock
	x, y  int
}

func cellLine(text string, width int) []Cell {
	result := make([]Cell, max(0, width))
	col := 0
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		glyph := g.Str()
		if glyph == "\n" || glyph == "\r" {
			continue
		}
		w := g.Width()
		if glyph == "\t" {
			glyph = "    "
			w = 4
		}
		if w < 1 {
			continue
		}
		if col+w > width {
			break
		}
		result[col] = Cell{Grapheme: glyph}
		for step := 1; step < w; step++ {
			result[col+step].Continuation = true
		}
		col += w
	}
	return result
}

func cellLineText(cells []Cell) string {
	var result strings.Builder
	for _, cell := range cells {
		if cell.Continuation {
			continue
		}
		if cell.Grapheme == "" {
			result.WriteByte(' ')
		} else {
			result.WriteString(cell.Grapheme)
		}
	}
	return strings.TrimRight(result.String(), " ")
}
func renderCanvas(width, height int, children []locatedBlock) []string {
	if height == 0 {
		return nil
	}
	canvas := make([][]Cell, height)
	for i := range canvas {
		canvas[i] = make([]Cell, width)
	}
	for _, part := range children {
		for y, line := range part.block.lines {
			yy := part.y + y
			if yy < 0 || yy >= height || part.x >= width {
				continue
			}
			cells := cellLine(line, max(0, width-part.x))
			if part.x < 0 {
				continue
			}
			copy(canvas[yy][part.x:], cells)
		}
	}
	lines := make([]string, height)
	for y := range canvas {
		lines[y] = cellLineText(canvas[y])
	}
	return lines
}

func offsetBlock(out *layoutBlock, in layoutBlock, x, y int) {
	for _, spot := range in.spots {
		spot.Line += y
		spot.Start += x
		spot.End += x
		out.spots = append(out.spots, spot)
	}
	for _, el := range in.elements {
		el.Bounds.X += x
		el.Bounds.Y += y
		out.elements = append(out.elements, el)
	}
}
func (out *layoutBlock) mark(n Node, id string, width int) {
	actual := 0
	for _, line := range out.lines {
		if w := CellWidth(line); w > actual {
			actual = w
		}
	}
	if n.Kind == "row" || n.Kind == "grid" || n.Kind == "panel" {
		actual = width
	}
	out.elements = append(out.elements, Element{
		ID: id, Kind: n.Kind, Bounds: Rect{Width: min(actual, width), Height: len(out.lines)},
		Glow: n.Attr("glow", "none"), GlowColor: n.Attr("glow-color", ""),
		Accent: n.Attr("accent", ""), HeaderStyle: n.Attr("header-style", "default"), RowStyle: n.Attr("row-style", "plain"),
	})
}
func controlNode(n Node) bool {
	switch n.Kind {
	case "button", "link", "plugin-action", "toggle", "choice", "input", "number":
		return true
	}
	return false
}
func (v View) renderLayoutNodes(nodes []Node, values map[string]string, width, depth int, path string) (layoutBlock, error) {
	result := layoutBlock{}
	for index, n := range nodes {
		childPath := fmt.Sprintf("%s/%d:%s", path, index, n.Kind)
		child, err := v.renderLayoutNode(n, values, width, depth+1, childPath)
		if err != nil {
			return layoutBlock{}, err
		}
		y := len(result.lines)
		result.lines = append(result.lines, child.lines...)
		offsetBlock(&result, child, 0, y)
		if len(result.lines) > 3000 {
			return layoutBlock{}, errors.New("SML output row limit exceeded")
		}
	}
	return result, nil
}

// weightedGridWidths allocates complete display-cell columns. Every visible
// pane gets at least 14 cells; otherwise the parent grid collapses to one
// column instead of rendering broken or overlapping borders.
func weightedGridWidths(width, gap, count int, weights []int) ([]int, bool) {
	if count < 1 || count > 8 || gap < 0 {
		return nil, false
	}
	available := width - gap*(count-1)
	minWidth := 14
	if count == 1 {
		minWidth = 12
	}
	if available < count*minWidth {
		return nil, false
	}
	total := 0
	for i := 0; i < count; i++ {
		weight := 1
		if i < len(weights) {
			weight = weights[i]
		}
		if weight < 1 || weight > 16 {
			return nil, false
		}
		total += weight
	}
	result := make([]int, count)
	remainingWidth, remainingWeight := available, total
	for i := 0; i < count; i++ {
		weight := 1
		if i < len(weights) {
			weight = weights[i]
		}
		if i == count-1 {
			result[i] = remainingWidth
		} else {
			result[i] = remainingWidth * weight / remainingWeight
		}
		if result[i] < minWidth {
			return nil, false
		}
		remainingWidth -= result[i]
		remainingWeight -= weight
	}
	return result, true
}

func (v View) renderLayoutNode(n Node, values map[string]string, width, depth int, path string) (layoutBlock, error) {
	if depth > maxDepth {
		return layoutBlock{}, errors.New("SML layout depth exceeded")
	}
	width = max(12, min(240, width))
	result := layoutBlock{}
	id := n.Attr("id", path)
	// Repeated component controls must get a stable, iteration-local identity.
	// Otherwise two <button id="inspect"> controls in a <repeat> produce a
	// duplicate hotspot error and the entire SML frame fails to render.
	if controlNode(n) && (strings.Contains(path, "/repeat:") ||
		strings.Contains(path, "/use:") || strings.Contains(path, "/include:")) {
		id = path + "/id:" + id
	}
	switch n.Kind {
	case "gradient", "derive":
		return result, nil
	case "column", "component", "screen":
		child, err := v.renderLayoutNodes(n.Children, values, width, depth, path)
		if err != nil {
			return result, err
		}
		result = child
	case "use":
		expanded, local, err := v.expandUse(n, values)
		if err != nil {
			return result, err
		}
		result, err = v.renderLayoutNodes(expanded, local, width, depth, path+"/use:"+n.Attr("name", ""))
		if err != nil {
			return result, err
		}
	case "include":
		component, err := v.component(n.Attr("name", ""))
		if err != nil {
			return result, err
		}
		result, err = v.renderLayoutNodes(component.Children, values, width, depth, path+"/include:"+n.Attr("name", ""))
		if err != nil {
			return result, err
		}
	case "if", "when":
		key := n.Attr("field", n.Attr("value", ""))
		value := values[key]
		matched := value != "" && value != "0" && value != "false" && value != "no"
		if n.Kind == "if" {
			matched = value == n.Attr("equals", "true")
		}
		yes := n.Children
		var no []Node
		for i, child := range n.Children {
			if child.Kind == "otherwise" {
				yes = n.Children[:i]
				no = child.Children
				break
			}
		}
		selected := yes
		if !matched {
			selected = no
		}
		var err error
		result, err = v.renderLayoutNodes(selected, values, width, depth, path+"/branch")
		if err != nil {
			return result, err
		}
	case "empty":
		if values[n.Attr("field", "")] != "" {
			return result, nil
		}
		var err error
		result, err = v.renderLayoutNodes(n.Children, values, width, depth, path+"/empty")
		if err != nil {
			return result, err
		}
	case "repeat":
		items, err := repeatItems(n, values[n.Attr("source", "")])
		if err != nil {
			return layoutBlock{}, err
		}
		seen := make(map[string]bool)
		for i, raw := range items {
			item := strings.TrimSpace(raw)
			identity, err := repeatInstanceID(n, item, i, seen)
			if err != nil {
				return layoutBlock{}, err
			}
			local := make(map[string]string, len(values)+2)
			for k, val := range values {
				local[k] = val
			}
			local[n.Attr("as", "item")] = item
			if indexAlias := n.Attr("index-as", ""); indexAlias != "" {
				local[indexAlias] = fmt.Sprintf("%d", i)
			}
			child, err := v.renderLayoutNodes(n.Children, local, width, depth, path+"/"+identity)
			if err != nil {
				return result, err
			}
			y := len(result.lines)
			result.lines = append(result.lines, child.lines...)
			offsetBlock(&result, child, 0, y)
			if len(result.lines) > 3000 {
				return layoutBlock{}, errors.New("SML repeat output limit exceeded")
			}
		}
	case "row", "grid":
		if len(n.Children) == 0 {
			return result, nil
		}
		cols := len(n.Children)
		if n.Kind == "grid" {
			cols = boundedInt(n.Attr("columns", "3"), 3, 1, 8)
		}
		if width < boundedInt(n.Attr("collapse-below", "65"), 65, 20, 240) {
			cols = 1
		}
		gap := boundedInt(n.Attr("gap", "2"), 2, 0, 8)
		cols = min(cols, len(n.Children))
		var weights []int
		if n.Kind == "grid" && n.Attr("weights", "") != "" {
			for _, part := range strings.Split(n.Attr("weights", ""), ",") {
				weights = append(weights, boundedInt(strings.TrimSpace(part), 1, 1, 16))
			}
		}
		if _, ok := weightedGridWidths(width, gap, cols, weights); !ok {
			cols = 1
		}
		for start := 0; start < len(n.Children); start += cols {
			count := min(cols, len(n.Children)-start)
			paneWidths, _ := weightedGridWidths(width, gap, count, weights)
			parts := make([]locatedBlock, 0, count)
			h := 0
			x := 0
			for col := 0; col < count; col++ {
				index := start + col
				childWidth := paneWidths[col]
				child, err := v.renderLayoutNode(n.Children[index], values, childWidth, depth+1, fmt.Sprintf("%s/%d:%s", path, index, n.Children[index].Kind))
				if err != nil {
					return layoutBlock{}, err
				}
				if len(child.lines) > h {
					h = len(child.lines)
				}
				parts = append(parts, locatedBlock{block: child, x: x})
				x += childWidth + gap
			}
			rendered := renderCanvas(width, h, parts)
			baseY := len(result.lines)
			result.lines = append(result.lines, rendered...)
			for _, part := range parts {
				offsetBlock(&result, part.block, part.x, baseY)
			}
			if len(result.lines) > 3000 {
				return layoutBlock{}, errors.New("SML layout exceeds output limit")
			}
		}
	case "panel":
		padding := boundedInt(n.Attr("padding", "1"), 1, 0, 3)
		noBorder := n.Attr("border", "round") == "none"
		innerWidth := width - 2 - 2*padding
		if noBorder {
			innerWidth = width - padding
		}
		child, err := v.renderLayoutNodes(n.Children, values, max(12, innerWidth), depth, path)
		if err != nil {
			return result, err
		}
		result.lines = stylePanel(n, child.lines, interpolate(n.Attr("title", ""), values), width)
		offsetX, offsetY := 1+padding, 1+max(0, padding-1)
		if noBorder {
			offsetX = padding
			offsetY = 0
			if n.Attr("title", "") != "" {
				offsetY++
			}
		}
		for _, spot := range child.spots {
			if !noBorder && n.Attr("align", "left") != "left" && spot.Line < len(child.lines) {
				slack := max(0, innerWidth-CellWidth(child.lines[spot.Line]))
				if n.Attr("align", "") == "center" {
					spot.Start += slack / 2
					spot.End += slack / 2
				}
				if n.Attr("align", "") == "right" {
					spot.Start += slack
					spot.End += slack
				}
			}
			// The nested renderer must use >=12 columns, but a narrow panel
			// with padding 2..3 can display fewer. Clip the actionable area
			// to the *visible* panel interior before translating it.
			// A clipped control must not invalidate the entire SML frame.
			spot.Start = max(0, min(innerWidth, spot.Start))
			spot.End = max(0, min(innerWidth, spot.End))
			if spot.End <= spot.Start {
				continue
			}
			spot.Line += offsetY
			spot.Start += offsetX
			spot.End += offsetX
			result.spots = append(result.spots, spot)
		}
		for _, el := range child.elements {
			el.Bounds.X += offsetX
			el.Bounds.Y += offsetY
			result.elements = append(result.elements, el)
		}
	case "section":
		title := trim(interpolate(n.Attr("title", ""), values), width)
		if title != "" {
			result.lines = append(result.lines, styleMetric(n, title, width), styleDivider(n, min(width, 60)))
		}
		child, err := v.renderLayoutNodes(n.Children, values, width, depth, path)
		if err != nil {
			return result, err
		}
		y := len(result.lines)
		result.lines = append(result.lines, child.lines...)
		offsetBlock(&result, child, 0, y)
	case "text":
		content := interpolate(n.Attr("value", strings.TrimSpace(n.Text)), values)
		padding := boundedInt(n.Attr("padding", "0"), 0, 0, 3)
		usable := max(1, width-2*padding-CellWidth(styleTone(n)))
		for _, line := range WrapText(content, usable) {
			result.lines = append(result.lines, styleMetric(n, line, width))
		}
	case "otherwise", "slot", "fill":
		return result, errors.New("unexpanded conditional, slot or fill")
	default:
		lines, err := v.node(n, values, width, depth)
		if err != nil {
			return result, err
		}
		result.lines = lines
		if controlNode(n) && len(lines) > 0 {
			interactive := controlEnabled(n, values)
			text := lines[0]
			lead := CellWidth(text) - CellWidth(strings.TrimLeft(text, " "))
			right := CellWidth(strings.TrimRight(text, " "))
			right = min(right, width)
			if interactive && right > lead {
				generic := ""
				if n.Kind == "button" || n.Kind == "link" {
					generic = n.Attr("action", "")
				}
				result.spots = append(result.spots, Hotspot{
					ID: id, Action: generic, Line: 0, Start: lead, End: right,
					Label: n.Attr("label", ""), Page: n.Attr("page", ""), Command: n.Attr("command", ""),
					Kind: n.Kind, Key: n.Attr("key", ""), Options: controlOptions(n),
					PluginID: n.Attr("plugin", ""), PluginAction: n.Attr("action", ""),
					Arg: n.Attr("arg", ""), Source: n.Attr("source", ""), Bindings: n.Attr("bindings", ""),
					Min: boundedInt(n.Attr("min", "0"), 0, 0, 1000000),
					Max: boundedInt(n.Attr("max", "100"), 100, 0, 1000000),
				})
			}
		}
	}
	if len(result.lines) > 3000 {
		return layoutBlock{}, errors.New("SML render output exceeds limit")
	}
	result.mark(n, id, width)
	return result, nil
}

// RenderFrame lays out the actual tree, rather than reverse-matching control
// text from rendered output. Both cells and hotspots use display cell columns.
func (v View) RenderFrame(values map[string]string, width int) (Frame, error) {
	if width < 12 || width > 240 {
		return Frame{}, errors.New("SML frame width must be 12..240")
	}
	resolved := v.ResolveBindings(values)
	block, err := v.renderLayoutNodes(v.Nodes, resolved, width, 0, "screen:"+v.Name)
	if err != nil {
		return Frame{}, err
	}
	frame := Frame{Width: width, Height: len(block.lines), Lines: block.lines, Elements: block.elements, Hotspots: block.spots}
	if frame.Height > 3000 {
		return Frame{}, errors.New("frame too tall")
	}
	frame.Cells = make([][]Cell, frame.Height)
	for y, line := range frame.Lines {
		frame.Cells[y] = cellLine(line, width)
	}
	idMap := map[string]bool{}
	for _, spot := range frame.Hotspots {
		if idMap[spot.ID] {
			return Frame{}, fmt.Errorf("duplicate visible control id %q", spot.ID)
		}
		idMap[spot.ID] = true
		if spot.Line < 0 || spot.Line >= frame.Height || spot.Start < 0 || spot.End > width {
			return Frame{}, fmt.Errorf("invalid control bounds for %q", spot.ID)
		}
		for x := spot.Start; x < spot.End; x++ {
			frame.Cells[spot.Line][x].ElementID = spot.ID
		}
	}
	return frame, nil
}
