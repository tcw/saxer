package saxreader

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcw/saxer/tagmatcher"
)

// collect runs sr over xml with query and returns copies of the matches.
func collect(t *testing.T, sr SaxReader, xml, query string, opts tagmatcher.Options) ([]Match, error) {
	t.Helper()
	var got []Match
	sr.EmitterFn = func(m *Match) bool {
		c := *m
		c.Content = clone(m.Content)
		c.StartTag = clone(m.StartTag)
		c.Path = clone(m.Path)
		got = append(got, c)
		return false
	}
	tm, err := tagmatcher.New(query, opts)
	require.NoError(t, err)
	err = sr.Read(strings.NewReader(xml), tm)
	return got, err
}

func clone(b []byte) []byte {
	if b == nil {
		return nil
	}
	return append([]byte{}, b...)
}

func contents(ms []Match) []string {
	var out []string
	for _, m := range ms {
		out = append(out, string(m.Content))
	}
	return out
}

func TestEndTagMustMatchStartTag(t *testing.T) {
	for _, xml := range []string{`<a></b>`, `<a><b></a></b>`, `<a><b>x</B></a>`} {
		_, err := readAll(New(), strings.NewReader(xml), "b")
		require.Error(t, err, xml)
		assert.Contains(t, err.Error(), "does not match")
	}
	got, err := readAll(New(), strings.NewReader("<a><b>x</b \n></a >"), "b")
	require.NoError(t, err)
	assert.Equal(t, []string{"<b>x</b \n>"}, got)
}

func TestMatchEndsByDepth(t *testing.T) {
	tests := []struct {
		xml, query string
		want       []string
	}{
		{`<a><b><b>x</b></b><b>y</b></a>`, "b", []string{`<b><b>x</b></b>`, `<b>y</b>`}},
		{`<a><b><a></a></b></a>`, "a/b", []string{`<b><a></a></b>`}},
		{`<b><c><b/></c></b>`, "b", []string{`<b><c><b/></c></b>`}},
		{`<a><b><c/><b x="1"/></b></a>`, "b", []string{`<b><c/><b x="1"/></b>`}},
	}
	for _, tt := range tests {
		got, err := readAll(New(), strings.NewReader(tt.xml), tt.query)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got, tt.xml)
	}
}

func TestDeepNestingAndManyAttributes(t *testing.T) {
	deep := strings.Repeat("<a>", 5000) + "x" + strings.Repeat("</a>", 5000)
	got, err := readAll(New(), strings.NewReader("<r>"+deep+"</r>"), "r/a")
	require.NoError(t, err)
	assert.Equal(t, []string{deep}, got)

	var attrs strings.Builder
	for i := 0; i < 500; i++ {
		fmt.Fprintf(&attrs, ` a%d="%d"`, i, i)
	}
	doc := "<r><b" + attrs.String() + ">v</b></r>"
	sr := New()
	sr.ElementBufferSize = 16 * KB
	got, err = readAll(sr, strings.NewReader(doc), "b?a499=499&a0=0")
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestSelfClosingTags(t *testing.T) {
	for _, tag := range []string{`<b x="1"/>`, `<b x="1"  />`, "<b\n x=\"1\"\n/>", `<b x='1'/>`} {
		got, err := readAll(New(), strings.NewReader("<a>"+tag+"</a>"), "b?x=1")
		require.NoError(t, err)
		assert.Equal(t, []string{tag}, got)
	}
	sr := New()
	sr.IsInnerXml = true
	got, err := readAll(sr, strings.NewReader(`<a><b x="1"/></a>`), "b")
	require.NoError(t, err)
	assert.Equal(t, []string{""}, got)
}

func TestMatchDetails(t *testing.T) {
	xml := "<a>\n<b\n x='1' y=\"&lt;2&gt;\">\nv\n</b>\n<b/></a>"
	got, err := collect(t, New(), xml, "b", tagmatcher.Options{})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, uint64(2), got[0].LineStart)
	assert.Equal(t, uint64(5), got[0].LineEnd)
	assert.Equal(t, "a/b", string(got[0].Path))
	assert.Equal(t, "<b\n x='1' y=\"&lt;2&gt;\">", string(got[0].StartTag))
	attrs, err := got[0].Attributes()
	require.NoError(t, err)
	require.Len(t, attrs, 2)
	assert.Equal(t, "y", string(attrs[1].Key))
	assert.Equal(t, "&lt;2&gt;", string(attrs[1].Value))

	assert.Equal(t, uint64(6), got[1].LineStart)
	assert.Equal(t, "<b/>", string(got[1].StartTag))
	attrs, err = got[1].Attributes()
	require.NoError(t, err)
	assert.Empty(t, attrs)
}

func TestSkipContent(t *testing.T) {
	sr := New()
	sr.SkipContent = true
	got, err := collect(t, sr, "<a>\n<b>x</b><b>\ny</b></a>", "b", tagmatcher.Options{})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Nil(t, got[0].Content)
	assert.Nil(t, got[0].StartTag)
	assert.Equal(t, []uint64{2, 2, 2, 3}, []uint64{got[0].LineStart, got[0].LineEnd, got[1].LineStart, got[1].LineEnd})

	// Nothing is copied, so a match larger than the content buffer is fine.
	sr.ContentBufferSize = 4
	_, err = collect(t, sr, "<a><b>"+strings.Repeat("x", 100)+"</b></a>", "b", tagmatcher.Options{})
	assert.NoError(t, err)
}

