package main

import (
	"testing"
//	"github.com/zacg/testify/assert"
	"bytes"
	"os"
)

func TestParseXml(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	SaxReader(reader,10,10,"hello")
}

func TestParseXmlOverFullNode(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	SaxReader(reader,4,10,"hello")
}

func TestParseXmlOverFullNoder(t *testing.T) {
	reader := bytes.NewReader([]byte("<helloA><helloB><helloC>C1</helloC><helloC>C2</helloC></helloB></helloA>"))
	SaxReader(reader,10,10,"helloA/helloB")
}

func TestParseWithEscapeAndCDATA(t *testing.T) {
	reader,_ := os.Open("test.xml")
	SaxReader(reader,1024,1024,"mediawiki")
}