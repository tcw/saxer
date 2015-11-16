package main

import (
	"fmt"
	"os"
	"path"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"runtime/pprof"
	"github.com/tcw/saxer/histBuffer"
)

var (
	pathExp = kingpin.Arg("pathExp", "Sax Path Expression").Required().String()
	filename = kingpin.Arg("xml-file", "file").Required().String()
	cpuProfile = kingpin.Flag("profile", "Profile parser").Short('c').Bool()
)

type StartElement struct {
	buffer   []byte
	position int
}

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	fmt.Println(*pathExp)

	//go tool pprof --pdf saxer cpu.pprof > callgraph.pdf
	//evince callgraph.pdf
	if *cpuProfile {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		fmt.Println("profiling!")
		defer pprof.StopCPUProfile()
	}

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

	SaxReader(file, 1024 * 4, 1024 * 4)
}

func SaxReader(reader io.Reader, bufferSize int, tmpNodeBufferSize int) {
	startElement := NewStartElement(tmpNodeBufferSize)
	buffer := make([]byte, bufferSize)
	readCount := 0
	inEscapeMode := false
	history := histBuffer.NewHistoryBuffer(tmpNodeBufferSize)
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
				history.Add(value)
				if value == byte('>') {
					if history.HasLast([]byte{'-', '-', '>'}) {
						fmt.Print("commment",index)
						inEscapeMode = false
						continue
					}
					if history.HasLast([]byte{']', ']', '>'}) {
						fmt.Print("CDATA",index)
						inEscapeMode = false
						continue
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
			if (elemStart == index - 1 && value == byte('!')) || (index == 0 && startElement.position == 1 && value == byte('!')){
				inEscapeMode = true
				startElement.position = 0
				elemStart = -1
				continue
			}
			if elemStop != -1 && elemStart == -1 && startElement.position > 0 {
				copy(startElement.buffer[startElement.position:], buffer[:elemStop])
				startElement.position = 0
				elemStart = -1
				elemStop = -1
			}
			if elemStart != -1 && elemStop != -1 {
				ElementType(buffer[elemStart:elemStop])
				elemStart = -1
				elemStop = -1
			}
		}
		if elemStart != -1 && elemStop != -1 && startElement.position > 0 {
			copy(startElement.buffer[startElement.position:], buffer)
			startElement.position = startElement.position + n
		}
		if elemStart != -1 {
			copy(startElement.buffer, buffer[:n])
			startElement.position = startElement.position + n
		}
		if elemStop != -1 {
			copy(startElement.buffer[startElement.position:], buffer[:elemStop])
			startElement.position = startElement.position + n
			ElementType(startElement.buffer[:startElement.position])
			startElement.position = 0
		}
		readCount = readCount + n
	}
	fmt.Println(readCount)
}

func ElementType(bytes []byte) {
	if bytes[1] == byte('/') {
		//End node
//		fmt.Println("End: ", string(bytes))
	}else if bytes[len(bytes) - 1] == byte('/') {
		//Start and end node
//		fmt.Println("Start and end: ", string(bytes))
	}else {
		//Start node
//		fmt.Println("Start: ", string(bytes))
	}
}

func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
