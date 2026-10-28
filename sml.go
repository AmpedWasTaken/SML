// Package sml provides a Go-first declarative terminal view engine.
//
// The public API is intentionally small. Rendering, parsing and validation
// live in internal/engine, allowing a tidy repository without changing
// github.com/AmpedWasTaken/SML imports or public type identities.
package sml

import "github.com/AmpedWasTaken/SML/internal/engine"

// Public types remain aliases, not copies, so fields, methods, slices and
// return types are fully compatible with the implementation.
type (
	// Node represents a validated markup element.
	Node = engine.Node
	// View is a parsed screen that can be rendered with host bindings.
	View = engine.View
	// Cell is one terminal display cell.
	Cell = engine.Cell
	// Rect describes terminal cell bounds.
	Rect = engine.Rect
	// Element records a rendered element and its bounds.
	Element = engine.Element
	// Frame contains rendered lines, cells, elements and interaction hotspots.
	Frame = engine.Frame
	// Hotspot describes an interactive control without executing actions.
	Hotspot = engine.Hotspot
	// CellRun is one changed row interval of a terminal-cell diff.
	CellRun = engine.CellRun
	// FrameDiff is a pure frame comparison result; hosts choose how to paint it.
	FrameDiff = engine.FrameDiff
	// ValueKind identifies a typed host-supplied presentation scalar.
	ValueKind = engine.ValueKind
	// Value stores a bounded, inert typed presentation scalar.
	Value = engine.Value
	// BindingUpdate is one atomic host-side state operation.
	BindingUpdate = engine.BindingUpdate
	// BindingChange describes the committed revision and changed keys.
	BindingChange = engine.BindingChange
	// BindingSnapshot is a detached state snapshot.
	BindingSnapshot = engine.BindingSnapshot
	// BindingState serializes typed presentation updates and snapshots.
	BindingState = engine.BindingState
)

// Parse checks SML markup and returns a view without executing it.
func Parse(data []byte) (View, error) { return engine.Parse(data) }

// ParseFile validates a regular on-disk SML view.
func ParseFile(path string) (View, error) { return engine.ParseFile(path) }

// LoadDir validates a directory of views and local reusable components.
func LoadDir(dir string) (map[string]View, error) { return engine.LoadDir(dir) }

// ParseActionBindings validates declarative action argument bindings.
func ParseActionBindings(spec string) (map[string]string, error) {
	return engine.ParseActionBindings(spec)
}

// CellWidth measures Unicode terminal text in display cells.
func CellWidth(s string) int { return engine.CellWidth(s) }

// ClipCells clips text without splitting grapheme clusters.
func ClipCells(s string, width int) string { return engine.ClipCells(s, width) }

// PadCells clips and pads text to the requested terminal display width.
func PadCells(s string, width int) string { return engine.PadCells(s, width) }

// WrapText wraps text at grapheme boundaries and word boundaries.
func WrapText(s string, width int) []string { return engine.WrapText(s, width) }

// DiffFrames compares two structural frames without terminal I/O.
// Changes to layout and interactive actions are reported independently from
// changed glyph runs. A new/different-size frame requests a full redraw.
func DiffFrames(previous, current Frame) (FrameDiff, error) {
	return engine.DiffFrames(previous, current)
}

const (
	KindNull   ValueKind = engine.KindNull
	KindString ValueKind = engine.KindString
	KindInt    ValueKind = engine.KindInt
	KindFloat  ValueKind = engine.KindFloat
	KindBool   ValueKind = engine.KindBool
)

// Typed value constructors create inert presentation data, never expressions.
func NullValue() Value           { return engine.NullValue() }
func StringValue(s string) Value { return engine.StringValue(s) }
func IntValue(n int64) Value     { return engine.IntValue(n) }
func FloatValue(n float64) Value { return engine.FloatValue(n) }
func BoolValue(b bool) Value     { return engine.BoolValue(b) }

// NewBindingState creates a bounded state store with a detached initial map.
func NewBindingState(initial map[string]Value) (*BindingState, error) {
	return engine.NewBindingState(initial)
}

// ValidateBindingValues checks typed presentation values without changing them.
func ValidateBindingValues(values map[string]Value) error {
	return engine.ValidateBindingValues(values)
}
