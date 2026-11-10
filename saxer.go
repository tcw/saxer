package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/tcw/saxer/markup"
	"github.com/tcw/saxer/saxreader"
	"github.com/tcw/saxer/tagmatcher"
)

// Build information, set by release builds with
// -ldflags "-X main.version=v1.2.3 -X main.commit=... -X main.date=...".
// When version is empty it is taken from the Go build information.
var (
	version = ""
	commit  = ""
	date    = ""
)

// Exit codes, as for grep.
const (
	exitMatch   = 0
	exitNoMatch = 1
	exitError   = 2
)

// errNoMatch is returned by the command when nothing matched.
var errNoMatch = errors.New("no matches")

type options struct {
	query           string
	isInnerXml      bool
	count           bool
	meta            bool
	json            bool
	firstN          int
	unescape        bool
	caseInsensitive bool
	omitNamespace   bool
	containMatch    bool
	wrapResult      bool
	singleLine      bool
	squash          bool
	tagBuffer       int
	contentBuf      int
	cpuProfile      bool
}

func main() {
	os.Exit(execute(newRootCmd(), os.Stderr))
}

// execute runs cmd and returns the exit code: 0 when something matched,
// 1 when nothing did and 2 on errors.
func execute(cmd *cobra.Command, stderr io.Writer) int {
	err := cmd.Execute()
	switch {
	case err == nil:
		return exitMatch
	case errors.Is(err, errNoMatch):
		return exitNoMatch
	default:
		fmt.Fprintln(stderr, "Error:", err)
		return exitError
	}
}

func newRootCmd() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "saxer [flags] <query> [file]",
		Short: "Fast xml exploration tool for very large files",
		Long: "Saxer queries xml with a subset of xpath, streaming, in little memory.\n" +
			"Reads from file, or from stdin when no file is given. Gzip input is detected and decompressed.\n\n" +
			"Query: step/step/..., each step name[~text][?key[=value]&...]; '*' matches any element.\n\n" +
			"Exit status is 0 when something matched, 1 when nothing did and 2 on errors.",
		Version: buildVersion(),
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.query = args[0]
			filename := ""
			if len(args) == 2 {
				filename = args[1]
			}
			return run(opts, filename, cmd.InOrStdin(), cmd.OutOrStdout())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetVersionTemplate(versionInfo() + "\n")

	f := cmd.Flags()
	f.SortFlags = false
	// Matching.
	f.BoolVarP(&opts.caseInsensitive, "case", "s", false, "Turn on case insensitivity")
	f.BoolVarP(&opts.omitNamespace, "omit-ns", "o", false, "Omit namespace in tag-name matches")
	f.BoolVarP(&opts.containMatch, "contains", "c", false, "Match tag names, attributes and text by contains (not equals)")
	f.IntVarP(&opts.firstN, "firstN", "f", 0, "First n matches (0 = all matches)")
	// Output.
	f.BoolVarP(&opts.isInnerXml, "inner", "i", false, "Inner-xml of selected element")
	f.BoolVarP(&opts.count, "count", "n", false, "Number of matches")
	f.BoolVarP(&opts.meta, "meta", "m", false, "Line numbers and path of matches, tab separated")
	f.BoolVarP(&opts.json, "json", "j", false, "One JSON object per match and line: content, line_start, line_end, path, attributes")
	f.BoolVarP(&opts.unescape, "unescape", "u", false, "Unescape xml escape tokens (&lt; &gt; &#38; ...)")
	f.BoolVarP(&opts.singleLine, "single-line", "l", false, "Each node will have a single line (line endings become spaces)")
	f.BoolVar(&opts.squash, "squash", false, "Squash runs of white space into one space (implies --single-line)")
	f.BoolVarP(&opts.wrapResult, "wrap", "w", false, "Wrap result in Xml tag")
	// Buffers.
	f.IntVar(&opts.tagBuffer, "tag-buf", 4, "Size of element tag buffer in KB - largest tag")
	f.IntVar(&opts.contentBuf, "cont-buf", 4, "Size of content buffer in MB - largest returned element")
	f.BoolVar(&opts.cpuProfile, "profile-cpu", false, "Write a CPU profile to cpu.pprof")
	return cmd
}

