package main

import (
	"testing"
	"bytes"
	"os"
)

func TestParseXmlOneNode(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	SaxReader(reader,10,10,"hello")
}

func TestParseXmlOneNodeEmptySearch(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	SaxReader(reader,10,10,"")
}

func TestParseXmlNodeConstrainedBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	SaxReader(reader,1,10,"hello")
}

func TestParseXmlNodesConstrainedBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	SaxReader(reader,1,8,"helloA/helloB")
}

func TestParseXmlNodesWithComments(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><!-- test<>--<><--><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	SaxReader(reader,10,10,"helloA/helloB")
}

func TestParseXmlNodesWithCdata(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><![CDATA[Hello<! World!]]><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	SaxReader(reader,10,10,"helloA/helloB")
}

func TestParseXmlNodesWithCdataAndComment(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><![CDATA[Hello<! World!]]><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	SaxReader(reader,10,10,"helloA/helloB")
}

func TestParseXmlNodesWithCdataAndCommentConstrainedBuffer(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><!-- test<>--<><--><helloB><![CDATA[Hello<! World!]]><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	SaxReader(reader,1,10,"helloA/helloB")
}

func TestParseWithEscapeAndCDATA(t *testing.T) {
	reader,_ := os.Open("test.xml")
	SaxReader(reader,1024,1024,"mediawiki")
}