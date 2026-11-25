package engine

import (
	"fmt"
	"sync"
)

// AssetStore serializes reloads and publishes a complete validated revision.
// A failed reload leaves the previous views and revision untouched. It does
// not watch files, perform terminal I/O, or change host action permissions.
// Media remains host-managed and is not part of the markup snapshot.
type AssetStore struct {
	mu       sync.RWMutex
	reloadMu sync.Mutex
	dir      string
	required []string
	views    map[string]View
	revision uint64
}

func NewAssetStore(dir string, required ...string) (*AssetStore, error) {
	s := &AssetStore{dir: dir, required: append([]string(nil), required...)}
	if _, err := s.Reload(); err != nil {
		return nil, err
	}
	return s, nil
}

// Reload validates the candidate before swapping the active revision.
func (s *AssetStore) Reload() (uint64, error) {
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()
	views, err := LoadDir(s.dir)
	if err == nil {
		for _, name := range s.required {
			if _, ok := views[name]; !ok {
				err = fmt.Errorf("required SML view %q is missing", name)
				break
			}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		return s.revision, err
	}
	s.views = views
	s.revision++
	return s.revision, nil
}

// View returns a detached view and its revision. Component nodes are private
// and shared read-only by the renderer.
func (s *AssetStore) View(name string) (View, uint64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.views[name]
	v.Nodes = cloneAssetNodes(v.Nodes)
	return v, s.revision, ok
}

func cloneAssetNodes(nodes []Node) []Node {
	if nodes == nil {
		return nil
	}
	out := make([]Node, len(nodes))
	for i, n := range nodes {
		out[i] = n
		if n.Attrs != nil {
			out[i].Attrs = make(map[string]string, len(n.Attrs))
			for k, value := range n.Attrs {
				out[i].Attrs[k] = value
			}
		}
		out[i].Children = cloneAssetNodes(n.Children)
	}
	return out
}
