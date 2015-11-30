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
								if strings.Contains(val, "=") {
									attrKeyVal := strings.Split(val, "=")
									tag.Add(tagPath.Attribute{attrKeyVal[0], attrKeyVal[1]})
								}else {
									tag.Add(tagPath.Attribute{val,""})
								}
							}
						}else {

						}
					}
				}
			}else {
				tag.Name = tagText
			}
			path.Add(tag)
		}
	}
	return path
}