package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tcw/saxer/contentbuffer"
	"github.com/tcw/saxer/saxreader"
	"github.com/tcw/saxer/tagmatcher"
)

const version = "0.0.7"

type options struct {
	query           string
	isInnerXml      bool
	count           bool
	meta            bool
	firstN          int
	unescape        bool
	caseInsensitive bool
	omitNamespace   bool
	containMatch    bool
	wrapResult      bool
	singleLine      bool
	tagBuffer       int
	contentBuf      int
	cpuProfile      bool
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:     "saxer [flags] <query> [file]",
		Short:   "Fast xml exploration tool for very large files",
		Long:    "Saxer queries xml with a subset of xpath. Reads from file, or from stdin when no file is given.",
		Version: version,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.query = args[0]
			filename := ""
			if len(args) == 2 {
				filename = args[1]
			}
			return run(opts, filename, cmd.InOrStdin(), cmd.OutOrStdout())
		},
		SilenceUsage: true,
	}

	f := cmd.Flags()
	f.SortFlags = false
	f.BoolVarP(&opts.isInnerXml, "inner", "i", false, "Inner-xml of selected element")
	f.BoolVarP(&opts.count, "count", "n", false, "Number of matches")
	f.BoolVarP(&opts.meta, "meta", "m", false, "Get query meta data - linenumbers and path of matches")
	f.IntVarP(&opts.firstN, "firstN", "f", 0, "First n matches (0 = all matches)")
	f.BoolVarP(&opts.unescape, "unescape", "u", false, "Unescape html escape tokens (&lt; &gt; ...)")
	f.BoolVarP(&opts.caseInsensitive, "case", "s", false, "Turn on case insensitivity")
	f.BoolVarP(&opts.omitNamespace, "omit-ns", "o", false, "Omit namespace in tag-name matches")
	f.BoolVarP(&opts.containMatch, "contains", "c", false, "Matching of tag-name and attributes is executed by contains (not equals)")
	f.BoolVarP(&opts.wrapResult, "wrap", "w", false, "Wrap result in Xml tag")
	f.BoolVarP(&opts.singleLine, "single-line", "l", false, "Each node will have a single line (Changes line ending!)")
	f.IntVar(&opts.tagBuffer, "tag-buf", 4, "Size of element tag buffer in KB - tag size")
	f.IntVar(&opts.contentBuf, "cont-buf", 4, "Size of content buffer in MB - returned elements size")
	f.BoolVar(&opts.cpuProfile, "profile-cpu", false, "Profile parser")
	return cmd
}

func run(opts *options, filename string, in io.Reader, out io.Writer) error {
	//go tool pprof --pdf saxer cpu.pprof > callgraph.pdf
	//evince callgraph.pdf

	if opts.cpuProfile {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		fmt.Fprintln(out, "profiling!")
		defer pprof.StopCPUProfile()
	}

	if strings.TrimSpace(filename) != "" {
		absFilename, err := filepath.Abs(filename)
		if err != nil {
			return fmt.Errorf("error finding file %s: %w", filename, err)
		}
		file, err := os.Open(absFilename)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()
		return SaxXmlInput(file, out, opts)
	}
	return SaxXmlInput(bufio.NewReader(in), out, opts)
}

func emitterMetaPrinter(out io.Writer, emitter chan contentbuffer.EmitterData, wg *sync.WaitGroup) {
	for ed := range emitter {
		fmt.Fprintf(out, "%d-%d    %s\n", ed.LineStart, ed.LineEnd, ed.NodePath)
		wg.Done()
	}
}

func emitterPrinter(out io.Writer, emitter chan string, wg *sync.WaitGroup, line bool, htmlEscape bool) {
	r := strings.NewReplacer("&quot;", "\"",
		"&apos;", "'",
		"&lt;", "<",
		"&gt;", ">",
		"&amp;", "&")
	for node := range emitter {
		if htmlEscape {
			node = r.Replace(node)
		}
		if line {
			fmt.Fprintln(out, strings.ReplaceAll(node, "\n", " "))
		} else {
			fmt.Fprintln(out, node)
		}
		wg.Done()
	}
}

func SaxXmlInput(reader io.Reader, out io.Writer, opts *options) error {
	var err error
	var sr saxreader.SaxReader
	sr = saxreader.NewSaxReaderNoEmitter()
	tm, err := tagmatcher.NewTagMatcher(opts.query)
	if err != nil {
		return err
	}
	if opts.containMatch {
		tm.EqualityFn = tagmatcher.EqFnContains
	} else {
		tm.EqualityFn = tagmatcher.EqFnEquals
	}
	tm.CaseSensitive = !opts.caseInsensitive
	tm.WithoutNamespace = opts.omitNamespace
	sr.IsInnerXml = opts.isInnerXml
	sr.ContentBufferSize = opts.contentBuf * saxreader.MB
	sr.ElementBufferSize = opts.tagBuffer * saxreader.KB
	if opts.wrapResult {
		fmt.Fprintln(out, "<saxer-result>")
	}
	if opts.count {
		var counter uint64 = 0
		emitterCounter := func(ed *contentbuffer.EmitterData) bool {
			counter++
			return false
		}
		sr.EmitterFn = emitterCounter
		err = sr.Read(reader, &tm)
		fmt.Fprintln(out, counter)
	} else if opts.meta {
		counter := 0
		elemChan := make(chan contentbuffer.EmitterData, 100)
		var wg sync.WaitGroup
		go emitterMetaPrinter(out, elemChan, &wg)
		emitter := func(ed *contentbuffer.EmitterData) bool {
			wg.Add(1)
			elemChan <- contentbuffer.EmitterData{Content: ed.Content, LineStart: ed.LineStart, LineEnd: ed.LineEnd, NodePath: ed.NodePath}
			if opts.firstN > 0 {
				counter++
				if counter >= opts.firstN {
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
		close(elemChan)
	} else {
		counter := 0
		elemChan := make(chan string, 100)
		var wg sync.WaitGroup
		go emitterPrinter(out, elemChan, &wg, opts.singleLine, opts.unescape)
		emitter := func(ed *contentbuffer.EmitterData) bool {
			wg.Add(1)
			elemChan <- ed.Content
			if opts.firstN > 0 {
				counter++
				if counter >= opts.firstN {
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
		close(elemChan)
	}
	if opts.wrapResult {
		fmt.Fprintln(out, "</saxer-result>")
	}
	return err
}
