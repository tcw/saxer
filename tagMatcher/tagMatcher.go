package tagMatcher

import (
	"github.com/tcw/saxer/queryParser"
	"github.com/tcw/saxer/tagPath"
	"strings"
)

type TagMatcher struct {
	query              tagPath.TagPath
	queryHasAttributes bool
	queryHasPath       bool
	path               tagPath.TagPath
	lastMatchPath      []string
	lastMatchPos       int
	tmpAttr            []int
	tmpAttrPos         int
	EqualityFn         func(string, string) bool
	CaseSensitive      bool
	WithoutNamespace   bool
}

var EqFnEqulas = func(query string, source string) bool {
	return query == source
}

var EqFnContains = func(query string, source string) bool {
	return strings.Contains(source, query)
}

func NewTagMatcher(queryString string) TagMatcher {
	path := tagPath.NewTagPath()
	last := make([]string, 1000)
	q := queryParser.Parse(queryString)
	qHasAttributes := false
	qHasPath := false
	tAttr := make([]int, 1024*4)
	for i := 0; i < q.PathPos; i++ {
		if q.Path[i].AttributePos > 0 {
			qHasAttributes = true
		}
		if len(q.Path[i].Name) > 0 {
			qHasPath = true
		}
	}

	return TagMatcher{query: *q, queryHasAttributes: qHasAttributes,
		queryHasPath:     qHasPath,
		path:             *path,
		lastMatchPath:    last,
		lastMatchPos:     0,
		tmpAttr:          tAttr,
		tmpAttrPos:       0,
		EqualityFn:       EqFnEqulas,
		CaseSensitive:    true,
		WithoutNamespace: false}
}

func (tm *TagMatcher) GetCurrentPath() string {
	return tm.path.GetCurrentPath()
}

func (tm *TagMatcher) AddTag(tagText string) {
	tagNameEnd := 0
	insideAttrValue := false
	insideAttrKey := false
	readSpace := false
	readEquals := false
	tm.tmpAttrPos = 0
	var separator rune = 0
	cleanedTag := tagText
	if strings.HasSuffix(tagText, "/") {
		cleanedTag = strings.TrimRight(tagText, "/")
	}
	trimmed := strings.TrimSpace(cleanedTag)
	for key, value := range trimmed {
		if value == rune(' ') && !insideAttrValue {
			if tagNameEnd == 0 {
				tagNameEnd = key
			}
			readSpace = true
		} else if value == rune('=') && !insideAttrValue {
			readEquals = true
			tm.tmpAttr[tm.tmpAttrPos] = key
			tm.tmpAttrPos++
		} else if readEquals && !insideAttrValue && (value == rune('\'') || value == rune('"')) {
			separator = value
			insideAttrValue = true
			insideAttrKey = false
			readEquals = false
			tm.tmpAttr[tm.tmpAttrPos] = key + 1
			tm.tmpAttrPos++
		} else if readSpace && !insideAttrValue {
			if !insideAttrKey {
				tm.tmpAttr[tm.tmpAttrPos] = key
				tm.tmpAttrPos++
			}
			insideAttrKey = true
		} else if value == separator && insideAttrValue {
			separator = 0
			insideAttrValue = false
			tm.tmpAttr[tm.tmpAttrPos] = key
			tm.tmpAttrPos++
		}
	}
	tag := tm.path.NextTag()
	if tm.tmpAttrPos == 0 {
		tag.Name = strings.TrimSpace(tagText)
	} else {
		tag.Name = tagText[:tagNameEnd]
	}
	if tm.tmpAttrPos%4 == 0 {
		for i := 0; i < tm.tmpAttrPos; i = i + 4 {
			tag.AddAttribute(strings.TrimSpace(tagText[tm.tmpAttr[i]:tm.tmpAttr[i+1]]), tagText[tm.tmpAttr[i+2]:tm.tmpAttr[i+3]])
		}
	} else {
		panic("Parser tag attribute error")
	}
}

func (np *TagMatcher) RemoveLast() {
	np.path.RemoveLast()
}

func (tm *TagMatcher) TagNameMatchesLastMatch() bool {
	if tm.lastMatchPos == tm.path.PathPos {
		for i := 0; i < tm.lastMatchPos; i++ {
			if tm.lastMatchPath[i] != tm.path.Path[i].Name {
				return false
			}
		}
	} else {
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
			if len(tm.query.Path[i].Name) != 0 {
				var queryTagName string
				var pathTagName string
				if tm.CaseSensitive {
					queryTagName = tm.query.Path[i].Name // TODO Could be done once at init
					pathTagName = tm.path.Path[i+delta].Name
				} else {
					queryTagName = strings.ToLower(tm.query.Path[i].Name) // TODO Could be done once at init
					pathTagName = strings.ToLower(tm.path.Path[i+delta].Name)
				}
				if tm.WithoutNamespace {
					tagNameParts := strings.Split(pathTagName, ":")
					if len(tagNameParts) > 1 {
						pathTagName = tagNameParts[1]
					}
					if !tm.EqualityFn(queryTagName, pathTagName) {
						return false
					}
				} else {
					if !tm.EqualityFn(queryTagName, pathTagName) {
						return false
					}
				}
			}

			queryAttr := tm.query.Path[i].Attributes
			pathAttr := tm.path.Path[i+delta].Attributes
			if !tm.CaseSensitive {
				toLowerCaseInPlace(queryAttr)
				toLowerCaseInPlace(pathAttr)
			}

			expectedMatches = tm.query.Path[i].AttributePos
			for j := 0; j < tm.query.Path[i].AttributePos; j++ { //TODO O(n^2) now, could be O(n(n − 1)/2)
				for g := 0; g < tm.path.Path[i+delta].AttributePos; g++ {
					if tm.EqualityFn(queryAttr[j].Key, pathAttr[g].Key) {
						if len(queryAttr[j].Value) == 0 || tm.EqualityFn(queryAttr[j].Value, pathAttr[g].Value) {
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
	} else {
		return false
	}
}

func toLowerCaseInPlace(elems []tagPath.Attribute) {
	for i := 0; i < len(elems); i++ {
		elems[i].Key = strings.ToLower(elems[i].Key)
		elems[i].Value = strings.ToLower(elems[i].Value)
	}

}
