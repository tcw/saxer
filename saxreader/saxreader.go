package saxreader

import (
	"errors"
	"fmt"
	"io"

	"github.com/tcw/saxer/contentbuffer"
	"github.com/tcw/saxer/tagbuffer"
	"github.com/tcw/saxer/tagmatcher"
)

// SaxReader holds the settings for reading a document. Read can be called
// any number of times; all parsing state lives in a parser per call.
type SaxReader struct {
	ElementBufferSize int // largest tag, in bytes
	ContentBufferSize int // largest matched element, in bytes
	ReaderBufferSize  int // size of each read from the input
	EmitterFn         func(*contentbuffer.EmitterData) bool
	IsInnerXml        bool
}

// Byte size units for the buffer sizes.
const (
	KB = 1024
	MB = 1024 * KB
)

// New returns a SaxReader with the default buffer sizes. Set EmitterFn
// before calling Read.
func New() SaxReader {
	return SaxReader{
		ElementBufferSize: 4 * KB,
		ContentBufferSize: 4 * MB,
		ReaderBufferSize:  4 * KB,
	}
}

// Read parses the xml from reader and calls EmitterFn for every element
// matching tm, until the input ends or EmitterFn returns true.
func (sr *SaxReader) Read(reader io.Reader, tm *tagmatcher.TagMatcher) error {
	p := &parser{
		matcher:  tm,
		tag:      tagbuffer.NewTagBuffer(sr.ElementBufferSize),
		content:  contentbuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn),
		innerXml: sr.IsInnerXml,
	}
	buffer := make([]byte, sr.ReaderBufferSize)
	for {
		n, readErr := reader.Read(buffer)
		for _, b := range buffer[:n] {
			if err := p.next(b); err != nil {
				return fmt.Errorf("error on line %d: %w", p.line+1, err)
			}
			if p.stopped {
				return nil
			}
		}
		if readErr == io.EOF {
			return p.end()
		}
		if readErr != nil {
			return fmt.Errorf("error reading xml: %w", readErr)
		}
	}
}

// Kinds of markup that is skipped until its terminator: <!-- -->, <![CDATA[ ]]>,
// <? ?> and declarations such as <!DOCTYPE ...>.
type markup int

const (
	markupNone        markup = iota
	markupBang               // seen "<!", kind not known yet
	markupComment            // <!-- -->
	markupCdata              // <![CDATA[ ]]>
	markupPI                 // <? ?>
	markupDeclaration        // <!DOCTYPE ...>, may contain [ ] and quoted strings
)

var markupNames = map[markup]string{
	markupBang:        "declaration",
	markupComment:     "comment",
	markupCdata:       "CDATA section",
	markupPI:          "processing instruction",
	markupDeclaration: "declaration",
}

// parser is the state of a single Read. It is fed one byte at a time, so
// tags, markup and quoted values may span any number of reads.
type parser struct {
	matcher  *tagmatcher.TagMatcher
	tag      tagbuffer.TagBuffer // the tag being read, from '<' up to but excluding '>'
	content  contentbuffer.ContentBuffer
	match    contentbuffer.EmitterData
	innerXml bool

	line      uint64 // newlines seen so far
	depth     int    // open elements
	inTag     bool
	inMarkup  markup
	quote     byte    // quote char of the value being read in a tag or declaration, or 0
	declDepth int     // nesting of [ ] in a declaration (DOCTYPE internal subset)
	tail      [2]byte // last two bytes of the current markup, to find its terminator
	recording bool    // inside a matched element, its bytes go to content
	stopped   bool    // the emitter asked to stop
}

func (p *parser) next(b byte) error {
	if p.recording {
		if err := p.content.Add(b); err != nil {
			return err
		}
	}
	if b == '\n' {
		p.line++
	}
	switch {
	case p.inMarkup != markupNone:
		p.markupByte(b)
	case p.inTag:
		return p.tagByte(b)
	case b == '<':
		p.inTag = true
		p.tag.Reset()
		return p.tag.Add(b)
	}
	return nil
}

// end checks that the document did not stop in the middle of something.
func (p *parser) end() error {
	switch {
	case p.inMarkup != markupNone:
		return fmt.Errorf("unexpected end of input inside %s", markupNames[p.inMarkup])
	case p.inTag:
		return errors.New("unexpected end of input inside tag")
	case p.depth > 0:
		return fmt.Errorf("unexpected end of input, %d element(s) not closed", p.depth)
	}
	return nil
}

