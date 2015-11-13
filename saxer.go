package main

import (
	"fmt"
	"os"
	"path"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
)

var (
	pathExp = kingpin.Arg("pathExp", "Sax Path Expression").Required().String()
	filename = kingpin.Arg("xml-file", "file").Required().String()
)

type StartElement struct {
	buffer   []byte
	position int
}

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	fmt.Println(*pathExp)

	absFilename, err := abs(*filename)
	if err != nil {
		panic(err.Error())
	}
	SaxFile(absFilename)
}

func NewStartElement(bufferSize int) StartElement {
	return StartElement{buffer: make([]byte, bufferSize), position: 0}
}

func SaxFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	startElement := NewStartElement(1024 * 4)
	SaxReader(file, 1024 * 4, startElement)
}

func SaxReader(reader io.Reader, bufferSize int, startElement StartElement) {
	buffer := make([]byte, bufferSize)
	startElemFrom := -1
	startElemTo := -1

	var readCount int = 0
	for {
		n, err := reader.Read(buffer)

		if n != 0 && err != nil {
			panic("Error while reading xml")
		}
		if n == 0 {
			break
		}
		for index, value := range buffer {
			if value == byte('<') {
				startElemFrom = index
			}
			if value == byte('>') {
				startElemTo = index
			}
		}
		readCount = readCount + n
		if startElemFrom != -1 && startElemTo == -1 &&startElement.position > 0 {
			copy(startElement.buffer[startElement.position:], buffer[:n])
			startElement.position = startElement.position + n
		}
		if startElemFrom != -1 && startElemTo == -1 &&startElement.position == 0 {
			copy(startElement.buffer, buffer[startElemFrom:n])
			startElement.position = startElement.position + n
		}
		if startElemFrom != -1 && startElemTo != -1 && startElement.position == 0{
			startElem(buffer[startElemFrom:startElemTo])
			startElemFrom = -1
			startElemTo = -1
		}
		if startElemFrom != -1 && startElemTo != -1 && startElement.position > 0 {
			copy(startElement.buffer[startElement.position:], buffer[:startElemTo])
			startElemFrom = -1
			startElemTo = -1
			startElem(startElement.buffer)
			startElement.position = 0
		}
	}
	fmt.Println(readCount)
}

func startElem(bytes []byte) {
	fmt.Println(string(bytes))
}

func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
