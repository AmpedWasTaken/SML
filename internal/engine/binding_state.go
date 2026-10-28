package engine

import (
	"errors"
	"sort"
	"sync"
)

// BindingUpdate is one proposed host-owned presentation change. Delete removes
// a key instead of storing an empty or null value.
type BindingUpdate struct {
	Key    string
	Value  Value
	Delete bool
}

// BindingChange reports a committed revision and the keys whose values
// actually changed, sorted lexicographically for deterministic host handling.
type BindingChange struct {
	Revision    uint64
	ChangedKeys []string
}

// BindingSnapshot is a detached, host-owned copy. Mutating its map cannot
// change a BindingState or any other snapshot.
type BindingSnapshot struct {
	Revision uint64
	Values   map[string]Value
}

// BindingState is a bounded presentation store for hosts that prefer typed
// updates. It does not execute code, render automatically or spawn goroutines.
// Calls are mutex-serialized; rendering happens on separately copied snapshots.
type BindingState struct {
	mu       sync.RWMutex
	values   map[string]Value
	revision uint64
}

// NewBindingState validates and copies the initial values. The host must
// independently decide which authenticated data may enter this store.
func NewBindingState(initial map[string]Value) (*BindingState, error) {
	if err := ValidateBindingValues(initial); err != nil {
		return nil, err
	}
	return &BindingState{values: copyBindingValues(initial)}, nil
}

func copyBindingValues(input map[string]Value) map[string]Value {
	copied := make(map[string]Value, len(input))
	for key, value := range input {
		copied[key] = value
	}
	return copied
}

// Snapshot returns a consistent detached copy and its revision. Call
// View.RenderTypedFrame(snapshot.Values, width) on the rendering goroutine.
func (s *BindingState) Snapshot() BindingSnapshot {
	if s == nil {
		return BindingSnapshot{Values: map[string]Value{}}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return BindingSnapshot{Revision: s.revision, Values: copyBindingValues(s.values)}
}

// Apply validates the entire batch, then commits all changes atomically.
// A rejected batch leaves the original revision and values unchanged.
// Zero-effect batches do not increment the revision.
func (s *BindingState) Apply(updates ...BindingUpdate) (BindingChange, error) {
	if s == nil {
		return BindingChange{}, errors.New("nil SML binding state")
	}
	if len(updates) > maxPresentationBindings {
		return BindingChange{}, errors.New("too many SML presentation updates")
	}
	seen := make(map[string]bool, len(updates))
	for _, update := range updates {
		if !bindingName.MatchString(update.Key) || len(update.Key) >= 8 && update.Key[:8] == "derived." {
			return BindingChange{}, errors.New("invalid or protected SML presentation binding name")
		}
		if seen[update.Key] {
			return BindingChange{}, errors.New("duplicate SML presentation update")
		}
		seen[update.Key] = true
		if update.Delete {
			if update.Value != (Value{}) {
				return BindingChange{}, errors.New("deleted SML presentation binding must not have a value")
			}
		} else if err := update.Value.validate(); err != nil {
			return BindingChange{}, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	next := copyBindingValues(s.values)
	changed := make([]string, 0, len(updates))
	for _, update := range updates {
		before, exists := next[update.Key]
		if update.Delete {
			if exists {
				delete(next, update.Key)
				changed = append(changed, update.Key)
			}
			continue
		}
		if !exists || before != update.Value {
			next[update.Key] = update.Value
			changed = append(changed, update.Key)
		}
	}
	if err := ValidateBindingValues(next); err != nil {
		return BindingChange{}, err
	}
	if len(changed) == 0 {
		return BindingChange{Revision: s.revision}, nil
	}
	sort.Strings(changed)
	s.values = next
	s.revision++
	return BindingChange{Revision: s.revision, ChangedKeys: changed}, nil
}
