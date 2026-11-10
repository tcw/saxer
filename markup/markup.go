// Package markup holds small, allocation free helpers for the parts of XML
// syntax that saxer looks inside: tag names, attributes, entities and the
// text of an element.
package markup

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

// Attr is an attribute of a tag. Value is raw: entities are not decoded.
type Attr struct {
	Key   []byte
	Value []byte
}

// IsSpace reports whether c is XML white space.
func IsSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// SplitName splits the inside of a start tag (without '<', '>' and a
// trailing '/') into the tag name and the rest, which holds the attributes.
func SplitName(tag []byte) (name, rest []byte) {
	for i, c := range tag {
		if IsSpace(c) {
			return tag[:i], tag[i:]
		}
	}
	return tag, nil
}

// ParseAttributes appends the attributes in rest, the part of a start tag
// after its name, to dst. Values must be quoted with ' or ".
func ParseAttributes(rest []byte, dst []Attr) ([]Attr, error) {
	i := 0
	for {
		for i < len(rest) && IsSpace(rest[i]) {
			i++
		}
		if i == len(rest) {
			return dst, nil
		}
		keyStart := i
		for i < len(rest) && rest[i] != '=' && !IsSpace(rest[i]) {
			if rest[i] == '"' || rest[i] == '\'' {
				return dst, fmt.Errorf("malformed attribute %q", rest[keyStart:i+1])
			}
			i++
		}
		key := rest[keyStart:i]
		if len(key) == 0 {
			return dst, errors.New("attribute without a name")
		}
		for i < len(rest) && IsSpace(rest[i]) {
			i++
		}
		if i == len(rest) || rest[i] != '=' {
			return dst, fmt.Errorf("attribute %q has no value", key)
		}
		i++
		for i < len(rest) && IsSpace(rest[i]) {
			i++
		}
		if i == len(rest) || (rest[i] != '"' && rest[i] != '\'') {
			return dst, fmt.Errorf("value of attribute %q is not quoted", key)
		}
		quote := rest[i]
		end := bytes.IndexByte(rest[i+1:], quote)
		if end < 0 {
			return dst, fmt.Errorf("value of attribute %q has no closing quote", key)
		}
		dst = append(dst, Attr{Key: key, Value: rest[i+1 : i+1+end]})
		i += end + 2
	}
}

// AppendUnescaped appends src to dst with the entity references
// &lt; &gt; &amp; &quot; &apos; and character references such as &#38; and
// &#x26; decoded. Anything else, including unknown entities, is kept as is.
func AppendUnescaped(dst, src []byte) []byte {
	for {
		i := bytes.IndexByte(src, '&')
		if i < 0 {
			return append(dst, src...)
		}
		dst = append(dst, src[:i]...)
		src = src[i:]
		end := bytes.IndexByte(src, ';')
		if end < 0 || end > 12 {
			dst = append(dst, '&')
			src = src[1:]
			continue
		}
		if r, ok := entity(src[1:end]); ok {
			dst = utf8.AppendRune(dst, r)
			src = src[end+1:]
			continue
		}
		dst = append(dst, '&')
		src = src[1:]
	}
}

func entity(name []byte) (rune, bool) {
	switch string(name) {
	case "lt":
		return '<', true
	case "gt":
		return '>', true
	case "amp":
		return '&', true
	case "quot":
		return '"', true
	case "apos":
		return '\'', true
	}
	if len(name) < 2 || name[0] != '#' {
		return 0, false
	}
	base, digits := 10, name[1:]
	if digits[0] == 'x' || digits[0] == 'X' {
		base, digits = 16, digits[1:]
	}
	n, err := strconv.ParseUint(string(digits), base, 32)
	if err != nil || n == 0 || n > utf8.MaxRune {
		return 0, false
	}
	return rune(n), true
}

// AppendText appends the text of the xml fragment src to dst: character
// data with entities decoded and CDATA sections unwrapped, tags, comments
// and processing instructions left out. White space at both ends of the
// result is trimmed.
func AppendText(dst, src []byte) []byte {
	start := len(dst)
	for len(src) > 0 {
		i := bytes.IndexByte(src, '<')
		if i < 0 {
			dst = AppendUnescaped(dst, src)
			break
		}
		dst = AppendUnescaped(dst, src[:i])
		src = src[i:]
		switch {
		case bytes.HasPrefix(src, []byte("<![CDATA[")):
			end := bytes.Index(src, []byte("]]>"))
			if end < 0 {
				end = len(src)
				dst = append(dst, src[9:]...)
				src = nil
				continue
			}
			dst = append(dst, src[9:end]...)
			src = src[end+3:]
		case bytes.HasPrefix(src, []byte("<!--")):
			src = skipPast(src, "-->")
		case bytes.HasPrefix(src, []byte("<?")):
			src = skipPast(src, "?>")
		default:
			src = src[tagEnd(src):]
		}
	}
	text := bytes.TrimSpace(dst[start:])
	n := copy(dst[start:], text)
	return dst[:start+n]
}

func skipPast(src []byte, term string) []byte {
	if i := bytes.Index(src, []byte(term)); i >= 0 {
		return src[i+len(term):]
	}
	return nil
}

// tagEnd returns the index after the '>' that ends the tag at the start of
// src, skipping quoted attribute values.
func tagEnd(src []byte) int {
	var quote byte
	for i, c := range src {
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return i + 1
		}
	}
	return len(src)
}
