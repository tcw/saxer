package main

import (
	"testing"
//	"github.com/zacg/testify/assert"
	"bytes"
)

func TestParseXml(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>"))
	startElement := NewStartElement(10)
	SaxReader(reader,10,startElement)
}

func TestParseXmlOverTwoBuffers(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>"))
	startElement := NewStartElement(10)
	SaxReader(reader,5,startElement)
}

func TestParseXmlOverFullNode(t *testing.T) {
	reader := bytes.NewReader([]byte("<hello>test</hello>"))
	startElement := NewStartElement(10)
	SaxReader(reader,5,startElement)
}