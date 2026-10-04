package queryparser

import (
	"fmt"
	"strings"

	"github.com/tcw/saxer/tagpath"
)

// Parse turns a query such as "a/b?id=1&ref" into a TagPath. Each "/"
// separated segment is a tag name, optionally followed by "?" and "&"
// separated attributes, each either "key" or "key=value".
func Parse(query string) (*tagpath.TagPath, error) {
	path := tagpath.NewTagPath()
	for _, value := range strings.Split(query, "/") {
		tagText := strings.TrimSpace(value)
		if len(tagText) == 0 {
			continue
		}
		if path.PathPos >= tagpath.MaxDepth {
			return nil, fmt.Errorf("query has more than %d levels", tagpath.MaxDepth)
		}
		if err := addTag(tagText, path.NextTag()); err != nil {
			return nil, fmt.Errorf("invalid query %q: %w", query, err)
		}
	}
	return path, nil
}

func addTag(tagText string, tag *tagpath.Tag) error {
	name, attrs, hasAttrs := strings.Cut(tagText, "?")
	tag.Name = name
	if !hasAttrs || len(attrs) == 0 {
		return nil
	}
	if strings.Contains(attrs, "?") {
		return fmt.Errorf("more than one '?' in %q", tagText)
	}
	for _, attr := range strings.Split(attrs, "&") {
		if tag.AttributePos >= tagpath.MaxAttributes {
			return fmt.Errorf("more than %d attributes in %q", tagpath.MaxAttributes, tagText)
		}
		key, value, _ := strings.Cut(attr, "=")
		if len(key) == 0 {
			return fmt.Errorf("attribute without a name in %q", tagText)
		}
		if strings.Contains(value, "=") {
			return fmt.Errorf("more than one '=' in attribute %q", attr)
		}
		tag.AddAttribute(key, value)
	}
	return nil
}
