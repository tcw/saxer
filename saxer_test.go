package main

import (
	"testing"
//	"github.com/zacg/testify/assert"
	"bytes"
	"os"
)

func TestParseXml(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>"))
	SaxReader(reader,10,10)
}

func TestParseXmlOverTwoBuffers(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>"))
	SaxReader(reader,5,10)
}

func TestParseXmlOverFullNode(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	SaxReader(reader,5,10)
}

func TestParseWithEscapeAndCDATA(t *testing.T) {
	reader,_ := os.Open("test.xml")
	SaxReader(reader,1024*4,1024*4)
}