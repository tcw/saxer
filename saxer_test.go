package main

import (
	"testing"
	"bytes"
//	"os"
	"github.com/zacg/testify/assert"
//"os"
)


func emitterEquals(t *testing.T, emitterOut chan string, expected ...string) {
	for _, value := range expected {
		assert.Equal(t, string(<-emitterOut), value)
	}
}

func TestParseXmlOneNode(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<hello>test</hello>")
	SaxReader(reader, 10, 10, "hello", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlOneNodeEmptySearch(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<hello>test</hello>")
	SaxReader(reader, 10, 10, "", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlNodeConstrainedBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<hello>test</hello>")
	SaxReader(reader, 1, 10, "hello", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlNodesConstrainedBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<helloC>C1</helloC>", "<helloC>C2</helloC>")
	SaxReader(reader, 1, 8, "helloA/helloB/helloC", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlNodesWithComments(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><!-- test<>--<><--><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<helloC>C1</helloC>", "<helloC>C2</helloC>")
	SaxReader(reader, 10, 10, "helloA/helloB/helloC", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlNodesWithCdata(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<helloC><![CDATA[Hello<! World!]]></helloC>", "<helloC>C2</helloC>")
	SaxReader(reader, 10, 10, "helloA/helloB/helloC", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlNodesWithCdataAndComment(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<helloC><![CDATA[Hello<! World!]]></helloC>", "<helloC>C2</helloC>")
	SaxReader(reader, 10, 10, "helloA/helloB/helloC", emitterChannel)
	close(emitterChannel)
}

func TestParseXmlNodesWithCdataAndCommentConstrainedBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><helloC><![CDATA[Hello<! World!]]></helloC><helloC>C2</helloC></helloB></helloA>"))
	emitterChannel := make(chan string)
	go emitterEquals(t, emitterChannel, "<helloC><![CDATA[Hello<! World!]]></helloC>", "<helloC>C2</helloC>")
	SaxReader(reader, 1, 10, "helloA/helloB/helloC", emitterChannel)
	close(emitterChannel)
}

//func TestParseWithEscapeAndCDATA(t *testing.T) {
//	reader,_ := os.Open("test.xml")
//	SaxReader(reader,1024,1024,"mediawiki")
//}