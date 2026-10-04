package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

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
	if opts.cpuProfile {
		stop, err := startCPUProfile("cpu.pprof")
		if err != nil {
			return err
		}
		defer stop()
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

// startCPUProfile writes a CPU profile to path until the returned stop
// function is called. Inspect it with: go tool pprof --pdf saxer cpu.pprof
func startCPUProfile(path string) (stop func(), err error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, err
	}
	fmt.Fprintln(os.Stderr, "profiling!")
	return func() {
		pprof.StopCPUProfile()
		f.Close()
	}, nil
}

var htmlUnescaper = strings.NewReplacer(
	"&quot;", "\"",
	"&apos;", "'",
	"&lt;", "<",
	"&gt;", ">",
	"&amp;", "&")

// nodePrinter returns the function that writes one match to w, as
// selected by the output flags.
func nodePrinter(w io.Writer, opts *options) func(*contentbuffer.EmitterData) {
	if opts.meta {
		return func(ed *contentbuffer.EmitterData) {
			fmt.Fprintf(w, "%d-%d    %s\n", ed.LineStart, ed.LineEnd, ed.NodePath)
		}
	}
	return func(ed *contentbuffer.EmitterData) {
		node := ed.Content
		if opts.unescape {
			node = htmlUnescaper.Replace(node)
		}
		if opts.singleLine {
			node = strings.ReplaceAll(node, "\n", " ")
		}
		fmt.Fprintln(w, node)
	}
}

func newTagMatcher(opts *options) (tagmatcher.TagMatcher, error) {
	tm, err := tagmatcher.NewTagMatcher(opts.query)
	if err != nil {
		return tm, err
	}
	if opts.containMatch {
		tm.EqualityFn = tagmatcher.EqFnContains
	} else {
		tm.EqualityFn = tagmatcher.EqFnEquals
	}
	tm.CaseSensitive = !opts.caseInsensitive
	tm.WithoutNamespace = opts.omitNamespace
	return tm, nil
}

func SaxXmlInput(reader io.Reader, out io.Writer, opts *options) error {
	tm, err := newTagMatcher(opts)
	if err != nil {
		return err
	}
	sr := saxreader.NewSaxReaderNoEmitter()
	sr.IsInnerXml = opts.isInnerXml
	sr.ContentBufferSize = opts.contentBuf * saxreader.MB
	sr.ElementBufferSize = opts.tagBuffer * saxreader.KB

	w := bufio.NewWriter(out)
	if opts.wrapResult {
		fmt.Fprintln(w, "<saxer-result>")
	}
	var matches uint64
	printNode := nodePrinter(w, opts)
	sr.EmitterFn = func(ed *contentbuffer.EmitterData) bool {
		matches++
		if opts.count {
			return false
		}
		printNode(ed)
		return opts.firstN > 0 && matches >= uint64(opts.firstN)
	}
	err = sr.Read(reader, &tm)
	if opts.count {
		fmt.Fprintln(w, matches)
	}
	if opts.wrapResult {
		fmt.Fprintln(w, "</saxer-result>")
	}
	if flushErr := w.Flush(); err == nil {
		err = flushErr
	}
	return err
}
