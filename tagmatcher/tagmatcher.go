// Package tagmatcher decides which steps of a query an element satisfies.
package tagmatcher

import (
	"bytes"
	"math/bits"
	"strings"
	"unicode/utf8"

	"github.com/tcw/saxer/markup"
	"github.com/tcw/saxer/queryparser"
)

// Options control how query names, attributes and text are compared with
// the document. The zero value matches exactly and case sensitively.
type Options struct {
	Contains        bool // a query name, value or text matches when it is a substring
	CaseInsensitive bool
	OmitNamespace   bool // ignore "ns:" prefixes on tag names
}

type attr struct {
	key, value []byte
	hasValue   bool
}

type step struct {
	name    []byte // nil matches any element
	attrs   []attr
	text    []byte
	hasText bool
}

// Matcher matches elements against the steps of a query. It keeps scratch
// buffers, so it is not safe for concurrent use.
type Matcher struct {
	steps []step
	opts  Options

	attrs   []markup.Attr // attributes of the tag being matched
	lower   []byte        // scratch for case folding
	unesc   []byte        // scratch for unescaped attribute values
	textBuf []byte        // scratch for element text
}

// New parses query and returns a Matcher for it.
func New(query string, opts Options) (*Matcher, error) {
	steps, err := queryparser.Parse(query)
	if err != nil {
		return nil, err
	}
	m := &Matcher{opts: opts}
	// Normalise the query once so matching only has to normalise the document.
	for _, s := range steps {
		st := step{hasText: s.HasText}
		if s.Name != "" {
			st.name = []byte(m.normaliseCase(m.omitNamespace(s.Name)))
		}
		if s.HasText {
			st.text = []byte(m.normaliseCase(s.Text))
		}
		for _, a := range s.Attrs {
			st.attrs = append(st.attrs, attr{
				key:      []byte(m.normaliseCase(a.Key)),
				value:    []byte(m.normaliseCase(a.Value)),
				hasValue: a.HasValue,
			})
		}
		m.steps = append(m.steps, st)
	}
	return m, nil
}

func (m *Matcher) normaliseCase(s string) string {
	if m.opts.CaseInsensitive {
		return strings.ToLower(s)
	}
	return s
}

func (m *Matcher) omitNamespace(name string) string {
	if m.opts.OmitNamespace {
		if _, local, found := strings.Cut(name, ":"); found {
			return local
		}
	}
	return name
}

// Steps returns the number of steps in the query.
func (m *Matcher) Steps() int {
	return len(m.steps)
}

// Last is the bit of the last step. An element whose Match result has it
// set matches the whole query.
func (m *Matcher) Last() uint64 {
	return 1 << (len(m.steps) - 1)
}

// HasText reports whether the query matches on element text, which is
// only known once the element has ended.
func (m *Matcher) HasText() bool {
	return m.steps[len(m.steps)-1].hasText
}

// Next returns the steps an element may satisfy given the result of
// Match for its parent: the first step, and every step that follows one
// the parent satisfied.
func (m *Matcher) Next(parent uint64) uint64 {
	next := parent<<1 | 1
	if len(m.steps) < 64 {
		next &= 1<<len(m.steps) - 1
	}
	return next
}

// Match returns the subset of the steps in candidates that a start tag
// satisfies, as a bit set. name is the tag name and rest the part of the
// tag after it. Attributes are parsed only when a step's name matches.
func (m *Matcher) Match(candidates uint64, name, rest []byte) (uint64, error) {
	var matched uint64
	parsed := false
	for c := candidates; c != 0; c &= c - 1 {
		i := bits.TrailingZeros64(c)
		s := &m.steps[i]
		if s.name != nil && !m.nameMatches(s.name, name) {
			continue
		}
		if !parsed {
			var err error
			if m.attrs, err = markup.ParseAttributes(rest, m.attrs[:0]); err != nil {
				return 0, err
			}
			parsed = true
		}
		if m.attributesMatch(s) {
			matched |= 1 << i
		}
	}
	return matched, nil
}

func (m *Matcher) nameMatches(query, name []byte) bool {
	if m.opts.OmitNamespace {
		if i := bytes.IndexByte(name, ':'); i >= 0 {
			name = name[i+1:]
		}
	}
	return m.equal(query, name)
}

// equal compares a normalised query value with a document value.
func (m *Matcher) equal(query, value []byte) bool {
	switch {
	case !m.opts.CaseInsensitive && !m.opts.Contains:
		return bytes.Equal(query, value)
	case !m.opts.Contains:
		return bytes.EqualFold(query, value)
	case !m.opts.CaseInsensitive:
		return bytes.Contains(value, query)
	}
	m.lower = appendLower(m.lower[:0], value)
	return bytes.Contains(m.lower, query)
}

// attributesMatch reports whether every attribute of s is matched by an
// attribute of the current tag. Values are compared after unescaping.
func (m *Matcher) attributesMatch(s *step) bool {
	for _, want := range s.attrs {
		found := false
		for _, have := range m.attrs {
			if !m.equal(want.key, have.Key) {
				continue
			}
			if !want.hasValue {
				found = true
				break
			}
			value := have.Value
			if bytes.IndexByte(value, '&') >= 0 {
				m.unesc = markup.AppendUnescaped(m.unesc[:0], value)
				value = m.unesc
			}
			if m.equal(want.value, value) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// TextMatches reports whether inner, the xml between the start and end
// tag of an element, has the text the last step asks for.
func (m *Matcher) TextMatches(inner []byte) bool {
	s := &m.steps[len(m.steps)-1]
	if !s.hasText {
		return true
	}
	m.textBuf = markup.AppendText(m.textBuf[:0], inner)
	return m.equal(s.text, m.textBuf)
}

// appendLower appends s in lower case to dst.
func appendLower(dst, s []byte) []byte {
	for i, c := range s {
		if c >= utf8.RuneSelf {
			return append(dst, bytes.ToLower(s[i:])...)
		}
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		dst = append(dst, c)
	}
	return dst
}
