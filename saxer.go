package main

import (
	"fmt"
	"os"
	"path"
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/tcw/saxer/saxReader"
	"io"
	"bufio"
	"strings"
	"log"
	"runtime/pprof"
)

var (
	query = kingpin.Arg("query", "Sax query expression").Required().String()
	filename = kingpin.Arg("file", "xml-file").String()
	isInnerXml = kingpin.Flag("inner", "Inner-xml of selected element (default false)").Short('i').Default("false").Bool()
	count = kingpin.Flag("count", "Number of matches (default false)").Short('n').Default("false").Bool()
	contentBuffer = kingpin.Flag("cont-buf", "Size of content buffer in MB - returned elements size").Short('e').Default("4").Int()
	tagBuffer = kingpin.Flag("tag-buf", "Size of element tag buffer in KB - tag size").Short('t').Default("4").Int()
	firstN = kingpin.Flag("firstN", "First n matches (default (0 = all matches))").Short('f').Default("0").Int()
	cpuProfile = kingpin.Flag("profile", "Profile parser").Short('p').Bool()
)

const ONE_KB  int = 1024
const ONE_MB  int = ONE_KB * ONE_KB

func main() {
	kingpin.Version("0.0.4")
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
		message := <-emitter
		if len(message) != 0{
			fmt.Println(message)
		}
	}
}

func SaxXmlInput(reader io.Reader) {
	var err error
	var sr saxReader.SaxReader
	sr = saxReader.NewSaxReaderNoEmitter()
	sr.IsInnerXml = *isInnerXml
	sr.ContentBufferSize = *contentBuffer * ONE_MB
	sr.ElementBufferSize = *tagBuffer * ONE_KB
	if *count {
		var counter uint64 = 0
		emitterCounter := func(element string) bool {
			counter++
			return false
		};
		sr.EmitterFn = emitterCounter
		err = sr.Read(reader, *query)
		fmt.Println(counter)
	}else if *firstN > 0 {
		counter := 0
		elemChan := make(chan string, 100)
		go emitterPrinter(elemChan)
		emitter := func(element string) bool{
			elemChan <- element
			counter++
			if counter >= *firstN{
				close(elemChan)
				return true
			}else {
				return false
			}
		};

		sr.EmitterFn = emitter
		err = sr.Read(reader, *query)
		for{
			if len(elemChan) == 0{
				break
			}
		}
	}else {
		elemChan := make(chan string, 100)
		defer close(elemChan)
		go emitterPrinter(elemChan)
		emitter := func(element string) bool{
			elemChan <- element
			return false
		};
		sr.EmitterFn = emitter
		err = sr.Read(reader, *query)
		for{
			if len(elemChan) == 0{
				break
			}
		}
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
