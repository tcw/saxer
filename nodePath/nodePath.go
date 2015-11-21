package nodePath
import (
	"strings"
)

type NodePath struct {
	pathQuery     []string
	path          []string
	lastMatchPath []string
	pos           int
	lastMatchPos  int
}

func NewNodePath(size int, query string) NodePath {
	path := make([]string, size)
	last := make([]string, size)
	split := strings.Split(query, "/")
	return NodePath{pathQuery:split, path: path, lastMatchPath:last, pos:0, lastMatchPos:0}
}

func (np *NodePath)GetPath() string {
	return strings.Join(np.path[:np.pos], "/")
}

func (np *NodePath)GetLastMatchPath() string {
	return strings.Join(np.lastMatchPath[:np.lastMatchPos], "/")
}

func (np *NodePath)Add(s string) {
	np.path[np.pos] = s
	np.pos++
}

func (np *NodePath)RemoveLast() {
	np.pos = np.pos - 1
}

func (np *NodePath) MatchesLastMatch() bool {
	if np.lastMatchPos == np.pos {
		for i := 0; i < np.lastMatchPos; i++ {
			if np.lastMatchPath[i] != np.path[i] {
				return false
			}
		}
	}else {
		return false
	}
	return true
}

func (np *NodePath) MatchesPath() bool {
	pathQueryLength := len(np.pathQuery)
	delta := np.pos - pathQueryLength
	if np.pos >= pathQueryLength {
		for i := pathQueryLength - 1; i >= 0; i-- {
			if np.pathQuery[i] != np.path[i + delta] {
				return false
			}
		}
		np.lastMatchPos = np.pos
		copy(np.lastMatchPath,np.path[:np.pos])
		return true
	}else {
		return false
	}
}


