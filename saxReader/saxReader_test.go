package saxReader

import (
	"testing"
	"bytes"
	"github.com/zacg/testify/assert"
)


func newTestSaxReader(emitterTestFn func(string)) SaxReader {
	return SaxReader{ElementBufferSize:10,
		ContentBufferSize:1024,
		ReaderBufferSize:10,
		PathDepthSize:10,
		EmitterFn : emitterTestFn,
		IsInnerXml:false,
		FilterEscapeSigns:false,
	}
}

func TestParseXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello>test</hello>")
}

func TestParseXmlOneNodeEmptySearch(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "")
	assert.Nil(t, err)
	assert.Equal(t, res, "")
}

func TestParseInnerXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.IsInnerXml = true
	err := saxReader.Read(reader, "hello")
	assert.Nil(t, err)
	assert.Equal(t, res, "test")
}


func TestParseXmlNodeConstrainedBuffer(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.ReaderBufferSize = 1
	err := saxReader.Read(reader, "hello")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello>test</hello>")
}

func TestParseXmlNodesConstrainedBuffer(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.ReaderBufferSize = 1
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC>C1</helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithComments(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><helloB><!-- test<>--<><--><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.ReaderBufferSize = 1
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC>C1</helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdata(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdataAndComment(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdataAndCommentConstrainedBuffer(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.ReaderBufferSize = 1
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithEscape(t *testing.T) {
	var actuals []string = make([]string, 10)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC>&lt;![CDATA[Hello<! World!]]></helloC>&lt;helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.FilterEscapeSigns = true
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithEscapeAndXmlTag(t *testing.T) {
	var actuals []string = make([]string, 100)
	var actualsPos int = 0
	emitter := func(element string) {
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?><helloA><!-- test<>--<><--><helloB><helloC><?xml version=\"1.0\" encoding=\"UTF-8\"?>&lt;![CDATA[Hello<! World!]]></helloC>&lt;helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.FilterEscapeSigns = true
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><?xml version=\"1.0\" encoding=\"UTF-8\"?><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlOneNodeWithLtError(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<he<llo>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello")
	assert.NotNil(t, err)
}

func TestParseXmlOneNodeWithGtError(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<he>>llo>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello")
	assert.NotNil(t, err)
}