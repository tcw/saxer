package tagmatcher

import (
	"fmt"
	"strings"

	"github.com/tcw/saxer/queryparser"
	"github.com/tcw/saxer/tagpath"
)

// Options control how query tag names and attributes are compared with
// the document. The zero value matches exactly and case sensitively.
type Options struct {
	Contains        bool // a query name or value matches when it is a substring
	CaseInsensitive bool
	OmitNamespace   bool // ignore "ns:" prefixes on tag names
}

type TagMatcher struct {
	query         tagpath.TagPath
	queryIsEmpty  bool
	path          tagpath.TagPath
	lastMatchPath []string
	lastMatchPos  int
	tmpAttr       []int
	tmpAttrPos    int
	opts          Options
	equal         func(query, source string) bool
}

func equals(query, source string) bool {
	return query == source
}

func contains(query, source string) bool {
	return strings.Contains(source, query)
}

func NewTagMatcher(queryString string, opts Options) (TagMatcher, error) {
	q, err := queryparser.Parse(queryString)
	if err != nil {
		return TagMatcher{}, err
	}
	tm := TagMatcher{
		query:         *q,
		queryIsEmpty:  true,
		path:          *tagpath.NewTagPath(),
		lastMatchPath: make([]string, tagpath.MaxDepth),
		tmpAttr:       make([]int, 4*tagpath.MaxAttributes),
		opts:          opts,
		equal:         equals,
	}
	if opts.Contains {
		tm.equal = contains
	}
	// Normalise the query once so matching only has to normalise the document.
	for i := 0; i < q.PathPos; i++ {
		tag := &tm.query.Path[i]
		tag.Name = tm.normaliseName(tag.Name)
		for j := 0; j < tag.AttributePos; j++ {
			tag.Attributes[j].Key = tm.normaliseCase(tag.Attributes[j].Key)
			tag.Attributes[j].Value = tm.normaliseCase(tag.Attributes[j].Value)
		}
		if tag.Name != "" || tag.AttributePos > 0 {
			tm.queryIsEmpty = false
		}
	}
	return tm, nil
}

func (tm *TagMatcher) normaliseCase(s string) string {
	if tm.opts.CaseInsensitive {
		return strings.ToLower(s)
	}
	return s
}

func (tm *TagMatcher) normaliseName(name string) string {
	if tm.opts.OmitNamespace {
		if _, local, found := strings.Cut(name, ":"); found {
			name = local
		}
	}
	return tm.normaliseCase(name)
}

func (tm *TagMatcher) GetCurrentPath() string {
	return tm.path.GetCurrentPath()
}

func (tm *TagMatcher) AddTag(tagText string) error {
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
		if tm.tmpAttrPos >= 4*tagpath.MaxAttributes {
			return fmt.Errorf("more than %d attributes in tag <%s>", tagpath.MaxAttributes, tagText)
		}
		if isSpace(value) && !insideAttrValue {
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
	if tm.path.PathPos >= tagpath.MaxDepth {
		return fmt.Errorf("elements nested deeper than %d levels are not supported", tagpath.MaxDepth)
	}
	tag := tm.path.NextTag()
	if tm.tmpAttrPos == 0 {
		tag.Name = strings.TrimSpace(tagText)
	} else {
		tag.Name = tagText[:tagNameEnd]
	}
	if tm.tmpAttrPos%4 != 0 {
		return fmt.Errorf("malformed attributes in tag <%s>", tagText)
	}
	for i := 0; i < tm.tmpAttrPos; i = i + 4 {
		key := strings.TrimSpace(tagText[tm.tmpAttr[i]:tm.tmpAttr[i+1]])
		value := tagText[tm.tmpAttr[i+2]:tm.tmpAttr[i+3]]
		tag.AddAttribute(tm.normaliseCase(key), tm.normaliseCase(value))
	}
	return nil
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
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

// MatchesPath reports whether the innermost elements of the current path
// match the query, and if so remembers the path as the last match.
func (tm *TagMatcher) MatchesPath() bool {
	queryLen := tm.query.PathPos
	if tm.queryIsEmpty || tm.path.PathPos < queryLen {
		return false
	}
	offset := tm.path.PathPos - queryLen
	for i := queryLen - 1; i >= 0; i-- {
		if !tm.tagMatches(&tm.query.Path[i], &tm.path.Path[i+offset]) {
			return false
		}
	}
	tm.lastMatchPos = tm.path.PathPos
	for i := 0; i < tm.path.PathPos; i++ {
		tm.lastMatchPath[i] = tm.path.Path[i].Name
	}
	return true
}

func (tm *TagMatcher) tagMatches(query, tag *tagpath.Tag) bool {
	if query.Name != "" && !tm.equal(query.Name, tm.normaliseName(tag.Name)) {
		return false
	}
	return tm.attributesMatch(query, tag)
}

// attributesMatch reports whether every query attribute is matched by an
// attribute of tag. A query attribute without a value matches on key only.
// Attribute keys and values of both are already case normalised.
func (tm *TagMatcher) attributesMatch(query, tag *tagpath.Tag) bool {
	for _, want := range query.Attributes[:query.AttributePos] {
		found := false
		for _, have := range tag.Attributes[:tag.AttributePos] {
			if tm.equal(want.Key, have.Key) && (want.Value == "" || tm.equal(want.Value, have.Value)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
