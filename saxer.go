package main

import (
	"fmt"
	"os"
	"path"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"runtime/pprof"
	"github.com/tcw/saxer/saxReader"
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

func emitterPrinter(emitter chan string) {
	for {
		fmt.Println(<-emitter)
	}
}

func SaxFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	elemChan := make(chan string, 1000)
	go emitterPrinter(elemChan)
	emitter := func(element string) {
		elemChan <- element
	};
	saxReader := saxReader.NewSaxReader(file, emitter, *pathExp)
	saxReader.Read()
}


func abs(name string) (string, error) {
	if path.IsAbs(name) {
		return name, nil
	}
	wd, err := os.Getwd()
	return path.Join(wd, name), err
}
