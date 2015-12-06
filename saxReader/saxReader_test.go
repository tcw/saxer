package saxReader

import (
	"testing"
	"bytes"
	"github.com/zacg/testify/assert"
	"fmt"
)


func newTestSaxReader(emitterTestFn func(string) bool) SaxReader {
	return SaxReader{ElementBufferSize:100,
		ContentBufferSize:1024,
		ReaderBufferSize:10,
		PathDepthSize:10,
		EmitterFn : emitterTestFn,
		IsInnerXml:false,
	}
}

func TestParseXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello>test</hello>")
}

func TestParseXmlOneNodeEmptySearch(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "")
	assert.Nil(t, err)
	assert.Equal(t, res, "")
}

func TestParseInnerXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
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
	emitter := func(element string) bool {
		res = element
		return false
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
	emitter := func(element string) bool {
		actuals[actualsPos] = element
		actualsPos++
		return false
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
	emitter := func(element string) bool {
		actuals[actualsPos] = element
		actualsPos++
		return false
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
	emitter := func(element string) bool {
		actuals[actualsPos] = element
		actualsPos++
		return false
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
	emitter := func(element string) bool {
		actuals[actualsPos] = element
		actualsPos++
		return false
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
	emitter := func(element string) bool {
		actuals[actualsPos] = element
		actualsPos++
		return false
	};
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	saxReader.ReaderBufferSize = 1
	err := saxReader.Read(reader, "helloA/helloB/helloC")
	assert.Nil(t, err)
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithLtEscapeTag(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<helloA><helloB>&lt;helloC>&lt;/helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "helloA/helloB")
	assert.Nil(t, err)
	assert.Equal(t, res, "<helloB>&lt;helloC>&lt;/helloC></helloB>")
}

func TestParseXmlOneNodeWithLtError(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<he<llo>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello")
	assert.NotNil(t, err)
}

func TestParseXmlOneNodeWithEndElementBeforeStartElementError(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("</hello>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello")
	fmt.Println(err)
	assert.NotNil(t, err)
}

func TestParseXmlOneNodeOneAttributeDoubleQuote(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello id=\"123\">test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello?id=123")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello id=\"123\">test</hello>")
}

func TestParseXmlOneNodeOneAttributeSingle(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello id='123'>test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello?id=123")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello id='123'>test</hello>")
}

func TestParseXmlOneNodeTwoAttributes(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello id=\"123\" ref=\"42\">test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello?id=123&ref=42")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello id=\"123\" ref=\"42\">test</hello>")
}

func TestParseXmlOneNodeTwoAttributesNoMatch(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello id=\"123\" ref=\"42\">test</hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello?id=123&ref=421")
	assert.Nil(t, err)
	assert.NotEqual(t, res, "<hello id=\"123\" ref=\"42\">test</hello>")
}

func TestParseXmlTest(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello id=\"123\" ref=\"42\"><hello2 idx=\"1234\" refx=\"421\">test</hello2></hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "hello?id=123&ref=42/hello2?idx=1234&refx=421")
	assert.Nil(t, err)
	assert.Equal(t, res, "<hello2 idx=\"1234\" refx=\"421\">test</hello2>")
}

func TestParseXmlTestS(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello><text xml:space=\"preserve\">this</text></hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "text?xml:space")
	assert.Nil(t, err)
	assert.Equal(t, res, "<text xml:space=\"preserve\">this</text>")
}

func TestParseXmlTestS2(t *testing.T) {
	res := ""
	emitter := func(element string) bool {
		res = element
		return false
	};
	reader := bytes.NewReader([]byte("<hello><text xml:space=\"preserve\">this</text><hello2>Test2</hello2><hello3>Test3</hello3></hello>"))
	saxReader := newTestSaxReader(emitter)
	err := saxReader.Read(reader, "?xml:space")
	assert.Nil(t, err)
	assert.Equal(t, res, "<text xml:space=\"preserve\">this</text>")
}