// versionInfo returns the version line printed by --version, with the
// commit and build date when they are known.
func versionInfo() string {
	s := "saxer " + buildVersion()
	if commit != "" {
		s += "\ncommit: " + commit
	}
	if date != "" {
		s += "\nbuilt: " + date
	}
	return s
}

// buildVersion returns the version of the module from the build
// information, or the VCS revision for a development build.
func buildVersion() string {
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "devel"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	revision, modified := "", false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if revision == "" {
		return "devel"
	}
	v := "devel-" + revision[:min(12, len(revision))]
	if modified {
		v += "-dirty"
	}
	return v
}

func run(opts *options, filename string, in io.Reader, out io.Writer) error {
	if opts.json && (opts.meta || opts.wrapResult) {
		return errors.New("--json cannot be combined with --meta or --wrap")
	}
	if opts.cpuProfile {
		stop, err := startCPUProfile("cpu.pprof")
		if err != nil {
			return err
		}
		defer stop()
	}

	if filename != "" {
		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer file.Close()
		in = file
	}
	input, err := decompress(in)
	if err != nil {
		return err
	}
	matches, err := SaxXmlInput(input, out, opts)
	if err == nil && matches == 0 {
		err = errNoMatch
	}
	return err
}

// decompress returns r, decompressed when it starts with the gzip magic bytes.
func decompress(r io.Reader) (io.Reader, error) {
	var head [2]byte
	n, err := io.ReadFull(r, head[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("error reading xml: %w", err)
	}
	r = io.MultiReader(bytes.NewReader(head[:n]), r)
	if n < 2 || head[0] != 0x1f || head[1] != 0x8b {
		return r, nil
	}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("error reading gzip input: %w", err)
	}
	return zr, nil
}

// startCPUProfile writes a CPU profile to path until the returned stop
// function is called. Inspect it with: go tool pprof saxer cpu.pprof
func startCPUProfile(path string) (stop func(), err error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, err
	}
	fmt.Fprintln(os.Stderr, "profiling to", path)
	return func() {
		pprof.StopCPUProfile()
		f.Close()
	}, nil
}

// printer writes matches in the format selected by the options.
type printer struct {
	w       *bufio.Writer
	opts    *options
	scratch []byte
	text    []byte
	value   []byte
	attrs   []markup.Attr
}

func (p *printer) print(m *saxreader.Match) {
	switch {
	case p.opts.meta:
		b := strconv.AppendUint(p.scratch[:0], m.LineStart, 10)
		b = append(b, '-')
		b = strconv.AppendUint(b, m.LineEnd, 10)
		b = append(b, '\t')
		b = append(b, m.Path...)
		p.scratch = append(b, '\n')
		p.w.Write(p.scratch)
	case p.opts.json:
		p.printJSON(m)
	default:
		p.w.Write(p.content(m.Content))
		p.w.WriteByte('\n')
	}
}

// content applies the unescape and single line options to an element.
func (p *printer) content(b []byte) []byte {
	if p.opts.unescape && bytes.IndexByte(b, '&') >= 0 {
		p.text = markup.AppendUnescaped(p.text[:0], b)
		b = p.text
	}
	switch {
	case p.opts.squash:
		p.scratch = squash(p.scratch[:0], b)
		b = p.scratch
	case p.opts.singleLine:
		p.scratch = singleLine(p.scratch[:0], b)
		b = p.scratch
	}
	return b
}

// singleLine appends b to dst with every line ending, \n, \r\n or \r,
// replaced by a space.
func singleLine(dst, b []byte) []byte {
	for {
		i := bytes.IndexAny(b, "\r\n")
		if i < 0 {
			return append(dst, b...)
		}
		dst = append(dst, b[:i]...)
		dst = append(dst, ' ')
		if b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n' {
			i++
		}
		b = b[i+1:]
	}
}

