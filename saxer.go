package main

import (
	"bufio"
	"fmt"
	"github.com/tcw/saxer/contentBuffer"
	"github.com/tcw/saxer/saxReader"
	"github.com/tcw/saxer/tagMatcher"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"strings"
	"sync"
)

var (
	query         = kingpin.Arg("query", "Sax query expression").Required().String()
	filename      = kingpin.Arg("file", "xml-file").String()
	isInnerXml    = kingpin.Flag("inner", "Inner-xml of selected element (default false)").Short('i').Default("false").Bool()
	count         = kingpin.Flag("count", "Number of matches (default false)").Short('n').Default("false").Bool()
	meta          = kingpin.Flag("meta", "Get query meta data - linenumbers and path of matches (default false)").Short('m').Default("false").Bool()
	firstN        = kingpin.Flag("firstN", "First n matches (default (0 = all matches))").Short('f').Default("0").Int()
	unescape      = kingpin.Flag("unescape", "Unescape html escape tokens (&lt; &gt; ...)").Short('u').Default("false").Bool()
	caseSesitive  = kingpin.Flag("case", "Turn on case insensitivity").Short('s').Default("false").Bool()
	omitNamespace = kingpin.Flag("omit-ns", "Omit namespace in tag-name matches").Short('o').Default("false").Bool()
	containMatch  = kingpin.Flag("contains", "Maching of tag-name and attributes is executed by contains (not equals)").Short('c').Default("false").Bool()
	wrapResult    = kingpin.Flag("wrap", "Wrap result in Xml tag").Short('w').Default("false").Bool()
	singleLine    = kingpin.Flag("single-line", "Each node will have a single line (Changes line ending!)").Short('l').Default("false").Bool()
	tagBuffer     = kingpin.Flag("tag-buf", "Size of element tag buffer in KB - tag size").Default("4").Int()
	contentBuf    = kingpin.Flag("cont-buf", "Size of content buffer in MB - returned elements size").Default("4").Int()
	cpuProfile    = kingpin.Flag("profile-cpu", "Profile parser").Bool()
)

const ONE_KB int = 1024
const ONE_MB int = ONE_KB * ONE_KB

func main() {
	kingpin.Version("0.0.7")
	kingpin.Parse()

	//go tool pprof --pdf saxer cpu.pprof > callgraph.pdf
	//evince callgraph.pdf

	if *cpuProfile {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal(err)
		}
		perr := pprof.StartCPUProfile(f)
		if perr != nil {
			log.Fatal(perr)
		}
		fmt.Println("profiling!")
		defer pprof.StopCPUProfile()
	}

	if strings.TrimSpace(*filename) != "" {
		absFilename, err := abs(*filename)
		if err != nil {
			fmt.Printf("Error finding file: %s\n", absFilename)
		}
		file, err := os.Open(absFilename)
		if err != nil {
			fmt.Printf("Error opening file: %s\n", absFilename)
		}
		if file == nil {
			fmt.Printf("No file content found in file: %s\n", absFilename)
		}

		SaxXmlInput(file)
		ferr := file.Close()
		if ferr != nil {
			fmt.Printf("Error opening file: %s\n", absFilename)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		SaxXmlInput(reader)
	}
}

func emitterMetaPrinter(emitter chan contentBuffer.EmitterData, wg *sync.WaitGroup) {
	for {
		ed := <-emitter
		fmt.Printf("%d-%d    %s\n", ed.LineStart, ed.LineEnd, ed.NodePath)
		wg.Done()
	}
}

func emitterPrinter(emitter chan string, wg *sync.WaitGroup, line bool, htmlEscape bool) {
	r := strings.NewReplacer("&quot;", "\"",
		"&apos;", "'",
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&")
	for {
		node := <-emitter
		if htmlEscape {
			node = r.Replace(node)
		}
		if line {
			fmt.Println(strings.Replace(node, "\n", " ", -1))
		} else {
			fmt.Println(node)
		}
		wg.Done()
	}
}

func SaxXmlInput(reader io.Reader) {
	var err error
	var sr saxReader.SaxReader
	sr = saxReader.NewSaxReaderNoEmitter()
	tm := tagMatcher.NewTagMatcher(*query)
	if *containMatch {
		tm.EqualityFn = tagMatcher.EqFnContains
	} else {
		tm.EqualityFn = tagMatcher.EqFnEqulas
	}
	tm.CaseSensitive = !*caseSesitive
	tm.WithoutNamespace = *omitNamespace
	sr.IsInnerXml = *isInnerXml
	sr.ContentBufferSize = *contentBuf * ONE_MB
	sr.ElementBufferSize = *tagBuffer * ONE_KB
	if *wrapResult {
		fmt.Println("<saxer-result>")
	}
	if *count {
		var counter uint64 = 0
		emitterCounter := func(ed *contentBuffer.EmitterData) bool {
			counter++
			return false
		}
		sr.EmitterFn = emitterCounter
		err = sr.Read(reader, &tm)
		fmt.Println(counter)
	} else if *meta {
		counter := 0
		elemChan := make(chan contentBuffer.EmitterData, 100)
		var wg sync.WaitGroup
		go emitterMetaPrinter(elemChan, &wg)
		emitter := func(ed *contentBuffer.EmitterData) bool {
			wg.Add(1)
			elemChan <- contentBuffer.EmitterData{Content: ed.Content, LineStart: ed.LineStart, LineEnd: ed.LineEnd, NodePath: ed.NodePath}
			if *firstN > 0 {
				counter++
				if counter >= *firstN {
					return true
				} else {
					return false
				}
			}
			return false
		}
		sr.EmitterFn = emitter
		err = sr.Read(reader, &tm)
		wg.Wait()
	} else {
		counter := 0
		elemChan := make(chan string, 100)
		var wg sync.WaitGroup
		go emitterPrinter(elemChan, &wg, *singleLine, *unescape)
		emitter := func(ed *contentBuffer.EmitterData) bool {
			wg.Add(1)
			elemChan <- ed.Content
			if *firstN > 0 {
				counter++
				if counter >= *firstN {
					return true
				} else {
					return false
				}
			}
			return false
		}
		sr.EmitterFn = emitter
		err = sr.Read(reader, &tm)
		wg.Wait()
	}
	if *wrapResult {
		fmt.Println("</saxer-result>")
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
