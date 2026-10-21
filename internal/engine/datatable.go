package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// DataTable is a bounded presentation-only component. Dynamic rows are JSON
// passed through the existing SML value bindings, never commands or scripts.
type dataColumn struct {
	key, label string
	width      int
}

var tableStyles = map[string]bool{"compact": true, "compact-lite": true, "compact-classic": true, "default": true, "markdown": true, "rounded": true, "unicode": true, "ascii": true}

func validateTable(n Node) error {
	if n.Kind == "column" {
		return errors.New("table columns must be inside a table")
	}
	if n.Kind != "table" {
		return nil
	}
	allowed := map[string]bool{"source": true, "style": true, "page": true, "page-size": true, "sort": true, "order": true, "filter": true, "empty": true}
	for k := range n.Attrs {
		if !allowed[k] {
			return fmt.Errorf("unknown table attribute %q", k)
		}
	}
	if !tableStyles[n.Attr("style", "compact-lite")] {
		return errors.New("unknown table style")
	}
	if n.Attr("source", "") == "" {
		return errors.New("table requires a source binding")
	}
	if len(n.Children) == 0 || len(n.Children) > 8 {
		return errors.New("table needs 1..8 columns")
	}
	keys := map[string]bool{}
	for _, child := range n.Children {
		if child.Kind != "column" || len(child.Children) != 0 || strings.TrimSpace(child.Text) != "" {
			return errors.New("table only accepts empty column elements")
		}
		for k := range child.Attrs {
			if k != "key" && k != "label" && k != "width" {
				return fmt.Errorf("unknown table column attribute %q", k)
			}
		}
		key := child.Attr("key", "")
		if !validName.MatchString(key) || keys[key] {
			return errors.New("invalid/duplicate table column key")
		}
		keys[key] = true
		if len(child.Attr("label", "")) > 48 {
			return errors.New("table column label too long")
		}
		w, err := strconv.Atoi(child.Attr("width", "16"))
		if err != nil || w < 4 || w > 60 {
			return errors.New("table column width must be 4..60")
		}
	}
	if sortKey := n.Attr("sort", ""); sortKey != "" && !keys[sortKey] {
		return errors.New("sort column not found")
	}
	order := n.Attr("order", "asc")
	if order != "asc" && order != "desc" {
		return errors.New("table order must be asc or desc")
	}
	for _, field := range []string{"page", "page-size"} {
		if raw := n.Attr(field, ""); raw != "" {
			if binding.MatchString(raw) && binding.FindString(raw) == raw {
				continue
			}
			i, err := strconv.Atoi(raw)
			if err != nil || i < 1 || i > 100 {
				return fmt.Errorf("table %s must be 1..100 or a data binding", field)
			}
		}
	}
	return nil
}

