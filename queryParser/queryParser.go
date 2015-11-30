package queryParser

import (
	"strings"
	"github.com/tcw/saxer/tagPath"
)

func Parse(query string) tagPath.TagPath {
	split := strings.Split(query, "/")
	path := tagPath.NewTagPath()
	for _, value := range split {
		tagText := strings.TrimSpace(value)
		if len(tagText) != 0 {
			path.Add(getTag(tagText))
		}
	}
	return path
}

func getTag(tagText string) tagPath.Tag{
	tag := tagPath.NewTag()
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
						tag.Add(getAttribute(val))
					}
				}else {
					tag.Add(getAttribute(elem[1]))
				}
			}
		}
	}else {
		tag.Name = tagText
	}
	return tag
}

func getAttribute(attr string) tagPath.Attribute {
	if strings.Contains(attr, "=") {
		attrKeyVal := strings.Split(attr, "=")
		return tagPath.Attribute{attrKeyVal[0], attrKeyVal[1]}
	}else {
		return tagPath.Attribute{attr, ""}
	}
}