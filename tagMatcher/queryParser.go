package tagMatcher

import "strings"

func parse(query string) []tag {
	split := strings.Split(query, "/")
	tagPath := NewTagPath()
	for key, value := range split {
		if len(strings.TrimSpace(value)) != 0{
			elem := strings.Split(value,"?")
			if len(elem) != 0{
				if len(elem[0]) != 0{

				}
			}
		}
	}
}