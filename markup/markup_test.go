package markup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitName(t *testing.T) {
	for _, tt := range []struct{ tag, name, rest string }{
		{"a", "a", ""},
		{"a x=\"1\"", "a", " x=\"1\""},
		{"xs:doors\n\tid='2'", "xs:doors", "\n\tid='2'"},
	} {
		name, rest := SplitName([]byte(tt.tag))
		assert.Equal(t, tt.name, string(name))
		assert.Equal(t, tt.rest, string(rest))
	}
}

func attrs(t *testing.T, rest string) [][2]string {
	t.Helper()
	got, err := ParseAttributes([]byte(rest), nil)
	require.NoError(t, err)
	var out [][2]string
	for _, a := range got {
		out = append(out, [2]string{string(a.Key), string(a.Value)})
	}
	return out
}

func TestParseAttributes(t *testing.T) {
	assert.Nil(t, attrs(t, ""))
	assert.Nil(t, attrs(t, "   \n"))
	assert.Equal(t, [][2]string{{"id", "1"}}, attrs(t, ` id="1"`))
	assert.Equal(t, [][2]string{{"attrvalue", "one two"}, {"attrvalue2", "1 2 3 4"}},
		attrs(t, "   attrvalue     =   \"one two\"  attrvalue2     =   '1 2 3 4'   "))
	assert.Equal(t, [][2]string{{"id", "1"}, {"lang", "en"}}, attrs(t, "\n\tid=\"1\"\r\n\tlang=\"en\""))
	assert.Equal(t, [][2]string{{"x", "1>2"}, {"y", `a"b`}}, attrs(t, ` x="1>2" y='a"b'`))
	assert.Equal(t, [][2]string{{"xml:space", "preserve"}, {"e", ""}}, attrs(t, ` xml:space="preserve" e=""`))
}

func TestParseAttributesManyAttributes(t *testing.T) {
	rest := ""
	for i := 0; i < 500; i++ {
		rest += ` a="1"`
	}
	got, err := ParseAttributes([]byte(rest), nil)
	require.NoError(t, err)
	assert.Len(t, got, 500)
}

func TestParseAttributesMalformed(t *testing.T) {
	for _, rest := range []string{` x=1`, ` x="1" y`, ` x`, ` x=`, ` x="1`, ` ="1"`, ` "1"`} {
		_, err := ParseAttributes([]byte(rest), nil)
		assert.Error(t, err, rest)
	}
}

func TestAppendUnescaped(t *testing.T) {
	for in, want := range map[string]string{
		"plain":                      "plain",
		"a&amp;b":                    "a&b",
		"&lt;x&gt; &quot;&apos;":     "<x> \"'",
		"&amp;lt;":                   "&lt;",
		"&#38;&#x26;&#X26;&#233;":    "&&&é",
		"&unknown; & &;&#xZZ; &#0;":  "&unknown; & &;&#xZZ; &#0;",
		"trailing &":                 "trailing &",
		"&averyveryverylongname; ok": "&averyveryverylongname; ok",
	} {
		assert.Equal(t, want, string(AppendUnescaped(nil, []byte(in))), in)
	}
}

func TestAppendText(t *testing.T) {
	for in, want := range map[string]string{
		"blue":                          "blue",
		"  blue\n ":                     "blue",
		"<b x=\"a>b\">one</b> two":      "one two",
		"a<!-- <c>x</c> -->b<?pi x?>c":  "abc",
		"<![CDATA[<raw> &amp;]]> &amp;": "<raw> &amp; &",
		"\n\t<Fuel>Diesel</Fuel>\n":     "Diesel",
		"<![CDATA[unterminated":         "unterminated",
		"<a/><b></b>":                   "",
	} {
		assert.Equal(t, want, string(AppendText(nil, []byte(in))), in)
	}
	assert.Equal(t, "keep:x", string(AppendText([]byte("keep:"), []byte(" x "))))
}
