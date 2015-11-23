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
	pathExp = kingpin.Arg("pathExp", "Sax Path Expression").Required().String()
	filename = kingpin.Arg("xml-file", "file").String()
//	cpuProfile = kingpin.Flag("profile", "Profile parser").Short('p').Bool()
)

func main() {
	kingpin.Version("0.0.1")
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

	var reader io.Reader
	if strings.TrimSpace(*filename) != "" {
		absFilename, err := abs(*filename)
		if err != nil {
			panic(err.Error())
		}
		reader, err := os.Open(absFilename)
		if err != nil {
			panic(err.Error())
		}
		defer reader.Close()
	}else {
		reader = bufio.NewReader(os.Stdin)
	}
	SaxXmlInput(reader)
}

func emitterPrinter(emitter chan string) {
	for {
		fmt.Println(<-emitter)
	}
}

func SaxXmlInput(reader io.Reader) {
	elemChan := make(chan string, 100)
	defer close(elemChan)
	go emitterPrinter(elemChan)
	emitter := func(element string) {
		elemChan <- element
	};
	saxReader := saxReader.NewSaxReader(reader, emitter, *pathExp)
	saxReader.Read()
}


func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
