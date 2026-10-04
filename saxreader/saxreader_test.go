package saxreader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tcw/saxer/contentbuffer"
	"github.com/tcw/saxer/tagmatcher"
)

func newTestSaxReader(emitterTestFn func(*contentbuffer.EmitterData) bool) SaxReader {

	return SaxReader{ElementBufferSize: 100,
		ContentBufferSize: 1024,
		ReaderBufferSize:  10,
		EmitterFn:         emitterTestFn,
		IsInnerXml:        false,
	}
}

func TestParseXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello>test</hello>")
}

func TestParseXmlOneNodeEmptySearch(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "")
}

func TestParseInnerXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	sax := newTestSaxReader(emitter)
	sax.IsInnerXml = true
	tm := newTagMatcher(t, "hello")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "test")
}

func TestParseXmlNodeConstrainedBuffer(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	sax := newTestSaxReader(emitter)
	sax.ReaderBufferSize = 1
	tm := newTagMatcher(t, "hello")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello>test</hello>")
}

func TestParseXmlNodesConstrainedBuffer(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(ed *contentbuffer.EmitterData) bool {
		actuals[actualsPos] = ed.Content
		actualsPos++
		return false
	}
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	sax := newTestSaxReader(emitter)
	sax.ReaderBufferSize = 1
	tm := newTagMatcher(t, "helloA/helloB/helloC")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC>C1</helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithComments(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(ed *contentbuffer.EmitterData) bool {
		actuals[actualsPos] = ed.Content
		actualsPos++
		return false
	}
	reader := bytes.NewReader([]byte("<helloA><helloB><!-- test<>--<><--><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	sax := newTestSaxReader(emitter)
	sax.ReaderBufferSize = 1
	tm := newTagMatcher(t, "helloA/helloB/helloC")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC>C1</helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdata(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(ed *contentbuffer.EmitterData) bool {
		actuals[actualsPos] = ed.Content
		actualsPos++
		return false
	}
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "helloA/helloB/helloC")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdataAndComment(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(ed *contentbuffer.EmitterData) bool {
		actuals[actualsPos] = ed.Content
		actualsPos++
		return false
	}
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "helloA/helloB/helloC")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdataAndCommentConstrainedBuffer(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(ed *contentbuffer.EmitterData) bool {
		actuals[actualsPos] = ed.Content
		actualsPos++
		return false
	}
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	sax := newTestSaxReader(emitter)
	sax.ReaderBufferSize = 1
	tm := newTagMatcher(t, "helloA/helloB/helloC")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithLtEscapeTag(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<helloA><helloB>&lt;helloC>&lt;/helloC></helloB></helloA>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "helloA/helloB")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<helloB>&lt;helloC>&lt;/helloC></helloB>")
}

func TestParseXmlOneNodeWithLtError(t *testing.T) {

	emitter := func(ed *contentbuffer.EmitterData) bool {
		return false
	}
	reader := bytes.NewReader([]byte("<he<llo>test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello")
	err := sax.Read(reader, &tm)
	assert.NotNil(t, err)
}

func TestParseXmlOneNodeWithEndElementBeforeStartError(t *testing.T) {
	emitter := func(ed *contentbuffer.EmitterData) bool {
		return false
	}
	reader := bytes.NewReader([]byte("</hello>test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello")
	err := sax.Read(reader, &tm)
	assert.NotNil(t, err)
}

func TestParseXmlOneNodeOneAttributeDoubleQuote(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello id=\"123\">test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello?id=123")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello id=\"123\">test</hello>")
}

func TestParseXmlOneNodeOneAttributeSingle(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello id='123'>test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello?id=123")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello id='123'>test</hello>")
}

func TestParseXmlOneNodeTwoAttributes(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello id=\"123\" ref=\"42\">test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello?id=123&ref=42")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello id=\"123\" ref=\"42\">test</hello>")
}

func TestParseXmlOneNodeTwoAttributesNoMatch(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello id=\"123\" ref=\"42\">test</hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello?id=123&ref=421")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.NotEqual(t, res, "<hello id=\"123\" ref=\"42\">test</hello>")
}

