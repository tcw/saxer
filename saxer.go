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
	"github.com/tcw/saxer/contentBuffer"
	"github.com/tcw/saxer/nodePath"
	"bytes"
	"github.com/tcw/saxer/elementBuffer"
)

var (
	pathExp = kingpin.Arg("pathExp", "Sax Path Expression").Required().String()
	filename = kingpin.Arg("xml-file", "file").Required().String()
	cpuProfile = kingpin.Flag("profile", "Profile parser").Short('c').Bool()
)

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

func SaxFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	emitter := make(chan string, 1000)
	go emitterHandler(emitter)
	SaxReader(file, 1024 * 4, 1024 * 4, *pathExp, emitter)
}

func emitterHandler(emitter chan string) {
	for {
		fmt.Println(<-emitter)
	}
}

func SaxReader(reader io.Reader, bufferSize int, tmpNodeBufferSize int, pathQuery string, emitter chan string) {
	eb := elementBuffer.NewElementBuffer(tmpNodeBufferSize)
	buffer := make([]byte, bufferSize)
	inEscapeMode := false
	history := histBuffer.NewHistoryBuffer(tmpNodeBufferSize)
	contentBuffer := contentBuffer.NewContentBuffer(1024 * 1024, emitter)
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
		eb.ResetLocalState()
		for index := 0; index < n; index++ {
			value := buffer[index]
			if isRecoding {
				contentBuffer.Add(value)
			}
			if value == 0x0A {
				lineNumber++
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
			if value == byte('<') {
				eb.LocalStart = index
			}
			if value == byte('>') {
				eb.LocalEnd = index
			}
			if ((eb.LocalStart != -1 && index != 0 && eb.LocalStart == index - 1) && value == byte('!')) ||
			(index == 0 && eb.Position == 1 && value == byte('!')) {
				inEscapeMode = true
				eb.ResetState()
			}else if eb.LocalStart != -1 && eb.LocalEnd != -1 && eb.Position == 0 {
				isRecoding = ElementType(buffer[eb.LocalStart:eb.LocalEnd], &contentBuffer, &nodePath, isRecoding)
				eb.ResetLocalState()
			}else if eb.LocalEnd != -1 {
				eb.Add(buffer[:eb.LocalEnd])
				isRecoding = ElementType(eb.GetBuffer(), &contentBuffer, &nodePath, isRecoding)
				eb.ResetState()
			}
		}
		if eb.LocalStart == -1 && eb.LocalEnd == -1 && eb.Position > 0 {
			eb.Add(buffer)
		}else if eb.LocalStart != -1 {
			eb.Add(buffer[eb.LocalStart:n])
		}
	}
}

func ElementType(nodeContent []byte, nodeBuffer *contentBuffer.ContentBuffer, nodePath *nodePath.NodePath, isRecoding bool) bool {
	if nodeContent[1] == byte('/') {
		if isRecoding {
			if nodePath.MatchesLastMatch() {
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
		}else {
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
