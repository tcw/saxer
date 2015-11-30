package queryParser
import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLevel1PathQuery(t *testing.T) {
	path := Parse("username")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,0,path.Path[0].AttributePos)
}

func TestLevel1AttrKeyQuery(t *testing.T) {
	path := Parse("?id")
	assert.Equal(t,"",path.Path[0].Name)
	assert.Equal(t,"id",path.Path[0].Attributes[0].Key)
}

func TestLevel1AttrKeyValQuery(t *testing.T) {
	path := Parse("?id=123")
	assert.Equal(t,"",path.Path[0].Name)
	assert.Equal(t,"id",path.Path[0].Attributes[0].Key)
	assert.Equal(t,"123",path.Path[0].Attributes[0].Value)
}

func TestLevel1TwoAttrKeyValPairsQuery(t *testing.T) {
	path := Parse("?id=123&ref=456")
	assert.Equal(t,"",path.Path[0].Name)
	assert.Equal(t,"id",path.Path[0].Attributes[0].Key)
	assert.Equal(t,"123",path.Path[0].Attributes[0].Value)
	assert.Equal(t,"ref",path.Path[0].Attributes[1].Key)
	assert.Equal(t,"456",path.Path[0].Attributes[1].Value)
}

func TestLevel1PathAttrKeyQuery(t *testing.T) {
	path := Parse("username?id")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,"id",path.Path[0].Attributes[0].Key)
}

func TestLevel1PathAttrKeyValueQuery(t *testing.T) {
	path := Parse("username?id=123")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,"id",path.Path[0].Attributes[0].Key)
	assert.Equal(t,"123",path.Path[0].Attributes[0].Value)
}

func TestLevel2PathQuery(t *testing.T) {
	path := Parse("username/system")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,"system",path.Path[1].Name)
}

func TestLevel2PathOnlyAttrKeyQuery(t *testing.T) {
	path := Parse("username/?id")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,"",path.Path[1].Name)
	assert.Equal(t,"id",path.Path[1].Attributes[0].Key)
}

func TestLevel2PathAndAttrKeyQuery(t *testing.T) {
	path := Parse("username/system?id")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,"system",path.Path[1].Name)
	assert.Equal(t,"id",path.Path[1].Attributes[0].Key)
}

func TestLevel2PathAndAttrKeyValueQuery(t *testing.T) {
	path := Parse("username/system?id=123")
	assert.Equal(t,"username",path.Path[0].Name)
	assert.Equal(t,"system",path.Path[1].Name)
	assert.Equal(t,"id",path.Path[1].Attributes[0].Key)
	assert.Equal(t,"123",path.Path[1].Attributes[0].Value)
}