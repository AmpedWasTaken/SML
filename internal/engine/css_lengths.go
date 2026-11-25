package engine

import (
	"fmt"
	"strconv"
	"strings"
)

type lengthExpression struct {
	source string
	pos    int
	width  int
}

func isLengthExpression(value string) bool {
	for _, name := range []string{"calc(", "min(", "max(", "clamp("} {
		if strings.HasPrefix(value, name) {
			return true
		}
	}
	return false
}

// Evaluate in hundredths of a cell; round only at the outermost boundary.
func evaluateLengthExpression(value string, available int) (int, error) {
	if len(value) > 64 || !isLengthExpression(value) {
		return 0, fmt.Errorf("invalid SML length expression")
	}
	p := lengthExpression{source: value, width: max(0, min(240, available))}
	result, err := p.term(0)
	p.space()
	if err != nil {
		return 0, err
	}
	if p.pos != len(p.source) {
		return 0, fmt.Errorf("unexpected trailing length expression")
	}
	return result / 100, nil
}

func (p *lengthExpression) space() {
	for p.pos < len(p.source) && (p.source[p.pos] == ' ' || p.source[p.pos] == '\t') {
		p.pos++
	}
}

func (p *lengthExpression) consume(c byte) bool {
	p.space()
	if p.pos < len(p.source) && p.source[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *lengthExpression) term(depth int) (int, error) {
	p.space()
	if depth > 4 || p.pos == len(p.source) {
		return 0, fmt.Errorf("empty or deeply nested length")
	}
	for _, name := range []string{"calc", "min", "max", "clamp"} {
		if strings.HasPrefix(p.source[p.pos:], name+"(") {
			p.pos += len(name) + 1
			return p.function(name, depth+1)
		}
	}
	start := p.pos
	for p.pos < len(p.source) && p.source[p.pos] >= '0' && p.source[p.pos] <= '9' {
		p.pos++
	}
	if start == p.pos || p.pos-start > 3 {
		return 0, fmt.Errorf("expected bounded integer length")
	}
	value, err := strconv.Atoi(p.source[start:p.pos])
	if err != nil {
		return 0, err
	}
	if p.pos < len(p.source) && p.source[p.pos] == '%' {
		p.pos++
		if value > 100 {
			return 0, fmt.Errorf("percentage must be 0..100")
		}
		return value * p.width, nil
	}
	if strings.HasPrefix(p.source[p.pos:], "ch") {
		p.pos += 2
	} else if value != 0 {
		return 0, fmt.Errorf("nonzero length requires ch or %%")
	}
	if value > 240 {
		return 0, fmt.Errorf("length must be 0..240 cells")
	}
	return value * 100, nil
}

func (p *lengthExpression) function(name string, depth int) (int, error) {
	first, err := p.term(depth)
	if err != nil {
		return 0, err
	}
	if name == "calc" {
		result := first
		for !p.consume(')') {
			sign := 1
			if p.consume('-') {
				sign = -1
			} else if !p.consume('+') {
				return 0, fmt.Errorf("calc supports addition and subtraction only")
			}
			next, err := p.term(depth)
			if err != nil {
				return 0, err
			}
			result += sign * next
		}
		return result, nil
	}
	args := []int{first}
	for p.consume(',') {
		if len(args) >= 4 {
			return 0, fmt.Errorf("length function accepts at most four arguments")
		}
		next, err := p.term(depth)
		if err != nil {
			return 0, err
		}
		args = append(args, next)
	}
	if !p.consume(')') || len(args) < 2 {
		return 0, fmt.Errorf("length function requires comma-separated arguments")
	}
	switch name {
	case "clamp":
		if len(args) != 3 {
			return 0, fmt.Errorf("clamp requires exactly three arguments")
		}
		return max(args[0], min(args[1], args[2])), nil
	case "min":
		for _, next := range args[1:] {
			first = min(first, next)
		}
	case "max":
		for _, next := range args[1:] {
			first = max(first, next)
		}
	}
	return first, nil
}
