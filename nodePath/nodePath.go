package nodePath
import (
	"strings"
)

type NodePath struct {
	pathQuery []string
	path      []string
	pos       int
}

func NewNodePath(size int, query string) NodePath {
	i := make([]string, size)
	split := strings.Split(query, "/")
	return NodePath{pathQuery:split, path: i, pos:0}
}

func (np *NodePath)GetPath() string{
	return strings.Join(np.path[:np.pos],"/")
}

func (np *NodePath)Add(s string) {
	np.path[np.pos] = s
	np.pos++
}

func (np *NodePath)RemoveLast() {
	np.pos = np.pos - 1
}

func (np *NodePath) MatchesPath() bool {
	if np.pos == len(np.pathQuery){
		for i := 0; i < np.pos; i++ {
			if np.pathQuery[i] != np.path[i]{
				return false
			}
		}
	}else{
		return false
	}
	return true
}



