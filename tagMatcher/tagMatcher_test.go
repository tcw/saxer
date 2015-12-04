package tagMatcher
import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestAddTagWithAttributeWithSpace(t *testing.T) {
	tm :=NewTagMatcher(10,"mediawiki")
	tm.AddTag("mediawiki   attrvalue     =   \"one two\"  attrvalue2     =   '1 2 3 4'   ")

	assert.Equal(t,tm.path.Path[0].Name,"mediawiki")
	assert.Equal(t,tm.path.Path[0].Attributes[0].Key,"attrvalue")
	assert.Equal(t,tm.path.Path[0].Attributes[0].Value,"one two")
	assert.Equal(t,tm.path.Path[0].Attributes[1].Key,"attrvalue2")
	assert.Equal(t,tm.path.Path[0].Attributes[1].Value,"1 2 3 4")
}

func TestAddTagWithOnlyTagName(t *testing.T) {
	tm :=NewTagMatcher(10,"mediawiki")
	tm.AddTag("mediawiki")

	assert.Equal(t,tm.path.Path[0].Name,"mediawiki")
}