func TestParseXmlTest(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello id=\"123\" ref=\"42\"><hello2 idx=\"1234\" refx=\"421\">test</hello2></hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "hello?id=123&ref=42/hello2?idx=1234&refx=421")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello2 idx=\"1234\" refx=\"421\">test</hello2>")
}

func TestParseXmlTestS(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello><text xml:space=\"preserve\">this</text></hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "text?xml:space")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<text xml:space=\"preserve\">this</text>")
}

func TestParseXmlTestS2(t *testing.T) {
	res := ""
	emitter := func(ed *contentbuffer.EmitterData) bool {
		res = ed.Content
		return false
	}
	reader := bytes.NewReader([]byte("<hello><text xml:space=\"preserve\">this</text><hello2>Test2</hello2><hello3>Test3</hello3></hello>"))
	sax := newTestSaxReader(emitter)
	tm := newTagMatcher(t, "?xml:space")
	err := sax.Read(reader, &tm)
	assert.Nil(t, err)
	assert.Equal(t, res, "<text xml:space=\"preserve\">this</text>")
}

// readAll runs sr over r with query and collects every emitted element.
func readAll(sr SaxReader, r io.Reader, query string) ([]string, error) {
	var got []string
	sr.EmitterFn = func(ed *contentbuffer.EmitterData) bool {
		got = append(got, ed.Content)
		return false
	}
	tm, err := tagmatcher.NewTagMatcher(query, tagmatcher.Options{})
	if err != nil {
		return nil, err
	}
	err = sr.Read(r, &tm)
	return got, err
}

func newTagMatcher(t *testing.T, query string) tagmatcher.TagMatcher {
	t.Helper()
	tm, err := tagmatcher.NewTagMatcher(query, tagmatcher.Options{})
	require.NoError(t, err)
	return tm
}

// Wrappers that deliver the same bytes in different, but legal, ways.
var readerWrappers = map[string]func(io.Reader) io.Reader{
	"plain":   func(r io.Reader) io.Reader { return r },
	"onebyte": iotest.OneByteReader,
	"half":    iotest.HalfReader,
	"dataerr": iotest.DataErrReader,
}

// The result must not depend on how the input is chunked by the reader.
func TestReadIsIndependentOfChunking(t *testing.T) {
	tests := []struct {
		name  string
		xml   string
		query string
		want  []string
	}{
		{"attributes", `<a><b x="1">one</b><b x="2">two</b></a>`, "a/b",
			[]string{`<b x="1">one</b>`, `<b x="2">two</b>`}},
		{"attribute filter", `<a><b x="1">one</b><b x="2">two</b></a>`, "b?x=2",
			[]string{`<b x="2">two</b>`}},
		{"self-closing", `<a><b x="1"/><b x="2"/></a>`, "b",
			[]string{`<b x="1"/>`, `<b x="2"/>`}},
		{"comment and cdata", `<a><!-- <b> --><b><![CDATA[<b>]]></b></a>`, "b",
			[]string{`<b><![CDATA[<b>]]></b>`}},
		{"prolog and doctype", `<?xml version="1.0"?><!DOCTYPE a><a><b>v</b></a>`, "b",
			[]string{`<b>v</b>`}},
		{"attributes on new lines", "<a>\n<b\n  x=\"1\"\n  y=\"2\">v</b></a>", "b?y=2",
			[]string{"<b\n  x=\"1\"\n  y=\"2\">v</b>"}},
		{"long tag", `<a><b attr="` + strings.Repeat("x", 50) + `">v</b></a>`, "b",
			[]string{`<b attr="` + strings.Repeat("x", 50) + `">v</b>`}},
	}
	for _, tt := range tests {
		for wrapName, wrap := range readerWrappers {
			for _, bufSize := range []int{1, 2, 3, 5, 8, 64, 4096} {
				t.Run(fmt.Sprintf("%s/%s/buf%d", tt.name, wrapName, bufSize), func(t *testing.T) {
					sr := New()
					sr.ReaderBufferSize = bufSize
					got, err := readAll(sr, wrap(strings.NewReader(tt.xml)), tt.query)
					require.NoError(t, err)
					assert.Equal(t, tt.want, got)
				})
			}
		}
	}
}