func TestTextMatch(t *testing.T) {
	cars := `<cars><car><color>blue</color></car><car><color> red </color></car><car><color><![CDATA[red]]></color></car></cars>`
	tests := []struct {
		name, xml, query string
		opts             tagmatcher.Options
		want             []string
	}{
		{"text", cars, "color~red", tagmatcher.Options{}, []string{`<color> red </color>`, `<color><![CDATA[red]]></color>`}},
		{"path and text", cars, "car/color~blue", tagmatcher.Options{}, []string{`<color>blue</color>`}},
		{"text of descendants", cars, "car~blue", tagmatcher.Options{}, []string{`<car><color>blue</color></car>`}},
		{"no match", cars, "color~green", tagmatcher.Options{}, nil},
		{"case insensitive", `<a><c>Red</c></a>`, "c~RED", tagmatcher.Options{CaseInsensitive: true}, []string{`<c>Red</c>`}},
		{"contains", `<a><c>dark red</c><c>blue</c></a>`, "c~red", tagmatcher.Options{Contains: true}, []string{`<c>dark red</c>`}},
		{"entities", `<a><c>a &amp; b</c></a>`, `c~a \& b`, tagmatcher.Options{}, []string{`<c>a &amp; b</c>`}},
		{"inner match when outer does not", `<a><b><b>x</b>y</b><b>x</b></a>`, "b~x", tagmatcher.Options{}, []string{`<b>x</b>`, `<b>x</b>`}},
		{"outer match wins", `<a><b><b>x</b></b></a>`, "b~x", tagmatcher.Options{}, []string{`<b><b>x</b></b>`}},
		{"several inner matches", `<a><b><b>x</b><c><b>x</b></c><b>z</b></b></a>`, "b~x", tagmatcher.Options{}, []string{`<b>x</b>`, `<b>x</b>`}},
		{"empty text", `<a><b/><b> </b><b>x</b></a>`, "b~", tagmatcher.Options{}, []string{`<b/>`, `<b> </b>`}},
		{"text and attribute", `<a><b k="1">x</b><b k="2">x</b></a>`, "b~x?k=2", tagmatcher.Options{}, []string{`<b k="2">x</b>`}},
		{"wildcard and text", `<a><b>x</b><c>x</c><d><e>x</e></d></a>`, "a/*~x", tagmatcher.Options{}, []string{`<b>x</b>`, `<c>x</c>`, `<d><e>x</e></d>`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collect(t, New(), tt.xml, tt.query, tt.opts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, contents(got))
		})
	}
}

func TestTextMatchDetails(t *testing.T) {
	sr := New()
	sr.IsInnerXml = true
	got, err := collect(t, sr, "<a>\n<b k='1'><b>x</b>y</b>\n<b>x</b></a>", "b~x", tagmatcher.Options{})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "x", string(got[0].Content))
	assert.Equal(t, "<b>", string(got[0].StartTag))
	assert.Equal(t, "a/b/b", string(got[0].Path))
	assert.Equal(t, uint64(2), got[0].LineStart)
	assert.Equal(t, "a/b", string(got[1].Path))
	assert.Equal(t, uint64(3), got[1].LineStart)

	sr.SkipContent = true
	got, err = collect(t, sr, "<a><b>x</b><b>y</b></a>", "b~y", tagmatcher.Options{})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Nil(t, got[0].Content)
}

func TestTextMatchStops(t *testing.T) {
	var got []string
	sr := New()
	sr.EmitterFn = func(m *Match) bool {
		got = append(got, string(m.Content))
		return true
	}
	tm := newTagMatcher(t, "b~x")
	require.NoError(t, sr.Read(strings.NewReader(`<a><b><b>x</b><b>x</b>y</b><b>x</b></a>`), tm))
	assert.Equal(t, []string{"<b>x</b>"}, got)
}

// Comments, CDATA sections and text much larger than the read buffer.
func TestLargeMarkupAcrossReads(t *testing.T) {
	big := strings.Repeat("0123456789-]>?", 1000)
	xml := "<!--" + big + "--><a><?pi " + big + "?><b>" + big + "<![CDATA[" + big + "]]><!--" + big + "--></b></a>"
	for _, size := range []int{1, 7, 16, 4096} {
		sr := New()
		sr.ReaderBufferSize = size
		got, err := readAll(sr, strings.NewReader(xml), "b")
		require.NoError(t, err)
		assert.Equal(t, []string{"<b>" + big + "<![CDATA[" + big + "]]><!--" + big + "--></b>"}, got)

		got, err = readAll(sr, strings.NewReader(xml), "b~"+strings.ReplaceAll(big, "?", `\?`)+strings.ReplaceAll(big, "?", `\?`))
		require.NoError(t, err)
		assert.Len(t, got, 1)
	}
}

func TestErrorsReportTheirLine(t *testing.T) {
	_, err := readAll(New(), strings.NewReader("<a>\n\n<b></c>"), "b")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 3")
}

func TestTagErrors(t *testing.T) {
	for xml, want := range map[string]string{
		"<>":          "empty tag",
		"<a>< b></a>": "without a name",
		"<a></>":      "does not match",
		"</a>":        "before start tag",
	} {
		_, err := readAll(New(), strings.NewReader(xml), "b")
		require.Error(t, err, xml)
		assert.Contains(t, err.Error(), want, xml)
	}
}

func TestTextMatchIsIndependentOfChunking(t *testing.T) {
	xml := `<a><b>x</b><b><b>x</b>y</b><c><![CDATA[x]]></c><!-- <b>x</b> --></a>`
	for _, wrap := range readerWrappers {
		for _, size := range []int{1, 2, 3, 5, 64} {
			sr := New()
			sr.ReaderBufferSize = size
			got, err := readAll(sr, wrap(strings.NewReader(xml)), "*~x")
			require.NoError(t, err)
			assert.Equal(t, []string{"<b>x</b>", "<b>x</b>", "<c><![CDATA[x]]></c>"}, got)
		}
	}
}
