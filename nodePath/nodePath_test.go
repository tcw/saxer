package nodePath_test
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

func TestOneNodeNoMatch(t *testing.T) {
	np := nodePath.NewNodePath(10,"a")
	np.Add("b")
	assert.False(t,np.MatchesPath())
	assert.Equal(t,np.GetLastMatchPath(),"")
}

func TestTwoNodeAbsoluteMatch(t *testing.T) {
	np := nodePath.NewNodePath(10,"a/b")
	np.Add("a")
	np.Add("b")
	assert.True(t,np.MatchesPath())
	assert.Equal(t,np.GetLastMatchPath(),"a/b")
}

func TestTwoNodeRelativeMatch(t *testing.T) {
	np := nodePath.NewNodePath(10,"a/b")
	np.Add("c")
	np.Add("a")
	np.Add("b")
	assert.True(t,np.MatchesPath())
	assert.Equal(t,np.GetLastMatchPath(),"c/a/b")
}

func TestTwoNodeNoRelativeMatch(t *testing.T) {
	np := nodePath.NewNodePath(10,"a/b")
	np.Add("c")
	np.Add("d")
	np.Add("b")
	assert.False(t,np.MatchesPath())
	assert.Equal(t,np.GetLastMatchPath(),"")
}

func TestTwoNodeRelativeMatchLast(t *testing.T) {
	np := nodePath.NewNodePath(10,"b/c")
	np.Add("a")
	np.Add("b")
	np.Add("c")
	assert.True(t,np.MatchesPath())
	assert.True(t,np.MatchesLastMatch())
	np.Add("d")
	assert.False(t,np.MatchesPath())
	assert.False(t,np.MatchesLastMatch())
	np.RemoveLast()
	assert.True(t,np.MatchesLastMatch())
}

func TestTwoNodeAbsolutMatchLast(t *testing.T) {
	np := nodePath.NewNodePath(10,"a/b/c")
	np.Add("a")
	np.Add("b")
	np.Add("c")
	assert.True(t,np.MatchesPath())
	assert.True(t,np.MatchesLastMatch())
	np.RemoveLast()
	np.Add("d")
	assert.False(t,np.MatchesPath())
	assert.False(t,np.MatchesLastMatch())
	np.RemoveLast()
	np.Add("c")
	assert.True(t,np.MatchesPath())
	assert.True(t,np.MatchesLastMatch())
}