// squash appends b to dst with every run of white space replaced by one
// space, and white space at both ends removed.
func squash(dst, b []byte) []byte {
	b = bytes.TrimSpace(b)
	space := false
	for _, c := range b {
		if markup.IsSpace(c) {
			space = true
			continue
		}
		if space {
			dst = append(dst, ' ')
			space = false
		}
		dst = append(dst, c)
	}
	return dst
}

func (p *printer) printJSON(m *saxreader.Match) {
	b := append(p.value[:0], `{"content":`...)
	b = appendJSONString(b, p.content(m.Content))
	b = append(b, `,"line_start":`...)
	b = strconv.AppendUint(b, m.LineStart, 10)
	b = append(b, `,"line_end":`...)
	b = strconv.AppendUint(b, m.LineEnd, 10)
	b = append(b, `,"path":`...)
	b = appendJSONString(b, m.Path)
	b = append(b, `,"attributes":{`...)
	p.attrs, _ = m.AppendAttributes(p.attrs[:0]) // already checked when the element matched
	for i, a := range p.attrs {
		if i > 0 {
			b = append(b, ',')
		}
		b = appendJSONString(b, a.Key)
		b = append(b, ':')
		p.text = markup.AppendUnescaped(p.text[:0], a.Value)
		b = appendJSONString(b, p.text)
	}
	p.value = append(b, "}}\n"...)
	p.w.Write(p.value)
}

// appendJSONString appends s as a JSON string. Invalid UTF-8 becomes U+FFFD.
func appendJSONString(dst, s []byte) []byte {
	const hex = "0123456789abcdef"
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c >= 0x20 && c != '"' && c != '\\' && c < utf8.RuneSelf {
			i++
			continue
		}
		if c >= utf8.RuneSelf {
			r, size := utf8.DecodeRune(s[i:])
			if r != utf8.RuneError || size != 1 {
				i += size
				continue
			}
			dst = append(dst, s[start:i]...)
			dst = append(dst, `�`...)
			i++
			start = i
			continue
		}
		dst = append(dst, s[start:i]...)
		switch c {
		case '"', '\\':
			dst = append(dst, '\\', c)
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
		}
		i++
		start = i
	}
	dst = append(dst, s[start:]...)
	return append(dst, '"')
}

// SaxXmlInput writes the matches of opts.query in reader to out and
// returns the number of matches.
func SaxXmlInput(reader io.Reader, out io.Writer, opts *options) (uint64, error) {
	tm, err := tagmatcher.New(opts.query, tagmatcher.Options{
		Contains:        opts.containMatch,
		CaseInsensitive: opts.caseInsensitive,
		OmitNamespace:   opts.omitNamespace,
	})
	if err != nil {
		return 0, err
	}
	sr := saxreader.New()
	sr.IsInnerXml = opts.isInnerXml
	sr.ContentBufferSize = opts.contentBuf * saxreader.MB
	sr.ElementBufferSize = opts.tagBuffer * saxreader.KB
	sr.SkipContent = opts.count || opts.meta

	w := bufio.NewWriterSize(out, 64*saxreader.KB)
	if opts.wrapResult {
		w.WriteString("<saxer-result>\n")
	}
	var matches uint64
	p := &printer{w: w, opts: opts}
	sr.EmitterFn = func(m *saxreader.Match) bool {
		matches++
		if opts.count {
			return false
		}
		p.print(m)
		return opts.firstN > 0 && matches >= uint64(opts.firstN)
	}
	err = sr.Read(reader, tm)
	if opts.count {
		fmt.Fprintln(w, matches)
	}
	if opts.wrapResult {
		w.WriteString("</saxer-result>\n")
	}
	if flushErr := w.Flush(); err == nil {
		err = flushErr
	}
	return matches, err
}
