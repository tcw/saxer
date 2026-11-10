// Package saxreader streams xml and reports the elements that match a
// query, without building a tree.
package saxreader

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/tcw/saxer/markup"
	"github.com/tcw/saxer/tagmatcher"
)

// SaxReader holds the settings for reading a document. Read can be called
// any number of times; all parsing state lives in a parser per call.
type SaxReader struct {
	ElementBufferSize int // largest tag, in bytes
	ContentBufferSize int // largest matched element, in bytes
	ReaderBufferSize  int // size of each read from the input
	EmitterFn         func(*Match) bool
	IsInnerXml        bool
	// SkipContent leaves Content and StartTag of a Match nil, so matches
	// are not copied. Use it when only counting or locating matches.
	SkipContent bool
}

// Match is an element that matched the query. Its byte slices are only
// valid during the call to EmitterFn; copy them to keep them.
type Match struct {
	Content   []byte // the element, or its inner xml with IsInnerXml
	StartTag  []byte // the start tag of the element, '<' to '>'
	Path      []byte // names of the element and its ancestors, joined by '/'
	LineStart uint64 // line of the '<' that starts the element
	LineEnd   uint64 // line of the '>' that ends the element
}

// Attributes parses the attributes of the matched element. Values are
// raw, entities are not decoded.
func (m *Match) Attributes() ([]markup.Attr, error) {
	return m.AppendAttributes(nil)
}

// AppendAttributes is Attributes appending to dst, to reuse its memory.
func (m *Match) AppendAttributes(dst []markup.Attr) ([]markup.Attr, error) {
	if len(m.StartTag) < 2 {
		return dst, nil
	}
	inner := bytes.TrimSuffix(m.StartTag[1:len(m.StartTag)-1], []byte("/"))
	_, rest := markup.SplitName(inner)
	return markup.ParseAttributes(rest, dst)
}

// Byte size units for the buffer sizes.
const (
	KB = 1024
	MB = 1024 * KB
)

var (
	ErrTagFull     = errors.New("tag is larger than the tag buffer, use --tag-buf to increase buffer")
	ErrContentFull = errors.New("content buffer is full, use --cont-buf to increase buffer")
)

// New returns a SaxReader with the default buffer sizes. Set EmitterFn
// before calling Read.
func New() SaxReader {
	return SaxReader{
		ElementBufferSize: 4 * KB,
		ContentBufferSize: 4 * MB,
		ReaderBufferSize:  64 * KB,
	}
}

// Read parses the xml from reader and calls EmitterFn for every element
// matching m, until the input ends or EmitterFn returns true.
func (sr *SaxReader) Read(reader io.Reader, m *tagmatcher.Matcher) error {
	p := &parser{
		sr:        sr,
		m:         m,
		r:         reader,
		buf:       make([]byte, max(sr.ReaderBufferSize, 1)),
		last:      m.Last(),
		textMode:  m.HasText(),
		keepBytes: !sr.SkipContent || m.HasText(),
	}
	err := p.run()
	var re *readError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &re):
		return fmt.Errorf("error reading xml: %w", re.err)
	default:
		return fmt.Errorf("error on line %d: %w", p.lineAt(p.pos), err)
	}
}

type readError struct{ err error }

func (e *readError) Error() string { return e.err.Error() }

// candidate is a matching element whose text is not known yet.
type candidate struct {
	depth      int
	start      int // offsets in content
	innerStart int
	lineStart  uint64
}

// pending is a candidate whose text matched, waiting for its enclosing
// candidates to end, since one of those may match instead.
type pending struct {
	start, innerStart, innerEnd, end int
	lineStart, lineEnd               uint64
	path                             []byte
}

// parser is the state of a single Read. The input is read into buf, which
// keeps the unparsed bytes buf[pos:end]; a tag that does not fit is kept
// and more input is read after it.
type parser struct {
	sr      *SaxReader
	m       *tagmatcher.Matcher
	r       io.Reader
	readErr error

	buf      []byte
	pos, end int
	eof      bool

	line     uint64 // newlines in the input before buf[lineScan]
	lineScan int

	// Open elements.
	path   []byte   // names joined by '/'
	starts []int    // start of each name in path
	states []uint64 // query steps each element satisfies

	last      uint64 // bit of the last query step
	textMode  bool   // the query matches on text, matches are decided at their end
	keepBytes bool   // copy the bytes of matches

	// A match outside text mode.
	inMatch    bool
	matchDepth int
	lineStart  uint64
	innerStart int

	// The bytes of the outermost match, from buf[recFrom:] when recording.
	recording bool
	recFrom   int
	content   []byte

	cands    []candidate
	pendings []pending

	match   Match
	stopped bool
}

