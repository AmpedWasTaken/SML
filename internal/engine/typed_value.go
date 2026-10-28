package engine

import (
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ValueKind identifies the type of a host-supplied presentation value.
// Values are inert scalars; SML never evaluates them as expressions.
type ValueKind uint8

const (
	KindNull ValueKind = iota
	KindString
	KindInt
	KindFloat
	KindBool
)

// Value is an immutable, comparable presentation scalar. Zero Value means null.
// Structured records and lists are deliberately not implicit string bindings;
// they require a separate bounded data-provider contract.
type Value struct {
	kind    ValueKind
	text    string
	integer int64
	number  float64
	boolean bool
}

func NullValue() Value                   { return Value{} }
func StringValue(s string) Value         { return Value{kind: KindString, text: s} }
func IntValue(n int64) Value             { return Value{kind: KindInt, integer: n} }
func FloatValue(n float64) Value         { return Value{kind: KindFloat, number: n} }
func BoolValue(b bool) Value             { return Value{kind: KindBool, boolean: b} }
func (v Value) Kind() ValueKind          { return v.kind }
func (v Value) Text() (string, bool)     { return v.text, v.kind == KindString }
func (v Value) Int64() (int64, bool)     { return v.integer, v.kind == KindInt }
func (v Value) Float64() (float64, bool) { return v.number, v.kind == KindFloat }
func (v Value) Bool() (bool, bool)       { return v.boolean, v.kind == KindBool }

// String returns the deterministic plain-text representation used by the
// existing string-map renderer. Untrusted escape sequences are filtered later
// by the established SML interpolation and component safety checks.
func (v Value) String() string {
	switch v.kind {
	case KindNull:
		return ""
	case KindString:
		return v.text
	case KindInt:
		return strconv.FormatInt(v.integer, 10)
	case KindFloat:
		return strconv.FormatFloat(v.number, 'g', -1, 64)
	case KindBool:
		if v.boolean {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

const maxPresentationValueBytes = 4096
const maxPresentationBindings = 256
const maxPresentationBindingBytes = 128 * 1024

func (v Value) validate() error {
	switch v.kind {
	case KindNull, KindInt, KindBool:
		return nil
	case KindString:
		if len(v.text) > maxPresentationValueBytes || !utf8.ValidString(v.text) {
			return errors.New("SML presentation string must be valid UTF-8 and at most 4096 bytes")
		}
		return nil
	case KindFloat:
		if math.IsNaN(v.number) || math.IsInf(v.number, 0) {
			return errors.New("SML presentation number must be finite")
		}
		return nil
	default:
		return errors.New("unsupported SML presentation value kind")
	}
}

// ValidateBindingValues checks a complete typed snapshot without mutating it.
// Derived bindings are protected: the existing derivation engine computes them
// from host-approved inputs, so callers cannot supply derived.* directly.
func ValidateBindingValues(values map[string]Value) error {
	if len(values) > maxPresentationBindings {
		return errors.New("too many SML presentation bindings")
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	total := 0
	for _, key := range keys {
		if !bindingName.MatchString(key) || strings.HasPrefix(key, "derived.") {
			return errors.New("invalid or protected SML presentation binding name")
		}
		val := values[key]
		if err := val.validate(); err != nil {
			return err
		}
		total += len(key) + len(val.String())
		if total > maxPresentationBindingBytes {
			return errors.New("SML presentation bindings exceed 128KiB")
		}
	}
	return nil
}

func stringBindings(values map[string]Value) (map[string]string, error) {
	if err := ValidateBindingValues(values); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(values))
	for k, v := range values {
		out[k] = v.String()
	}
	return out, nil
}

// RenderTypedFrame adapts validated scalar values into the existing renderer.
// All v1-v3 RenderFrame and RenderInteractive semantics remain unchanged.
func (v View) RenderTypedFrame(values map[string]Value, width int) (Frame, error) {
	bindings, err := stringBindings(values)
	if err != nil {
		return Frame{}, err
	}
	return v.RenderFrame(bindings, width)
}