func (p *parser) tagByte(b byte) error {
	switch {
	case p.quote != 0:
		if b == p.quote {
			p.quote = 0
		}
	case b == '"' || b == '\'':
		p.quote = b
	case b == '<':
		return errors.New("found two '<' chars in a row")
	case b == '>':
		p.inTag = false
		return p.handleTag(p.tag.Bytes())
	case p.tag.Len() == 1 && (b == '!' || b == '?'):
		p.inTag = false
		p.inMarkup = markupPI
		if b == '!' {
			p.inMarkup = markupBang
		}
		p.tail = [2]byte{}
		return nil
	}
	return p.tag.Add(b)
}

func (p *parser) markupByte(b byte) {
	if p.inMarkup == markupBang {
		switch b {
		case '-':
			p.inMarkup = markupComment
		case '[':
			p.inMarkup = markupCdata
		default:
			p.inMarkup = markupDeclaration
		}
	}
	switch p.inMarkup {
	case markupComment:
		p.endMarkupAfter(b, '-', '-')
	case markupCdata:
		p.endMarkupAfter(b, ']', ']')
	case markupPI:
		p.endMarkupAfter(b, 0, '?')
	case markupDeclaration:
		switch {
		case p.quote != 0:
			if b == p.quote {
				p.quote = 0
			}
		case b == '"' || b == '\'':
			p.quote = b
		case b == '[':
			p.declDepth++
		case b == ']':
			p.declDepth--
		case b == '>' && p.declDepth <= 0:
			p.inMarkup = markupNone
			p.declDepth = 0
		}
	}
	p.tail = [2]byte{p.tail[1], b}
}

// endMarkupAfter ends the current markup when b is a '>' that follows the
// bytes first and second. A zero first matches any byte.
func (p *parser) endMarkupAfter(b, first, second byte) {
	if b == '>' && p.tail[1] == second && (first == 0 || p.tail[0] == first) {
		p.inMarkup = markupNone
	}
}

// handleTag handles a complete tag, from '<' up to but excluding '>'.
func (p *parser) handleTag(tag []byte) error {
	switch {
	case len(tag) < 2:
		return errors.New("found empty tag <>")
	case tag[1] == '/':
		return p.endTag(tag)
	case tag[len(tag)-1] == '/':
		return p.selfClosingTag(tag)
	default:
		return p.startTag(tag)
	}
}

func (p *parser) startTag(tag []byte) error {
	if err := p.matcher.AddTag(string(tag[1:])); err != nil {
		return err
	}
	p.depth++
	if p.recording || !p.matcher.MatchesPath() {
		return nil
	}
	if !p.innerXml {
		if err := p.recordTag(tag); err != nil {
			return err
		}
	}
	p.match.LineStart = p.line + 1
	p.recording = true
	return nil
}

func (p *parser) endTag(tag []byte) error {
	if p.depth == 0 {
		return errors.New("found end tag before start tag")
	}
	p.depth--
	if p.recording && p.matcher.TagNameMatchesLastMatch() {
		if p.innerXml {
			// The end tag was recorded byte by byte, '<' to '>'.
			p.content.Backup(len(tag) + 1)
		}
		p.recording = false
		p.emit()
	}
	p.matcher.RemoveLast()
	return nil
}

// selfClosingTag handles <tag/>. Inside a match it is already recorded.
func (p *parser) selfClosingTag(tag []byte) error {
	if p.recording {
		return nil
	}
	if err := p.matcher.AddTag(string(tag[1:])); err != nil {
		return err
	}
	defer p.matcher.RemoveLast()
	if !p.matcher.MatchesPath() {
		return nil
	}
	if err := p.recordTag(tag); err != nil {
		return err
	}
	p.match.LineStart = p.line + 1
	p.emit()
	return nil
}

// recordTag adds the tag that starts a match to the content, which only
// records bytes after it.
func (p *parser) recordTag(tag []byte) error {
	if err := p.content.AddArray(tag); err != nil {
		return err
	}
	return p.content.Add('>')
}

// emit hands the recorded match to the emitter and clears it.
func (p *parser) emit() {
	p.match.NodePath = p.matcher.GetCurrentPath()
	p.match.LineEnd = p.line + 1
	p.stopped = p.content.Emit(&p.match)
	p.match.Reset()
	p.content.Reset()
}