func (p *parser) run() error {
	for {
		if p.pos == p.end {
			if p.eof {
				return p.finish()
			}
			if err := p.fill(); err != nil {
				return err
			}
			continue
		}
		i := bytes.IndexByte(p.buf[p.pos:p.end], '<')
		if i < 0 {
			p.pos = p.end
			continue
		}
		p.pos += i
		if err := p.markup(); err != nil {
			return err
		}
		if p.stopped {
			return nil
		}
	}
}

// finish checks that the document did not stop inside an element.
func (p *parser) finish() error {
	if n := len(p.starts); n > 0 {
		return fmt.Errorf("unexpected end of input, %d element(s) not closed", n)
	}
	return nil
}

// fill drops the parsed bytes before pos and reads more input after end.
func (p *parser) fill() error {
	if p.readErr != nil {
		return &readError{p.readErr}
	}
	keep := p.pos
	p.lineAt(keep)
	if p.recording {
		if err := p.flush(keep); err != nil {
			return err
		}
		p.recFrom -= keep
	}
	p.end = copy(p.buf, p.buf[keep:p.end])
	p.pos = 0
	p.lineScan -= keep
	if p.end == len(p.buf) {
		buf := make([]byte, 2*len(p.buf))
		copy(buf, p.buf[:p.end])
		p.buf = buf
	}
	for {
		n, err := p.r.Read(p.buf[p.end:])
		p.end += n
		switch {
		case err == io.EOF:
			p.eof = true
			return nil
		case err != nil && n > 0:
			p.readErr = err // report it once the bytes read are parsed
			return nil
		case err != nil:
			return &readError{err}
		case n > 0:
			return nil
		}
	}
}

// ensure reads until at least n bytes after pos are buffered, and reports
// whether they are.
func (p *parser) ensure(n int) (bool, error) {
	for p.end-p.pos < n && !p.eof {
		if err := p.fill(); err != nil {
			return false, err
		}
	}
	return p.end-p.pos >= n, nil
}

// lineAt returns the line number of buf[i].
func (p *parser) lineAt(i int) uint64 {
	if i < p.lineScan {
		return p.line + 1 - uint64(bytes.Count(p.buf[i:p.lineScan], []byte{'\n'}))
	}
	p.line += uint64(bytes.Count(p.buf[p.lineScan:i], []byte{'\n'}))
	p.lineScan = i
	return p.line + 1
}

// markup handles the markup starting with the '<' at pos.
func (p *parser) markup() error {
	ok, err := p.ensure(3)
	if err != nil {
		return err
	}
	if p.end-p.pos < 2 {
		return errors.New("unexpected end of input inside tag")
	}
	switch p.buf[p.pos+1] {
	case '?':
		return p.skipPast(p.pos+2, "?>", "processing instruction")
	case '/':
		return p.tag(true)
	case '!':
		if !ok {
			return errors.New("unexpected end of input inside declaration")
		}
		switch p.buf[p.pos+2] {
		case '-':
			return p.skipPast(p.pos+2, "-->", "comment")
		case '[':
			return p.skipPast(p.pos+2, "]]>", "CDATA section")
		}
		return p.declaration()
	}
	return p.tag(false)
}

// skipPast moves pos past the first term at or after from.
func (p *parser) skipPast(from int, term, what string) error {
	for {
		if i := bytes.Index(p.buf[from:p.end], []byte(term)); i >= 0 {
			p.pos = from + i + len(term)
			return nil
		}
		if p.eof {
			return fmt.Errorf("unexpected end of input inside %s", what)
		}
		// Keep the bytes that may start term.
		p.pos = max(from, p.end-len(term)+1)
		if err := p.fill(); err != nil {
			return err
		}
		from = p.pos
	}
}

