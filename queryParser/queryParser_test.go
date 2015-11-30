package queryParser
import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestLevel1PathQuery(t *testing.T) {
	path := Parse("username")
	assert.Equal(t,path.Path[0].Name,"username")
	assert.Equal(t,path.Path[0].AttributePos,0)
}
