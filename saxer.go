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


	readCount := 0
	inEscapeMode := false
	for {
		n, err := reader.Read(buffer)

		if n != 0 && err != nil {
			panic("Error while reading xml")
		}
		if n == 0 {
			break
		}
		elemStart := -1
		elemStop := -1

		for index, value := range buffer {
			if inEscapeMode {
				if value == byte('>'){
					if index > 1 {
						if buffer[index-1] == byte('-') && buffer[index-2] == byte('-'){
							fmt.Print(index)
							inEscapeMode = false
						}
					}
				}
				continue
			}
			if value == byte('<') {
				elemStart = index
			}
			if value == byte('>') {
				elemStop = index
			}
			if value == byte('!'){
				if elemStart == index -1{
					fmt.Print("Ecsape mode")
					inEscapeMode = true
					continue
				}
				if index == 0 && startElement.position > 0{
					//TODO forrige buffer
				}
			}

			if elemStop != -1 && elemStart == -1 && startElement.position > 0{
				copy(startElement.buffer[startElement.position:], buffer[:elemStop])
				startElement.position = 0
				elemStart = -1
				elemStop = -1
			}
			if elemStart != -1 && elemStop != -1{
				startElem(buffer[elemStart:elemStop])
				elemStart = -1
				elemStop = -1
			}
		}
		if elemStart != -1 && elemStop != -1 && startElement.position > 0 {
			copy(startElement.buffer[startElement.position:], buffer)
			startElement.position = startElement.position + n
		}

		if elemStart != -1  {
			copy(startElement.buffer, buffer[:n])
			startElement.position = startElement.position + n
		}
		if elemStop != -1  {
			copy(startElement.buffer[startElement.position:], buffer[:elemStop])
			startElement.position = 0
		}

		readCount = readCount + n
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
