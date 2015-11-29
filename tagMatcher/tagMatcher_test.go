package tagMatcher
import (
	"testing"
	"github.com/tcw/saxer/nodePath"
	"github.com/zacg/testify/assert"
)

func TestOneNodeMatch(t *testing.T) {
	np := nodePath.NewNodePath(100,"a")
	np.Add("a")
	assert.True(t,np.MatchesPath())
	assert.Equal(t,np.GetLastMatchPath(),"a")
}
