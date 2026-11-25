package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const maxAssetEntries = 1024
const maxAssetViews = 256

// LoadDir validates nested views and captures their referenced components.
// Screen names remain globally unique, independent of folder organization.
func LoadDir(dir string) (map[string]View, error) {
	directory := filepath.Join(dir, "views")
	info, err := os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]View{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("SML views must be a real directory")
	}
	views := map[string]View{}
	sources := map[string]string{}
	components := map[string]Node{}
	entries := 0
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		entries++
		if entries > maxAssetEntries {
			return errors.New("too many SML assets")
		}
		rel, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		if strings.Count(rel, string(filepath.Separator)) > maxDepth {
			return fmt.Errorf("SML asset nesting limit exceeded: %s", rel)
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sml") {
			return nil
		}
		name := strings.TrimSuffix(entry.Name(), ".sml")
		if !validName.MatchString(name) {
			return fmt.Errorf("invalid SML filename: %s", rel)
		}
		if previous, exists := sources[name]; exists {
			return fmt.Errorf("duplicate SML view %q: %s and %s", name, previous, rel)
		}
		if len(views) >= maxAssetViews {
			return errors.New("too many SML views")
		}
		view, err := ParseFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if view.Name != name {
			return fmt.Errorf("SML view name/file mismatch: %s", rel)
		}
		view.assets = dir
		view.components = components
		if err := validateIncludes(view, view.Nodes, 0, map[string]bool{}); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		sources[name], views[name] = rel, view
		return nil
	})
	if err != nil {
		return nil, err
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
			part, exists := v.components[name]
			if !exists {
				var err error
				part, err = v.loadComponent(name)
				if err != nil {
					return fmt.Errorf("component %q: %w", name, err)
				}
				v.components[name] = part
			}
			if node.Kind == "use" {
				if _, _, err := v.expandUse(node, nil); err != nil {
					return err
				}
			}
			seen[name] = true
			if err := validateIncludes(v, part.Children, depth+1, seen); err != nil {
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
