package tagmatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcw/saxer/markup"
)

// matches reports whether the innermost of the nested tags, given as the
// inside of their start tags, matches query.
func matches(t *testing.T, query string, opts Options, tags ...string) bool {
	t.Helper()
	m, err := New(query, opts)
	require.NoError(t, err)
	var state uint64
	for _, tag := range tags {
		name, rest := markup.SplitName([]byte(tag))
		state, err = m.Match(m.Next(state), name, rest)
		require.NoError(t, err)
	}
	return state&m.Last() != 0
}

func TestMatch(t *testing.T) {
	ci := Options{CaseInsensitive: true}
	contains := Options{Contains: true}
	omit := Options{OmitNamespace: true}
	tests := []struct {
		name  string
		query string
		opts  Options
		tags  []string
		want  bool
	}{
		{"name", "mediawiki", Options{}, []string{"mediawiki"}, true},
		{"name in path", "mediawiki", Options{}, []string{"hello", "mediawiki"}, true},
		{"attribute", "text?xml:space=preserve", Options{}, []string{"hello", `text xml:space="preserve"`}, true},
		{"attribute only", "?xml:space=preserve", Options{}, []string{"hello", `text xml:space="preserve"`}, true},
		{"attributes with spaces", "mediawiki?attrvalue=one two&attrvalue2=1 2 3 4", Options{},
			[]string{"mediawiki   attrvalue     =   \"one two\"  attrvalue2     =   '1 2 3 4'   "}, true},
		{"attributes on new lines", "mediawiki?id=1&lang=en", Options{}, []string{"mediawiki\n\tid=\"1\"\r\n\tlang=\"en\""}, true},
		{"case insensitive", "mediawiki", ci, []string{"mediaWiki"}, true},
		{"case insensitive key", "?id", ci, []string{`mediaWiki Id="1234"`}, true},
		{"case insensitive value", "?id=test", ci, []string{`mediaWiki Id="Test"`}, true},
		{"case sensitive value", "?id=test", Options{}, []string{`mediaWiki id="Test"`}, false},
		{"case sensitive name", "mediawiki", Options{}, []string{"mediaWiki"}, false},
		{"case insensitive unicode", "ÆRØ", ci, []string{"ærø"}, true},
		{"contains", "medi", contains, []string{"mediaWiki"}, true},
		{"contains case insensitive", "WIKI", Options{Contains: true, CaseInsensitive: true}, []string{"mediaWiki"}, true},
		{"equals is not contains", "medi", Options{}, []string{"mediaWiki"}, false},
		{"contains attribute key", "?ref", contains, []string{`mediaWiki referance="12345"`}, true},
		{"omit namespace", "doors", omit, []string{"xs:doors"}, true},
		{"omit namespace in query", "xs:doors", omit, []string{"ys:doors"}, true},
		{"namespace by default", "doors", Options{}, []string{"xs:doors"}, false},
		{"every query attribute must match", "?ab&zz", contains, []string{`a ab="1" abc="2"`}, false},
		{"query attribute matching several", "?ab", contains, []string{`a ab="1" abc="2"`}, true},
		{"path", "a/b", Options{}, []string{"x", "a", "b"}, true},
		{"path with gap", "a/b", Options{}, []string{"a", "x", "b"}, false},
		{"path attribute on parent", "a?id=1/b", Options{}, []string{`a id="1"`, "b"}, true},
		{"path attribute on parent no match", "a?id=1/b", Options{}, []string{`a id="2"`, "b"}, false},
		{"wildcard", "car/*/Fuel", Options{}, []string{"cars", "car", "engine", "Fuel"}, true},
		{"wildcard needs a level", "car/*/Fuel", Options{}, []string{"car", "Fuel"}, false},
		{"wildcard alone", "*", Options{}, []string{"anything"}, true},
		{"repeated names", "a/a/b", Options{}, []string{"a", "a", "a", "b"}, true},
		{"unescaped value", `?name=a\&b`, Options{}, []string{`x name="a&amp;b"`}, true},
		{"unescaped value lt", "?v=<1>", Options{}, []string{`x v="&lt;1&gt;"`}, true},
		{"raw value does not match", "?name=a&amp;b", Options{}, []string{`x name="a&amp;b"`}, false},
		{"empty value", "?e=", Options{}, []string{`x e=""`}, true},
		{"empty value must be empty", "?e=", Options{}, []string{`x e="1"`}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, matches(t, tt.query, tt.opts, tt.tags...))
		})
	}
}

func TestMatchMalformedAttributes(t *testing.T) {
	m, err := New("a", Options{})
	require.NoError(t, err)
	_, err = m.Match(m.Next(0), []byte("a"), []byte(" x=1"))
	assert.Error(t, err)
	// Attributes of tags whose name cannot match are not parsed.
	_, err = m.Match(m.Next(0), []byte("b"), []byte(" x=1"))
	assert.NoError(t, err)
}

func TestTextMatches(t *testing.T) {
	m, err := New("color~blue", Options{})
	require.NoError(t, err)
	assert.True(t, m.HasText())
	assert.True(t, m.TextMatches([]byte(" blue ")))
	assert.True(t, m.TextMatches([]byte("<![CDATA[blue]]>")))
	assert.False(t, m.TextMatches([]byte("Blue")))
	assert.False(t, m.TextMatches([]byte("dark blue")))

	m, err = New("color~BLUE", Options{CaseInsensitive: true, Contains: true})
	require.NoError(t, err)
	assert.True(t, m.TextMatches([]byte("dark <b>Blue</b>")))

	m, err = New("color", Options{})
	require.NoError(t, err)
	assert.False(t, m.HasText())
	assert.True(t, m.TextMatches([]byte("anything")))
}

func TestNewRejectsInvalidQuery(t *testing.T) {
	for _, q := range []string{"a?b?c", "", "a~x/b"} {
		_, err := New(q, Options{})
		assert.Error(t, err, q)
	}
}

func TestNextLimitsToQuery(t *testing.T) {
	m, err := New("a/b", Options{})
	require.NoError(t, err)
	assert.Equal(t, uint64(0b01), m.Next(0))
	assert.Equal(t, uint64(0b11), m.Next(0b01))
	assert.Equal(t, uint64(0b01), m.Next(0b10))
}
