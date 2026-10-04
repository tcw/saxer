package queryparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcw/saxer/tagpath"
)

func parse(t *testing.T, query string) *tagpath.TagPath {
	t.Helper()
	path, err := Parse(query)
	require.NoError(t, err)
	return path
}

func TestLevel1PathQuery(t *testing.T) {
	path := parse(t, "username")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, 0, path.Path[0].AttributePos)
}

func TestLevel1AttrKeyQuery(t *testing.T) {
	path := parse(t, "?id")
	assert.Equal(t, "", path.Path[0].Name)
	assert.Equal(t, "id", path.Path[0].Attributes[0].Key)
}

func TestLevel1AttrKeyValQuery(t *testing.T) {
	path := parse(t, "?id=123")
	assert.Equal(t, "", path.Path[0].Name)
	assert.Equal(t, "id", path.Path[0].Attributes[0].Key)
	assert.Equal(t, "123", path.Path[0].Attributes[0].Value)
}

func TestLevel1TwoAttrKeyValPairsQuery(t *testing.T) {
	path := parse(t, "?id=123&ref=456")
	assert.Equal(t, "", path.Path[0].Name)
	assert.Equal(t, "id", path.Path[0].Attributes[0].Key)
	assert.Equal(t, "123", path.Path[0].Attributes[0].Value)
	assert.Equal(t, "ref", path.Path[0].Attributes[1].Key)
	assert.Equal(t, "456", path.Path[0].Attributes[1].Value)
}

func TestLevel1PathAttrKeyQuery(t *testing.T) {
	path := parse(t, "username?id")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, "id", path.Path[0].Attributes[0].Key)
}

func TestLevel1PathAttrKeyValueQuery(t *testing.T) {
	path := parse(t, "username?id=123")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, "id", path.Path[0].Attributes[0].Key)
	assert.Equal(t, "123", path.Path[0].Attributes[0].Value)
}

func TestLevel2PathQuery(t *testing.T) {
	path := parse(t, "username/system")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, "system", path.Path[1].Name)
}

func TestLevel2PathOnlyAttrKeyQuery(t *testing.T) {
	path := parse(t, "username/?id")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, "", path.Path[1].Name)
	assert.Equal(t, "id", path.Path[1].Attributes[0].Key)
}

func TestLevel2PathAndAttrKeyQuery(t *testing.T) {
	path := parse(t, "username/system?id")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, "system", path.Path[1].Name)
	assert.Equal(t, "id", path.Path[1].Attributes[0].Key)
}

func TestLevel2PathAndAttrKeyValueQuery(t *testing.T) {
	path := parse(t, "username/system?id=123")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, "system", path.Path[1].Name)
	assert.Equal(t, "id", path.Path[1].Attributes[0].Key)
	assert.Equal(t, "123", path.Path[1].Attributes[0].Value)
}

func TestEmptyAttributeListIsIgnored(t *testing.T) {
	path := parse(t, "username?")
	assert.Equal(t, "username", path.Path[0].Name)
	assert.Equal(t, 0, path.Path[0].AttributePos)
}

func TestAttributeValueMayBeEmpty(t *testing.T) {
	path := parse(t, "?id=")
	assert.Equal(t, "id", path.Path[0].Attributes[0].Key)
	assert.Equal(t, "", path.Path[0].Attributes[0].Value)
}

func TestInvalidQueries(t *testing.T) {
	for _, query := range []string{
		"a?b?c",
		"a?id=1=2",
		"a?=1",
		"a?id&&ref",
		strings.Repeat("a/", tagpath.MaxDepth+1),
		"a?" + strings.Repeat("x&", tagpath.MaxAttributes+1),
	} {
		_, err := Parse(query)
		assert.Error(t, err, query)
	}
}
