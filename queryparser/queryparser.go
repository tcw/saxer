// Package queryparser parses saxer queries.
//
// A query is a list of steps separated by '/'. Each step is
//
//	name[~text][?key[=value][&key[=value]]...]
//
// where name is an element name, or '*' or nothing for any element, text
// is the text the element must have (last step only), and the part after
// '?' lists attributes the element must have. A backslash makes the next
// character literal, so "a\&b" is the three characters a&b.
package queryparser

import (
	"errors"
	"fmt"
	"strings"
)

// MaxSteps is the largest number of steps in a query.
const MaxSteps = 64

// Attr is an attribute condition. Without a value only the key must exist.
type Attr struct {
	Key      string
	Value    string
	HasValue bool
}

// Step is one '/' separated part of a query.
type Step struct {
	Name    string // "" matches any element
	Text    string
	HasText bool
	Attrs   []Attr
}

// ErrEmpty is returned for a query without any step.
var ErrEmpty = errors.New("empty query")

// Parse parses query into its steps.
func Parse(query string) ([]Step, error) {
	parts, err := split(query)
	if err != nil {
		return nil, fmt.Errorf("invalid query %q: %w", query, err)
	}
	var steps []Step
	for _, p := range parts {
		step, err := parseStep(p)
		if err != nil {
			return nil, fmt.Errorf("invalid query %q: %w", query, err)
		}
		if step.Name == "" && !step.HasText && len(step.Attrs) == 0 && !p.wildcard {
			continue // empty step, as in "/a" or "a//b"
		}
		steps = append(steps, step)
	}
	switch {
	case len(steps) == 0:
		return nil, ErrEmpty
	case len(steps) > MaxSteps:
		return nil, fmt.Errorf("invalid query %q: more than %d levels", query, MaxSteps)
	}
	for _, s := range steps[:len(steps)-1] {
		if s.HasText {
			return nil, fmt.Errorf("invalid query %q: text match (~) is only allowed on the last level", query)
		}
	}
	return steps, nil
}

// token is a character of a query, and whether it was escaped.
type token struct {
	c       rune
	escaped bool
}

type part struct {
	tokens   []token
	wildcard bool
}

// split splits query into steps on unescaped '/', resolving escapes.
func split(query string) ([]part, error) {
	var parts []part
	var cur []token
	escaped := false
	for _, c := range query {
		switch {
		case escaped:
			cur = append(cur, token{c, true})
			escaped = false
		case c == '\\':
			escaped = true
		case c == '/':
			parts = append(parts, part{tokens: cur})
			cur = nil
		default:
			cur = append(cur, token{c, false})
		}
	}
	if escaped {
		return nil, errors.New("ends with a lone '\\'")
	}
	parts = append(parts, part{tokens: cur})
	for i := range parts {
		name := strings.TrimSpace(literal(parts[i].tokens))
		parts[i].wildcard = name == "*" && len(parts[i].tokens) > 0
	}
	return parts, nil
}

// literal returns the tokens as a string.
func literal(tokens []token) string {
	var b strings.Builder
	for _, t := range tokens {
		b.WriteRune(t.c)
	}
	return b.String()
}

// cut splits tokens at the first unescaped sep.
func cut(tokens []token, sep rune) (before, after []token, found bool) {
	for i, t := range tokens {
		if t.c == sep && !t.escaped {
			return tokens[:i], tokens[i+1:], true
		}
	}
	return tokens, nil, false
}

// splitAll splits tokens at every unescaped sep.
func splitAll(tokens []token, sep rune) [][]token {
	var out [][]token
	for {
		before, after, found := cut(tokens, sep)
		out = append(out, before)
		if !found {
			return out
		}
		tokens = after
	}
}

func parseStep(p part) (Step, error) {
	var step Step
	head, attrs, hasAttrs := cut(p.tokens, '?')
	if _, _, found := cut(attrs, '?'); found {
		return step, fmt.Errorf("more than one '?' in %q", literal(p.tokens))
	}
	name, text, hasText := cut(head, '~')
	if _, _, found := cut(text, '~'); found {
		return step, fmt.Errorf("more than one '~' in %q", literal(p.tokens))
	}
	step.Name = strings.TrimSpace(literal(name))
	if step.Name == "*" && !name[len(name)-1].escaped {
		step.Name = ""
	}
	if hasText {
		step.Text = literal(text)
		step.HasText = true
	}
	if !hasAttrs || len(attrs) == 0 {
		return step, nil
	}
	for _, attr := range splitAll(attrs, '&') {
		keyTokens, valueTokens, hasValue := cut(attr, '=')
		if _, _, found := cut(valueTokens, '='); found {
			return step, fmt.Errorf("more than one '=' in attribute %q", literal(attr))
		}
		key := strings.TrimSpace(literal(keyTokens))
		if key == "" {
			return step, fmt.Errorf("attribute without a name in %q", literal(p.tokens))
		}
		step.Attrs = append(step.Attrs, Attr{Key: key, Value: literal(valueTokens), HasValue: hasValue})
	}
	return step, nil
}