func TestReadDoctype(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{"simple", `<!DOCTYPE a><a><b>v</b></a>`},
		{"public id", `<!DOCTYPE a PUBLIC "-//x//EN" "a>b.dtd"><a><b>v</b></a>`},
		{"internal subset", `<!DOCTYPE a [<!ELEMENT a (b)><!ENTITY e "x>y">]><a><b>v</b></a>`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readAll(New(), strings.NewReader(tt.xml), "b")
			require.NoError(t, err)
			assert.Equal(t, []string{"<b>v</b>"}, got)
		})
	}
}

func TestReadGreaterThanInAttributeValue(t *testing.T) {
	got, err := readAll(New(), strings.NewReader(`<a><b x="1>2" y='3>4'>v</b></a>`), "b?x=1>2")
	require.NoError(t, err)
	assert.Equal(t, []string{`<b x="1>2" y='3>4'>v</b>`}, got)
}

func TestReadReturnsReaderError(t *testing.T) {
	readErr := errors.New("disk on fire")
	r := io.MultiReader(strings.NewReader("<a><b>v</b>"), iotest.ErrReader(readErr))
	_, err := readAll(New(), r, "b")
	assert.ErrorIs(t, err, readErr)
}

func TestReadTagLargerThanTagBuffer(t *testing.T) {
	sr := New()
	sr.ReaderBufferSize = 8
	sr.ElementBufferSize = 16
	_, err := readAll(sr, strings.NewReader(`<a><b attr="`+strings.Repeat("x", 100)+`">v</b></a>`), "b")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--tag-buf")
}

func TestReadContentLargerThanContentBuffer(t *testing.T) {
	sr := New()
	sr.ContentBufferSize = 16
	_, err := readAll(sr, strings.NewReader(`<a><b>`+strings.Repeat("x", 100)+`</b></a>`), "b")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--cont-buf")
}

func TestReadTruncatedDocument(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want []string
	}{
		{"inside element", `<a><b>one`, nil},
		{"unclosed root", `<a><b>x</b>`, []string{"<b>x</b>"}},
		{"inside tag", `<a><b x="1`, nil},
		{"inside comment", `<a><!-- x`, nil},
		{"inside cdata", `<a><![CDATA[ x`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readAll(New(), strings.NewReader(tt.xml), "b")
			assert.Equal(t, tt.want, got)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unexpected end of input")
		})
	}
}

func TestReadMalformedInputReturnsError(t *testing.T) {
	for _, xml := range []string{
		`<>`,
		`<a></a></a>`,
		`<a><b x=1>v</b></a>`,
		`<a><b x="1" y>v</b></a>`,
		strings.Repeat("<a>", 1000),
	} {
		t.Run(xml, func(t *testing.T) {
			_, err := readAll(New(), strings.NewReader(xml), "b")
			assert.Error(t, err)
		})
	}
}

func TestReadStopsWhenEmitterSaysSo(t *testing.T) {
	var got []string
	sr := New()
	sr.EmitterFn = func(ed *contentbuffer.EmitterData) bool {
		got = append(got, ed.Content)
		return len(got) == 2
	}
	tm := newTagMatcher(t, "b")
	err := sr.Read(strings.NewReader(`<a><b>1</b><b>2</b><b>3</b></a>`), &tm)
	require.NoError(t, err)
	assert.Equal(t, []string{"<b>1</b>", "<b>2</b>"}, got)
}

// The parser must never panic, and its result must not depend on chunking.
func FuzzRead(f *testing.F) {
	for _, seed := range []string{
		`<a><b x="1">one</b><b x="2">two</b></a>`,
		`<?xml version="1.0"?><!DOCTYPE a [<!ENTITY e "x">]><a><!-- c --><b><![CDATA[<]]></b></a>`,
		`<a><b x="1>2"/></a>`,
		`<a><b>`,
	} {
		f.Add(seed, uint8(3))
	}
	f.Fuzz(func(t *testing.T, xml string, bufSize uint8) {
		sr := New()
		sr.ElementBufferSize = 64
		sr.ContentBufferSize = 256
		want, wantErr := readAll(sr, strings.NewReader(xml), "b")
		sr.ReaderBufferSize = int(bufSize%16) + 1
		got, gotErr := readAll(sr, iotest.HalfReader(strings.NewReader(xml)), "b")
		if wantErr == nil {
			assert.NoError(t, gotErr)
			assert.Equal(t, want, got)
		}
	})
}
