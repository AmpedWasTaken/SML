package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadDir validates and loads named views and their local reusable components.
func LoadDir(dir string) (map[string]View, error) {
	directory := filepath.Join(dir, "views")
	items, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]View{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(items) > 100 {
		return nil, errors.New("too many SML assets")
	}
	views := map[string]View{}
	for _, entry := range items {
		if !strings.HasSuffix(entry.Name(), ".sml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".sml")
		if !validName.MatchString(name) {
			return nil, fmt.Errorf("invalid SML filename: %s", entry.Name())
		}
		path := filepath.Join(directory, entry.Name())
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > maxFileSize {
			return nil, fmt.Errorf("invalid SML asset: %s", entry.Name())
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		view, err := Parse(content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		if view.Name != name {
			return nil, fmt.Errorf("SML view name/file mismatch: %s", name)
		}
		view.assets = dir
		// Compile and resolve includes before publishing a new immutable snapshot.
		if err = validateIncludes(view, view.Nodes, 0, map[string]bool{}); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		views[name] = view
	}
	return views, nil
}
func validateIncludes(v View, nodes []Node, depth int, seen map[string]bool) error {
	if depth > maxDepth {
		return errors.New("SML include depth exceeded")
	}
	for _, node := range nodes {
		if node.Kind == "include" || node.Kind == "use" {
			name := node.Attr("name", "")
			if seen[name] {
				return fmt.Errorf("SML include cycle: %s", name)
			}
			part, err := v.component(name)
			if err != nil {
				return err
			}
			seen[name] = true
			if err = validateIncludes(v, part.Children, depth+1, seen); err != nil {
				return err
			}
			delete(seen, name)
		}
		if err := validateIncludes(v, node.Children, depth+1, seen); err != nil {
			return err
		}
	}
	return nil
}