// declaration skips a declaration such as <!DOCTYPE ...>, which may hold
// quoted strings and an internal subset in [ ].
func (p *parser) declaration() error {
	var quote byte
	depth := 0
	i := p.pos + 2
	for {
		for ; i < p.end; i++ {
			c := p.buf[i]
			switch {
			case quote != 0:
				if c == quote {
					quote = 0
				}
			case c == '"' || c == '\'':
				quote = c
			case c == '[':
				depth++
			case c == ']':
				depth--
			case c == '>' && depth <= 0:
				p.pos = i + 1
				return nil
			}
		}
		if p.eof {
			return errors.New("unexpected end of input inside declaration")
		}
		p.pos = i
		if err := p.fill(); err != nil {
			return err
		}
		i = p.pos
	}
}

// tag handles the start or end tag at pos.
func (p *parser) tag(isEnd bool) error {
	gt, err := p.scanTag()
	if err != nil {
		return err
	}
	if isEnd {
		err = p.endTag(p.pos, gt)
	} else {
		err = p.startTag(p.pos, gt)
	}
	if err != nil {
		return err
	}
	p.pos = gt + 1
	return nil
}

// scanTag reads until the '>' that ends the tag at pos and returns its
// index. A '>' in a quoted attribute value does not end the tag.
func (p *parser) scanTag() (int, error) {
	i := 1 // offset from pos
	var quote byte
	for {
		buf := p.buf[p.pos:p.end]
		for i < len(buf) {
			if quote != 0 {
				j := bytes.IndexByte(buf[i:], quote)
				if j < 0 {
					i = len(buf)
					break
				}
				i += j + 1
				quote = 0
				continue
			}
			switch buf[i] {
			case '>':
				if i > p.sr.ElementBufferSize {
					return 0, ErrTagFull
				}
				return p.pos + i, nil
			case '"', '\'':
				quote = buf[i]
			case '<':
				return 0, errors.New("found two '<' chars in a row")
			}
			i++
		}
		if i > p.sr.ElementBufferSize {
			return 0, ErrTagFull
		}
		if p.eof {
			return 0, errors.New("unexpected end of input inside tag")
		}
		if err := p.fill(); err != nil {
			return 0, err
		}
	}
}

// startTag handles the start tag buf[start:gt+1], which may be self-closing.
func (p *parser) startTag(start, gt int) error {
	inner := p.buf[start+1 : gt]
	selfClosing := len(inner) > 0 && inner[len(inner)-1] == '/'
	if selfClosing {
		inner = inner[:len(inner)-1]
	}
	name, rest := markup.SplitName(inner)
	if len(name) == 0 {
		if len(inner) == 0 && !selfClosing {
			return errors.New("found empty tag <>")
		}
		return errors.New("found tag without a name")
	}
	var state uint64
	if !p.inMatch { // inside a match nothing else is reported
		var parent uint64
		if n := len(p.states); n > 0 {
			parent = p.states[n-1]
		}
		var err error
		if state, err = p.m.Match(p.m.Next(parent), name, rest); err != nil {
			return err
		}
	}
	p.push(name, state)
	if state&p.last != 0 {
		p.beginMatch(start, gt)
	}
	if selfClosing {
		return p.closeElement(gt+1, gt)
	}
	return nil
}

// endTag handles the end tag buf[start:gt+1].
func (p *parser) endTag(start, gt int) error {
	name := p.buf[start+2 : gt]
	for len(name) > 0 && markup.IsSpace(name[len(name)-1]) {
		name = name[:len(name)-1]
	}
	n := len(p.starts)
	if n == 0 {
		return fmt.Errorf("found end tag </%s> before start tag", name)
	}
	if open := p.path[p.starts[n-1]:]; !bytes.Equal(name, open) {
		return fmt.Errorf("end tag </%s> does not match start tag <%s>", name, open)
	}
	return p.closeElement(start, gt)
}

func (p *parser) push(name []byte, state uint64) {
	if len(p.starts) > 0 {
		p.path = append(p.path, '/')
	}
	p.starts = append(p.starts, len(p.path))
	p.path = append(p.path, name...)
	p.states = append(p.states, state)
}

