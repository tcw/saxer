package queryparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parse(t *testing.T, query string) []Step {
	t.Helper()
	steps, err := Parse(query)
	require.NoError(t, err)
	return steps
}

func TestParse(t *testing.T) {
	tests := []struct {
		query string
		want  []Step
	}{
		{"username", []Step{{Name: "username"}}},
		{"?id", []Step{{Attrs: []Attr{{Key: "id"}}}}},
		{"?id=123", []Step{{Attrs: []Attr{{Key: "id", Value: "123", HasValue: true}}}}},
		{"?id=123&ref=456", []Step{{Attrs: []Attr{
			{Key: "id", Value: "123", HasValue: true}, {Key: "ref", Value: "456", HasValue: true}}}}},
		{"username?id", []Step{{Name: "username", Attrs: []Attr{{Key: "id"}}}}},
		{"username?id=123", []Step{{Name: "username", Attrs: []Attr{{Key: "id", Value: "123", HasValue: true}}}}},
		{"username/system", []Step{{Name: "username"}, {Name: "system"}}},
		{"username/?id", []Step{{Name: "username"}, {Attrs: []Attr{{Key: "id"}}}}},
		{"username/system?id=123", []Step{{Name: "username"}, {Name: "system", Attrs: []Attr{{Key: "id", Value: "123", HasValue: true}}}}},
		{"username?", []Step{{Name: "username"}}},
		{"?id=", []Step{{Attrs: []Attr{{Key: "id", HasValue: true}}}}},
		{"/a//b/", []Step{{Name: "a"}, {Name: "b"}}},
		{" a / b ", []Step{{Name: "a"}, {Name: "b"}}},
		// Wildcards.
		{"car/*/Fuel", []Step{{Name: "car"}, {}, {Name: "Fuel"}}},
		{"*", []Step{{}}},
		{"*?nr=1", []Step{{Attrs: []Attr{{Key: "nr", Value: "1", HasValue: true}}}}},
		// Text.
		{"color~blue", []Step{{Name: "color", Text: "blue", HasText: true}}},
		{"car/color~dark blue?lang=en", []Step{{Name: "car"}, {Name: "color", Text: "dark blue", HasText: true,
			Attrs: []Attr{{Key: "lang", Value: "en", HasValue: true}}}}},
		{"*~x", []Step{{Text: "x", HasText: true}}},
		{"b~", []Step{{Name: "b", HasText: true}}},
		// Escapes.
		{`?name=a\&b`, []Step{{Attrs: []Attr{{Key: "name", Value: "a&b", HasValue: true}}}}},
		{`?href=http:\/\/x\/y\?q\=1`, []Step{{Attrs: []Attr{{Key: "href", Value: "http://x/y?q=1", HasValue: true}}}}},
		{`a~1\~2\\`, []Step{{Name: "a", Text: `1~2\`, HasText: true}}},
		{`\*`, []Step{{Name: "*"}}},
		{`a?b=x\=y`, []Step{{Name: "a", Attrs: []Attr{{Key: "b", Value: "x=y", HasValue: true}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			assert.Equal(t, tt.want, parse(t, tt.query))
		})
	}
}

func TestParseManyLevelsAndAttributes(t *testing.T) {
	assert.Len(t, parse(t, strings.Repeat("a/", MaxSteps)), MaxSteps)
	assert.Len(t, parse(t, "a?"+strings.Repeat("x&", 500)+"x")[0].Attrs, 501)
}

func TestEmptyQuery(t *testing.T) {
	for _, query := range []string{"", " ", "/", "//", " / "} {
		_, err := Parse(query)
		assert.ErrorIs(t, err, ErrEmpty, query)
	}
}

func TestInvalidQueries(t *testing.T) {
	for _, query := range []string{
		"a?b?c",
		"a?id=1=2",
		"a?=1",
		"a?id&&ref",
		"a?id&",
		"a~x~y",
		"a~x/b",
		`a\`,
		strings.Repeat("a/", MaxSteps+1),
	} {
		_, err := Parse(query)
		assert.Error(t, err, query)
	}
}
