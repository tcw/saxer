package tagMatcher
import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestAddTagWithAttributeWithSpace(t *testing.T) {
	tm :=NewTagMatcher("mediawiki")
	tm.AddTag("mediawiki   attrvalue     =   \"one two\"  attrvalue2     =   '1 2 3 4'   ")

	assert.Equal(t,tm.path.Path[0].Name,"mediawiki")
	assert.Equal(t,tm.path.Path[0].Attributes[0].Key,"attrvalue")
	assert.Equal(t,tm.path.Path[0].Attributes[0].Value,"one two")
	assert.Equal(t,tm.path.Path[0].Attributes[1].Key,"attrvalue2")
	assert.Equal(t,tm.path.Path[0].Attributes[1].Value,"1 2 3 4")
}

func TestAddTagWithOnlyTagName(t *testing.T) {
	tm :=NewTagMatcher("mediawiki")
	tm.AddTag("mediawiki")

	assert.Equal(t,tm.path.Path[0].Name,"mediawiki")
}

func TestAddTagWithOnlyTagNameMatches(t *testing.T) {
	tm :=NewTagMatcher("mediawiki")
	tm.AddTag("mediawiki")
	assert.True(t,tm.MatchesPath())
}

func TestAddTagWithOnlyTwoTagNameMatches(t *testing.T) {
	tm :=NewTagMatcher("mediawiki")
	tm.AddTag("hello")
	tm.AddTag("mediawiki")
	assert.True(t,tm.MatchesPath())
}

func TestAddTagWithSecondTagNameAttributeMatches(t *testing.T) {
	tm :=NewTagMatcher("text?xml:space=preserve")
	tm.AddTag("hello")
	tm.AddTag("text xml:space=\"preserve\"")
	assert.True(t,tm.MatchesPath())
}

func TestAddTagWithSecondTagNameAttributeMatchesOnlyAttributeQuery(t *testing.T) {
	tm :=NewTagMatcher("?xml:space=preserve")
	tm.AddTag("hello")
	tm.AddTag("text xml:space=\"preserve\"")
	assert.True(t,tm.MatchesPath())
}

func TestAddTagWithSecondTagNameAttributeMatchesOnlyAttributeKeyQuery(t *testing.T) {
	tm :=NewTagMatcher("?xml:space")
	tm.AddTag("hello")
	tm.AddTag("text xml:space=\"preserve\"")
	assert.True(t,tm.MatchesPath())
}