func (p *parser) pop() {
	n := len(p.starts) - 1
	s := p.starts[n]
	if n > 0 {
		s-- // the '/' before the name
	}
	p.path = p.path[:s]
	p.starts = p.starts[:n]
	p.states = p.states[:n]
}

// beginMatch starts a match at the start tag buf[start:gt+1].
func (p *parser) beginMatch(start, gt int) {
	if p.textMode {
		if !p.recording {
			p.startRecording(start)
		}
		p.cands = append(p.cands, candidate{
			depth:      len(p.starts),
			start:      p.offset(start),
			innerStart: p.offset(gt + 1),
			lineStart:  p.lineAt(start),
		})
		return
	}
	p.inMatch = true
	p.matchDepth = len(p.starts)
	p.lineStart = p.lineAt(start)
	if p.keepBytes {
		p.startRecording(start)
		p.innerStart = p.offset(gt + 1)
	}
}

// closeElement closes the innermost element, whose end tag is
// buf[endStart:gt+1]; for a self-closing tag endStart is gt+1.
func (p *parser) closeElement(endStart, gt int) error {
	depth := len(p.starts)
	var err error
	switch {
	case p.textMode:
		if n := len(p.cands); n > 0 && p.cands[n-1].depth == depth {
			err = p.endCandidate(endStart, gt)
		}
	case p.inMatch && depth == p.matchDepth:
		err = p.endMatch(endStart, gt)
	}
	p.pop()
	return err
}

func (p *parser) endMatch(endStart, gt int) error {
	p.inMatch = false
	p.match = Match{Path: p.path, LineStart: p.lineStart, LineEnd: p.lineAt(gt)}
	if p.recording {
		if err := p.flush(gt + 1); err != nil {
			return err
		}
		p.recording = false
		innerEnd := len(p.content) - (gt + 1 - endStart)
		p.match.StartTag = p.content[:p.innerStart]
		p.match.Content = p.content
		if p.sr.IsInnerXml {
			p.match.Content = p.content[p.innerStart:innerEnd]
		}
	}
	p.stopped = p.sr.EmitterFn(&p.match)
	return nil
}

func (p *parser) endCandidate(endStart, gt int) error {
	if err := p.flush(gt + 1); err != nil {
		return err
	}
	c := p.cands[len(p.cands)-1]
	p.cands = p.cands[:len(p.cands)-1]
	end := len(p.content)
	innerEnd := end - (gt + 1 - endStart)
	if p.m.TextMatches(p.content[c.innerStart:innerEnd]) {
		// This element replaces the matches inside it.
		k := len(p.pendings)
		for k > 0 && p.pendings[k-1].start >= c.start {
			k--
		}
		p.pendings = append(p.pendings[:k], pending{
			start: c.start, innerStart: c.innerStart, innerEnd: innerEnd, end: end,
			lineStart: c.lineStart, lineEnd: p.lineAt(gt),
			path: append([]byte(nil), p.path...),
		})
	}
	if len(p.cands) > 0 {
		return nil
	}
	p.recording = false
	for _, pd := range p.pendings {
		p.match = Match{Path: pd.path, LineStart: pd.lineStart, LineEnd: pd.lineEnd}
		if !p.sr.SkipContent {
			p.match.StartTag = p.content[pd.start:pd.innerStart]
			p.match.Content = p.content[pd.start:pd.end]
			if p.sr.IsInnerXml {
				p.match.Content = p.content[pd.innerStart:pd.innerEnd]
			}
		}
		if p.stopped = p.sr.EmitterFn(&p.match); p.stopped {
			break
		}
	}
	p.pendings = p.pendings[:0]
	return nil
}

func (p *parser) startRecording(start int) {
	p.recording = true
	p.recFrom = start
	p.content = p.content[:0]
}

// offset returns the offset in content that buf[i] has, or will have.
func (p *parser) offset(i int) int {
	return len(p.content) + i - p.recFrom
}

// flush copies the recorded bytes before buf[upto] to content.
func (p *parser) flush(upto int) error {
	if len(p.content)+upto-p.recFrom > p.sr.ContentBufferSize {
		return ErrContentFull
	}
	p.content = append(p.content, p.buf[p.recFrom:upto]...)
	p.recFrom = upto
	return nil
}
