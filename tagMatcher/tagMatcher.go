package tagMatcher
import (
	"strings"
	"github.com/tcw/saxer/tagPath"
	"github.com/tcw/saxer/queryParser"
	"fmt"
)

type TagMatcher struct {
	query         tagPath.TagPath
	path          tagPath.TagPath
	lastMatchPath []string
	lastMatchPos  int
}

func NewTagMatcher(size int, queryString string) TagMatcher {
	path := tagPath.NewTagPath()
	last := make([]string, size)
	q := queryParser.Parse(queryString)
	return TagMatcher{query:q, path: path, lastMatchPath:last, lastMatchPos:0}
}


func (tm *TagMatcher)AddTag(tagText string) {
	tag :=tagPath.NewTag()
	elem := strings.Split(tagText," ")
	tag.Name = elem[0]
	if len(elem) > 1 {
		for i := 1; i < len(elem);i++  {
			attr := strings.Split(elem[i],"=")
			tag.Add(tagPath.Attribute{Key:attr[0],Value:strings.Trim(attr[1],"\"'")})
		}
	}
	tm.path.Add(tag)
}

func (np *TagMatcher)RemoveLast() {
	np.path.RemoveLast()
}

func (tm *TagMatcher) TagNameMatchesLastMatch() bool {
	if tm.lastMatchPos == tm.path.PathPos {
		for i := 0; i < tm.lastMatchPos; i++ {
			if tm.lastMatchPath[i] != tm.path.Path[i].Name {
				return false
			}
		}
	}else {
		return false
	}
	return true
}

func (tm *TagMatcher) MatchesPath() bool {
	pathQueryLength := tm.query.PathPos
	delta := tm.path.PathPos - pathQueryLength
	var actualMatches int = 0
	var expectedMatches int = 0
	if tm.path.PathPos >= pathQueryLength {
		for i := pathQueryLength - 1; i >= 0; i-- {
			fmt.Println(tm.query.Path[i].Name,tm.path.Path[i].Name)
			if len(tm.query.Path[i].Name) != 0 && tm.query.Path[i].Name != tm.path.Path[i + delta].Name {
				return false
			}
			queryAttr := tm.query.Path[i].Attributes
			pathAttr := tm.path.Path[i].Attributes
			expectedMatches = tm.query.Path[i].AttributePos
			fmt.Println(tm.query.Path[i].Attributes[:tm.query.Path[i].AttributePos],tm.path.Path[i].Attributes[:tm.path.Path[i].AttributePos])
			for j := 0; j > tm.query.Path[i].AttributePos; j++ {
				for g := 0; g < tm.path.Path[j].AttributePos; g++ {
					if queryAttr[j].Key == pathAttr[g].Key{
						if len(queryAttr[j].Value) == 0 || queryAttr[j].Value == pathAttr[g].Value{
							actualMatches++
						}
					}
				}
			}
			fmt.Println(expectedMatches,actualMatches)
			if expectedMatches != actualMatches{
				return false
			}
			actualMatches = 0
			expectedMatches = 0
		}
		tm.lastMatchPos = tm.path.PathPos
		for i := 0; i < tm.path.PathPos; i++ {
			tm.lastMatchPath[i] = tm.path.Path[i].Name
		}
		return true
	}else {
		return false
	}
}


