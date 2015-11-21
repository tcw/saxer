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
	"github.com/tcw/saxer/nodeBuffer"
	"github.com/tcw/saxer/nodePath"
	"bytes"
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
	emitter := make(chan string)
	go emitterHandler(emitter)
	SaxReader(file, 1024*4, 1024*4, *pathExp,emitter)
}

func emitterHandler(emitter chan string)  {
	for{
		fmt.Println(<-emitter)
	}
}

func SaxReader(reader io.Reader, bufferSize int, tmpNodeBufferSize int, pathQuery string,emitter chan string) {
	startElement := NewStartElement(tmpNodeBufferSize)
	buffer := make([]byte, bufferSize)
	inEscapeMode := false
	history := histBuffer.NewHistoryBuffer(tmpNodeBufferSize)
	nodeBuffer := nodeBuffer.NewNodeBuffer(1024 * 1024,emitter)
	nodePath := nodePath.NewNodePath(1000, pathQuery)
	isRecoding := false
	var lineNumber uint64 = 0
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
		for index := 0; index < n; index++ {
			value := buffer[index]
			if isRecoding {
				nodeBuffer.Add(value)
			}
			if inEscapeMode {
				history.Add(value)
				if value == byte('>') {
					if history.HasLast([]byte{'-', '-', '>'}) {
						inEscapeMode = false
						continue
					}
					if history.HasLast([]byte{']', ']', '>'}) {
						inEscapeMode = false
						continue
					}
				}
				continue
			}
			if value == 0x0A {
				lineNumber++
			}
			if value == byte('<') {
				elemStart = index
			}
			if value == byte('>') {
				elemStop = index
			}
			if ((elemStart != -1 && index != 0 && elemStart == index - 1) && value == byte('!')) ||
			(index == 0 && startElement.position == 1 && value == byte('!')) {
				inEscapeMode = true
				startElement.position = 0
				elemStart = -1
			}else if elemStart != -1 && elemStop != -1 && startElement.position == 0 {
				if elemStart > elemStop {
					panic(fmt.Sprintf("Parsing error at line %d, possible syntax error", lineNumber))
				}
				isRecoding = ElementType(buffer[elemStart:elemStop], &nodeBuffer, &nodePath, isRecoding)
				elemStart = -1
				elemStop = -1
			}else if elemStop != -1 {
				copy(startElement.buffer[startElement.position:], buffer[:elemStop])
				startElement.position = startElement.position + elemStop
				isRecoding = ElementType(startElement.buffer[:startElement.position], &nodeBuffer, &nodePath, isRecoding)
				startElement.position = 0
				elemStop = -1
			}
		}
		if elemStart == -1 && elemStop == -1 && startElement.position > 0 {
			copy(startElement.buffer[startElement.position:], buffer)
			startElement.position = startElement.position + n
		}else if elemStart != -1 {
			copy(startElement.buffer, buffer[elemStart:n])
			startElement.position = startElement.position + (n - elemStart)
		}
	}
}

func ElementType(nodeContent []byte, nodeBuffer *nodeBuffer.NodeBuffer, nodePath *nodePath.NodePath, isRecoding bool) bool {
	if nodeContent[1] == byte('/') {
		if isRecoding {
			if nodePath.MatchesPath() {
				nodeBuffer.Emit()
				nodeBuffer.Reset()
				nodePath.RemoveLast()
				return false
			}else {
				nodePath.RemoveLast()
				return true
			}
		}
		nodePath.RemoveLast()
		return false
	}else if nodeContent[len(nodeContent) - 1] == byte('/') {
		nodePath.Add(getNodeName(nodeContent))
		if nodePath.MatchesPath() {
			if !isRecoding {
				nodeBuffer.AddArray(nodeContent)
				nodeBuffer.Add(byte('>'))
				nodeBuffer.Emit()
				nodeBuffer.Reset()
			}
		}
		nodePath.RemoveLast()
		return false
	}else {
		nodename := getNodeName(nodeContent)
		nodePath.Add(nodename)
		if !isRecoding {
			if nodePath.MatchesPath() {
				nodeBuffer.AddArray(nodeContent)
				nodeBuffer.Add(byte('>'))
				return true
			}else {
				return false
			}
		}else{
			return true
		}
	}
}

func getNodeName(nodeContent []byte) string {
	idx := bytes.IndexByte(nodeContent, byte(' '))
	if idx == -1 {
		return string(nodeContent[1:])
	}else {
		return string(nodeContent[1:idx])
	}
}

func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
