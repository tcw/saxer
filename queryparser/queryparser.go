package queryparser

import (
	"github.com/tcw/saxer/tagpath"
	"strings"
)

func Parse(query string) *tagpath.TagPath {
	split := strings.Split(query, "/")
	path := tagpath.NewTagPath()
	for _, value := range split {
		tagText := strings.TrimSpace(value)
		if len(tagText) != 0 {
			addTag(tagText, path)
		}
	}
	return path
}

func addTag(tagText string, tp *tagpath.TagPath) {
	tag := tp.NextTag()
	if strings.Contains(tagText, "?") {
		elem := strings.Split(tagText, "?")
		if len(elem) != 0 {
			if len(elem[0]) > 0 {
				tag.Name = elem[0]
			}
			if len(elem[1]) > 0 {
				if strings.Contains(elem[1], "&") {
					attr := strings.Split(elem[1], "&")
					for _, val := range attr {
						addToAttribute(val, tag)
					}
				} else {
					addToAttribute(elem[1], tag)
				}
			}
		}
	} else {
		tag.Name = tagText
	}
}

func addToAttribute(attr string, tg *tagpath.Tag) {
	if strings.Contains(attr, "=") {
		attrKeyVal := strings.Split(attr, "=")
		tg.AddAttribute(attrKeyVal[0], attrKeyVal[1])
	} else {
		tg.AddAttribute(attr, "")
	}
}
