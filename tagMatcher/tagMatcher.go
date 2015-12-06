package tagMatcher
import (
	"strings"
	"github.com/tcw/saxer/tagPath"
	"github.com/tcw/saxer/queryParser"
	"github.com/google/gxui/math"
)

type TagMatcher struct {
	query              tagPath.TagPath
	queryHasAttributes bool
	queryHasPath       bool
	path               tagPath.TagPath
	lastMatchPath      []string
	lastMatchPos       int
}

func NewTagMatcher(size int, queryString string) TagMatcher {
	path := tagPath.NewTagPath()
	last := make([]string, size)
	q := queryParser.Parse(queryString)
	qHasAttributes := false
	qHasPath := false
	for i := 0; i < q.PathPos; i++ {
		if q.Path[i].AttributePos > 0 {
			qHasAttributes = true
		}
		if len(q.Path[i].Name) > 0 {
			qHasPath = true
		}
	}
	return TagMatcher{query:*q, queryHasAttributes:qHasAttributes, queryHasPath:qHasPath, path: *path, lastMatchPath:last, lastMatchPos:0}
}


func (tm *TagMatcher)AddTag(tagText string) {
	tagNameEnd := 0
	attr := make([]int, 100)
	attrPos := 0
	insideAttrValue := false
	insideAttrKey := false
	readSpace := false
	readEquals := false
	var seperator rune = 0
	noEnd := strings.TrimRight(tagText,"/")
	trimmed := strings.TrimSpace(noEnd)
	for key, value := range trimmed {
		if value == rune(' ') && !insideAttrValue {
			if tagNameEnd == 0 {
				tagNameEnd = key
			}
			readSpace = true
		}else if value == rune('=') && !insideAttrValue {
			readEquals = true
			attr[attrPos] = key
			attrPos++
		}else if readEquals && !insideAttrValue && (value == rune('\'') || value == rune('"')) {
			seperator = value
			insideAttrValue = true
			insideAttrKey = false
			readEquals = false
			attr[attrPos] = key + 1
			attrPos++
		}else if readSpace && !insideAttrValue {
			if !insideAttrKey {
				attr[attrPos] = key
				attrPos++
			}
			insideAttrKey = true
		}else if value == seperator && insideAttrValue {
			seperator = 0
			insideAttrValue = false
			attr[attrPos] = key
			attrPos++
		}
	}
	tag := tm.path.NextTag()
	if attrPos == 0 {
		tag.Name = strings.TrimSpace(tagText)
	}else {
		tag.Name = tagText[:tagNameEnd]
	}
	if math.Mod(attrPos,4) == 0{
		for i := 0; i < attrPos; i = i + 4 {
			tag.AddAttribute(strings.TrimSpace(tagText[attr[i]:attr[i + 1]]), tagText[attr[i + 2]:attr[i + 3]])
		}
	}else {
		panic("Parser tag sttribute error")
	}
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
	if tm.path.PathPos >= pathQueryLength && (tm.queryHasPath || tm.queryHasAttributes) {
		for i := pathQueryLength - 1; i >= 0; i-- {
			if len(tm.query.Path[i].Name) != 0 && tm.query.Path[i].Name != tm.path.Path[i + delta].Name {
				return false
			}
			queryAttr := tm.query.Path[i].Attributes
			pathAttr := tm.path.Path[i].Attributes
			expectedMatches = tm.query.Path[i].AttributePos
			for j := 0; j < tm.query.Path[i].AttributePos; j++ {
				for g := 0; g < tm.path.Path[i].AttributePos; g++ {
					if queryAttr[j].Key == pathAttr[g].Key {
						if len(queryAttr[j].Value) == 0 || queryAttr[j].Value == pathAttr[g].Value {
							actualMatches++
						}
					}
				}
			}
			if expectedMatches != actualMatches {
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