func renderDataTable(n Node, values map[string]string, width int) ([]string, error) {
	raw := values[n.Attr("source", "")]
	if len(raw) > 32768 {
		return nil, errors.New("table source exceeds 32KiB")
	}
	rows := []map[string]string{}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &rows); err != nil {
			return nil, errors.New("table source must be a JSON array of string-keyed objects")
		}
	}
	if len(rows) > 200 {
		return nil, errors.New("table exceeds 200 rows")
	}
	cols := make([]dataColumn, 0, len(n.Children))
	for _, child := range n.Children {
		cols = append(cols, dataColumn{key: child.Attr("key", ""), label: safe(child.Attr("label", child.Attr("key", ""))), width: boundedInt(child.Attr("width", "16"), 16, 4, 60)})
	}
	if filter := strings.ToLower(interpolate(n.Attr("filter", ""), values)); filter != "" {
		filtered := make([]map[string]string, 0, len(rows))
		for _, row := range rows {
			for _, col := range cols {
				if strings.Contains(strings.ToLower(row[col.key]), filter) {
					filtered = append(filtered, row)
					break
				}
			}
		}
		rows = filtered
	}
	if key := n.Attr("sort", ""); key != "" {
		desc := n.Attr("order", "asc") == "desc"
		sort.SliceStable(rows, func(i, j int) bool {
			a, b := strings.ToLower(rows[i][key]), strings.ToLower(rows[j][key])
			if desc {
				return a > b
			}
			return a < b
		})
	}
	page := boundedInt(interpolate(n.Attr("page", "1"), values), 1, 1, 100)
	pageSize := boundedInt(interpolate(n.Attr("page-size", "20"), values), 20, 1, 100)
	total := len(rows)
	first := (page - 1) * pageSize
	pages := (total + pageSize - 1) / pageSize
	if pages > 0 && page > pages {
		page = pages
		first = (page - 1) * pageSize
	}
	last := min(total, first+pageSize)
	if total == 0 {
		return []string{interpolate(n.Attr("empty", "No records available"), values)}, nil
	}
	rows = rows[first:last]
	// Compute a bounded width including border and spacing glyphs. Every
	// style is deterministic and does not emit control sequences.
	style := n.Attr("style", "compact-lite")
	boxed := style == "rounded" || style == "default" || style == "unicode"
	markdown := style == "markdown"
	ascii := style == "ascii" || style == "compact-classic"
	separator := " │ "
	left, right := "", ""
	if ascii {
		separator = " | "
	}
	if markdown {
		separator = " | "
		left = "| "
		right = " |"
	}
	if boxed {
		left = "│ "
		right = " │"
	}
	if style == "unicode" {
		separator = " ┃ "
		left = "┃ "
		right = " ┃"
	}
	paddingCost := utf8.RuneCountInString(left) + utf8.RuneCountInString(right) + (len(cols)-1)*utf8.RuneCountInString(separator)
	budget := max(len(cols), width-paddingCost)
	for i := range cols {
		remaining := len(cols) - i
		allocation := budget / remaining
		cols[i].width = max(1, min(cols[i].width, allocation))
		budget -= cols[i].width
	}
	formatRow := func(cells []string) string {
		parts := make([]string, len(cols))
		for i, value := range cells {
			parts[i] = pad(truncateCell(value, cols[i].width), cols[i].width)
		}
		return trim(left+strings.Join(parts, separator)+right, width)
	}
	border := func(prefix, joint, suffix, line string) string {
		pieces := make([]string, len(cols))
		for i, col := range cols {
			pieces[i] = strings.Repeat(line, col.width+2)
		}
		return trim(prefix+strings.Join(pieces, joint)+suffix, width)
	}
	headers := make([]string, len(cols))
	for i, col := range cols {
		headers[i] = col.label
	}
	output := []string{}
	if boxed {
		switch style {
		case "rounded":
			output = append(output, border("╭", "┬", "╮", "─"))
		case "unicode":
			output = append(output, border("┏", "┳", "┓", "━"))
		default:
			output = append(output, border("┌", "┬", "┐", "─"))
		}
	}
	output = append(output, formatRow(headers))
	switch style {
	case "rounded":
		output = append(output, border("├", "┼", "┤", "─"))
	case "unicode":
		output = append(output, border("┣", "╋", "┫", "━"))
	case "default":
		output = append(output, border("├", "┼", "┤", "─"))
	case "markdown":
		pieces := make([]string, len(cols))
		for i, col := range cols {
			pieces[i] = strings.Repeat("-", col.width)
		}
		output = append(output, trim("| "+strings.Join(pieces, " | ")+" |", width))
	case "ascii", "compact-classic":
		pieces := make([]string, len(cols))
		for i, col := range cols {
			pieces[i] = strings.Repeat("-", col.width)
		}
		output = append(output, trim(strings.Join(pieces, "-+-"), width))
	case "compact-lite":
		pieces := make([]string, len(cols))
		for i, col := range cols {
			pieces[i] = strings.Repeat("─", col.width)
		}
		output = append(output, trim(strings.Join(pieces, " ─ "), width))
	}
	for _, row := range rows {
		cells := make([]string, len(cols))
		for i, col := range cols {
			cells[i] = safe(row[col.key])
		}
		output = append(output, formatRow(cells))
	}
	if boxed {
		switch style {
		case "rounded":
			output = append(output, border("╰", "┴", "╯", "─"))
		case "unicode":
			output = append(output, border("┗", "┻", "┛", "━"))
		default:
			output = append(output, border("└", "┴", "┘", "─"))
		}
	}
	output = append(output, fmt.Sprintf("Page %d/%d · %d records", page, pages, total))
	return output, nil
}
func truncateCell(s string, w int) string {
	s = safe(s)
	s = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(s)
	if CellWidth(s) <= w {
		return s
	}
	if w < 2 {
		return trim(s, w)
	}
	return ClipCells(s, w-1) + "…"
}
