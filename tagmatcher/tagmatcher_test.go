package tagmatcher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcw/saxer/tagpath"
)

func TestAddTagWithAttributeWithSpace(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki")
	tm.AddTag("mediawiki   attrvalue     =   \"one two\"  attrvalue2     =   '1 2 3 4'   ")

	assert.Equal(t, tm.path.Path[0].Name, "mediawiki")
	assert.Equal(t, tm.path.Path[0].Attributes[0].Key, "attrvalue")
	assert.Equal(t, tm.path.Path[0].Attributes[0].Value, "one two")
	assert.Equal(t, tm.path.Path[0].Attributes[1].Key, "attrvalue2")
	assert.Equal(t, tm.path.Path[0].Attributes[1].Value, "1 2 3 4")
}

func TestAddTagWithOnlyTagName(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki")
	tm.AddTag("mediawiki")

	assert.Equal(t, tm.path.Path[0].Name, "mediawiki")
}

func TestAddTagWithOnlyTagNameMatches(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki")
	tm.AddTag("mediawiki")
	assert.True(t, tm.MatchesPath())
}

func TestAddTagWithOnlyTwoTagNameMatches(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki")
	tm.AddTag("hello")
	tm.AddTag("mediawiki")
	assert.True(t, tm.MatchesPath())
}

func TestAddTagWithSecondTagNameAttributeMatches(t *testing.T) {
	tm := newTagMatcher(t, "text?xml:space=preserve")
	tm.AddTag("hello")
	tm.AddTag("text xml:space=\"preserve\"")
	assert.True(t, tm.MatchesPath())
}

func TestAddTagWithSecondTagNameAttributeMatchesOnlyAttributeQuery(t *testing.T) {
	tm := newTagMatcher(t, "?xml:space=preserve")
	tm.AddTag("hello")
	tm.AddTag("text xml:space=\"preserve\"")
	assert.True(t, tm.MatchesPath())
}

func TestMatchCaseInsensitive(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki")
	tm.CaseSensitive = false
	tm.AddTag("mediaWiki")
	assert.True(t, tm.MatchesPath())
}

func TestMatchCaseInsensitiveAttributeKey(t *testing.T) {
	tm := newTagMatcher(t, "?id")
	tm.CaseSensitive = false
	tm.AddTag("mediaWiki Id=\"1234\"")
	assert.True(t, tm.MatchesPath())
}

func TestMatchCaseInsensitiveAttributeValue(t *testing.T) {
	tm := newTagMatcher(t, "?id=test")
	tm.CaseSensitive = false
	tm.AddTag("mediaWiki Id=\"Test\"")
	assert.True(t, tm.MatchesPath())
}

func TestMatchCaseSensitiveAttributeValue(t *testing.T) {
	tm := newTagMatcher(t, "?id=test")
	tm.CaseSensitive = true
	tm.AddTag("mediaWiki id=\"Test\"")
	assert.False(t, tm.MatchesPath())
}

func TestNotMatchCaseSensitive(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki")
	tm.CaseSensitive = true
	tm.AddTag("mediaWiki")
	assert.False(t, tm.MatchesPath())
}

func TestMatchContain(t *testing.T) {
	tm := newTagMatcher(t, "medi")
	tm.EqualityFn = EqFnContains
	tm.AddTag("mediaWiki")
	assert.True(t, tm.MatchesPath())
}

func TestNotMatchEquals(t *testing.T) {
	tm := newTagMatcher(t, "medi")
	tm.EqualityFn = EqFnEquals
	tm.AddTag("mediaWiki")
	assert.False(t, tm.MatchesPath())
}

func TestMatchContainAttributeValue(t *testing.T) {
	tm := newTagMatcher(t, "?ref")
	tm.EqualityFn = EqFnContains
	tm.AddTag("mediaWiki referance=\"12345\"")
	assert.True(t, tm.MatchesPath())
}

func TestAddTagWithAttributeOnNewLine(t *testing.T) {
	tm := newTagMatcher(t, "mediawiki?id=1&lang=en")
	assert.NoError(t, tm.AddTag("mediawiki\n\tid=\"1\"\r\n\tlang=\"en\""))
	assert.True(t, tm.MatchesPath())
}

func TestAddTagMalformedAttributes(t *testing.T) {
	tm := newTagMatcher(t, "a")
	assert.Error(t, tm.AddTag("a x=1"))
}

func TestAddTagTooDeep(t *testing.T) {
	tm := newTagMatcher(t, "a")
	for i := 0; i < tagpath.MaxDepth; i++ {
		assert.NoError(t, tm.AddTag("a"))
	}
	assert.Error(t, tm.AddTag("a"))
}

func TestAddTagTooManyAttributes(t *testing.T) {
	tm := newTagMatcher(t, "a")
	tag := "a"
	for i := 0; i <= tagpath.MaxAttributes; i++ {
		tag += fmt.Sprintf(" a%d=\"%d\"", i, i)
	}
	assert.Error(t, tm.AddTag(tag))
}

func newTagMatcher(t *testing.T, query string) TagMatcher {
	t.Helper()
	tm, err := NewTagMatcher(query)
	require.NoError(t, err)
	return tm
}

func TestNewTagMatcherRejectsInvalidQuery(t *testing.T) {
	_, err := NewTagMatcher("a?b?c")
	assert.Error(t, err)
}
