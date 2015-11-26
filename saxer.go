package main

import (
	"fmt"
	"os"
	"path"
	"gopkg.in/alecthomas/kingpin.v2"
//	"log"
//	"runtime/pprof"
	"github.com/tcw/saxer/saxReader"
	"io"
	"bufio"
	"strings"
)

var (
	query = kingpin.Arg("query", "Sax query expression").Required().String()
	filename = kingpin.Arg("file", "xml-file").String()
	isInnerXml = kingpin.Flag("inner", "Inner-xml of selected element (default false)").Short('i').Default("false").Bool()
	htmlConv = kingpin.Flag("htmlconv", "Converting html escape to ascii (default false)").Short('c').Default("false").Bool()
	count = kingpin.Flag("count", "Number of matches (default false)").Short('n').Default("false").Bool()
	contentBuffer = kingpin.Flag("cont-buf", "Size of content buffer in MB - returned elements size").Short('e').Default("4").Int()
	tagBuffer = kingpin.Flag("tag-buf", "Size of element tag buffer in KB - tag size").Short('t').Default("4").Int()

//	cpuProfile = kingpin.Flag("profile", "Profile parser").Short('p').Bool()
)

const ONE_KB  int = 1024
const ONE_MB  int = ONE_KB * ONE_KB

func main() {
	kingpin.Version("0.0.2")
	kingpin.Parse()

	//go tool pprof --pdf saxer cpu.pprof > callgraph.pdf
	//evince callgraph.pdf

	//	if *cpuProfile {
	//		f, err := os.Create("cpu.pprof")
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		pprof.StartCPUProfile(f)
	//		fmt.Println("profiling!")
	//		defer pprof.StopCPUProfile()
	//	}

	if strings.TrimSpace(*filename) != "" {
		absFilename, err := abs(*filename)
		if err != nil {
			panic(err.Error())
		}
		file, err := os.Open(absFilename)
		if err != nil {
			panic(err.Error())
		}
		defer file.Close()
		SaxXmlInput(file)
	}else {
		reader := bufio.NewReader(os.Stdin)
		SaxXmlInput(reader)
	}
}

func emitterPrinter(emitter chan string) {
	for {
		fmt.Println(<-emitter)
	}
}

func SaxXmlInput(reader io.Reader) {
	var err error
	var sr saxReader.SaxReader
	sr = saxReader.NewSaxReaderNoEmitter()
	sr.IsInnerXml = *isInnerXml
	sr.FilterEscapeSigns = *htmlConv
	sr.ContentBufferSize = *contentBuffer * ONE_MB
	sr.ElementBufferSize = *tagBuffer * ONE_KB
	if *count {
		var counter uint64 = 0
		emitterCounter := func(element string) {
			counter++
		};
		sr.EmitterFn = emitterCounter
		err = sr.Read(reader, *query)
		fmt.Println(counter)
	}else {
		elemChan := make(chan string, 100)
		defer close(elemChan)
		go emitterPrinter(elemChan)
		emitter := func(element string) {
			elemChan <- element
		};
		sr.EmitterFn = emitter
		err = sr.Read(reader, *query)
	}
	if err != nil {
		panic(err)
	}
}


func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
