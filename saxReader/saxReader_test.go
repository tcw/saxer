package saxReader

import (
	"testing"
	"bytes"
	"io"
	"github.com/zacg/testify/assert"
//	"os"
	"os"
)


func newTestSaxReader(reader io.Reader, emitterTestFn func(string), query string) SaxReader {
	return SaxReader{ElementBufferSize:10,
		ContentBufferSize:1024,
		ReaderBufferSize:10,
		PathDepthSize:10,
		Reader:reader,
		EmitterFn : emitterTestFn,
		PathQuery:query,
		IsInnerXml:false}
}

func TestParseXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(reader, emitter, "hello")
	saxReader.Read()
	assert.Equal(t, res, "<hello>test</hello>")
}

func TestParseXmlOneNodeEmptySearch(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(reader, emitter, "")
	saxReader.Read()
	assert.Equal(t, res, "")
}

func TestParseInnerXmlOneNode(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(reader, emitter, "hello")
	saxReader.IsInnerXml = true
	saxReader.Read()
	assert.Equal(t, res, "test")
}


func TestParseXmlNodeConstrainedBuffer(t *testing.T) {
	res := ""
	emitter := func(element string) {
		res = element
	};
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	saxReader := newTestSaxReader(reader, emitter, "hello")
	saxReader.ReaderBufferSize = 1
	saxReader.Read()
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
	saxReader := newTestSaxReader(reader, emitter, "helloA/helloB/helloC")
	saxReader.ReaderBufferSize = 1
	saxReader.Read()
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
	saxReader := newTestSaxReader(reader, emitter, "helloA/helloB/helloC")
	saxReader.ReaderBufferSize = 1
	saxReader.Read()
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
	saxReader := newTestSaxReader(reader, emitter, "helloA/helloB/helloC")
	saxReader.Read()
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
	saxReader := newTestSaxReader(reader, emitter, "helloA/helloB/helloC")
	saxReader.Read()
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseXmlNodesWithCdataAndCommentConstrainedBuffer(t *testing.T) {
	var actuals []string = make([]string,10)
	var actualsPos int = 0
	emitter := func(element string){
		actuals[actualsPos] = element
		actualsPos++
	};
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	saxReader := newTestSaxReader(reader,emitter, "helloA/helloB/helloC")
	saxReader.ReaderBufferSize = 1
	saxReader.Read()
	assert.Equal(t, actuals[0], "<helloC><![CDATA[Hello<! World!]]></helloC>")
	assert.Equal(t, actuals[1], "<helloC>C2</helloC>")
}

func TestParseWithEscapeAndCDATA(t *testing.T) {
	var actuals []string = make([]string,10)
	var actualsPos int = 0
	emitter := func(element string){
		actuals[actualsPos] = element
		actualsPos++
	};
	reader,err := os.Open("test.xml")
	if err != nil{
		t.Error(err)
	}
	saxReader := newTestSaxReader(reader,emitter, "mediawiki/page/revision/contributor/id")
	saxReader.ElementBufferSize = 100
	saxReader.Read()
	assert.Equal(t, actuals[0], "<id>14954744</id>")
	assert.Equal(t, actuals[1], "<id>3761856</id>")
	assert.Equal(t, actuals[2],  "<id>12070</id>")
	assert.Equal(t, actuals[3], "<id>212624</id>")
	assert.Equal(t, actuals[4], "<id>6569922</id>")
}