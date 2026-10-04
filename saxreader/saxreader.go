package saxreader

import (
	"fmt"
	"github.com/tcw/saxer/contentbuffer"
	"github.com/tcw/saxer/histbuffer"
	"io"

	"errors"
	"github.com/tcw/saxer/tagbuffer"
	"github.com/tcw/saxer/tagmatcher"
)

type SaxReader struct {
	ElementBufferSize int
	ContentBufferSize int
	ReaderBufferSize  int
	PathDepthSize     int
	EmitterFn         func(*contentbuffer.EmitterData) bool
	IsInnerXml        bool
}

const ONE_KB int = 1024
const ONE_MB int = ONE_KB * ONE_KB

func NewSaxReaderNoEmitter() SaxReader {

	return SaxReader{ElementBufferSize: ONE_KB * 4,
		ContentBufferSize: ONE_MB * 4,
		ReaderBufferSize:  ONE_KB * 4,
		PathDepthSize:     1000,
		EmitterFn:         nil,
		IsInnerXml:        false}
}

// Kinds of markup that is skipped until its terminator: <!-- -->, <![CDATA[ ]]>,
// <? ?> and declarations such as <!DOCTYPE ...>.
const (
	escapeNone        = 0
	escapeBang        = 1 // seen "<!", kind not known yet
	escapeComment     = 2
	escapeCdata       = 3
	escapePI          = 4
	escapeDeclaration = 5
)

var escapeNames = map[int]string{
	escapeBang:        "declaration",
	escapeComment:     "comment",
	escapeCdata:       "CDATA section",
	escapePI:          "processing instruction",
	escapeDeclaration: "declaration",
}

func (sr *SaxReader) Read(reader io.Reader, tm *tagmatcher.TagMatcher) error {
	tb := tagbuffer.NewTagBuffer(sr.ElementBufferSize)
	history := histbuffer.NewHistoryBuffer(ONE_KB * 4)
	contentBuf := contentbuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn)
	buffer := make([]byte, sr.ReaderBufferSize)
	emitterData := &contentbuffer.EmitterData{}
	escape := escapeNone
	var quote byte = 0    // quote char of the attribute value being read, inside tags and declarations
	declarationDepth := 0 // nesting of [ ] in a declaration (DOCTYPE internal subset)
	isRecoding := false
	stop := false
	var err error
	var lineNumber uint64 = 0

	for {
		n, readErr := reader.Read(buffer)
		tb.ResetLocalState()
		for index := 0; index < n; index++ {
			value := buffer[index]
			if isRecoding {
				bufferFullErr := contentBuf.Add(value)
				if bufferFullErr != nil {
					return bufferFullErr
				}
			}
			if value == 0x0A {
				lineNumber++
			}
			if escape != escapeNone {
				history.Add(value)
				if escape == escapeBang {
					switch value {
					case '-':
						escape = escapeComment
					case '[':
						escape = escapeCdata
					default:
						escape = escapeDeclaration
					}
				}
				switch escape {
				case escapeComment:
					if value == '>' && history.HasLast([]byte("-->")) {
						escape = escapeNone
					}
				case escapeCdata:
					if value == '>' && history.HasLast([]byte("]]>")) {
						escape = escapeNone
					}
				case escapePI:
					if value == '>' && history.HasLast([]byte("?>")) {
						escape = escapeNone
					}
				case escapeDeclaration:
					switch {
					case quote != 0:
						if value == quote {
							quote = 0
						}
					case value == '"' || value == '\'':
						quote = value
					case value == '[':
						declarationDepth++
					case value == ']':
						declarationDepth--
					case value == '>' && declarationDepth <= 0:
						escape = escapeNone
						declarationDepth = 0
					}
				}
				continue
			}
			inTag := tb.LocalStart != -1 || tb.Position > 0
			if inTag && quote != 0 {
				if value == quote {
					quote = 0
				}
				continue
			}
			if inTag && (value == '"' || value == '\'') {
				quote = value
				continue
			}
			if value == byte('<') {
				if inTag {
					return errors.New(fmt.Sprintf("Validation error found two '<' chars in a row (last on line %d)", lineNumber+1))
				}
				tb.LocalStart = index
			}
			if value == byte('>') {
				if inTag {
					tb.LocalEnd = index
				}
			}
			if ((tb.LocalStart != -1 && index != 0 && tb.LocalStart == index-1) && (value == byte('!') || value == byte('?'))) ||
				(index == 0 && tb.Position == 1 && (value == byte('!') || value == byte('?'))) {
				if value == '!' {
					escape = escapeBang
				} else {
					escape = escapePI
				}
				tb.ResetState()
			} else if tb.LocalStart != -1 && tb.LocalEnd != -1 && tb.Position == 0 {
				stop, isRecoding, err = TagHandler(buffer[tb.LocalStart:tb.LocalEnd], &tb, &contentBuf, tm, emitterData, isRecoding, sr.IsInnerXml, lineNumber)
				if stop {
					return nil
				}
				if err != nil {
					return fmt.Errorf("Error on line %d %w", lineNumber+1, err)
				}
				tb.ResetLocalState()
			} else if tb.LocalEnd != -1 {
				if err = tb.Add(buffer[:tb.LocalEnd]); err != nil {
					return fmt.Errorf("Error on line %d %w", lineNumber+1, err)
				}
				stop, isRecoding, err = TagHandler(tb.GetBuffer(), &tb, &contentBuf, tm, emitterData, isRecoding, sr.IsInnerXml, lineNumber)
				if stop {
					return nil
				}
				if err != nil {
					return fmt.Errorf("Error on line %d %w", lineNumber+1, err)
				}
				tb.ResetState()
			}
		}
		var addErr error
		if tb.LocalStart == -1 && tb.LocalEnd == -1 && tb.Position > 0 {
			addErr = tb.Add(buffer[:n])
		} else if tb.LocalStart != -1 {
			addErr = tb.Add(buffer[tb.LocalStart:n])
		}
		if addErr != nil {
			return fmt.Errorf("Error on line %d %w", lineNumber+1, addErr)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("error reading xml: %w", readErr)
		}
	}
	switch {
	case escape != escapeNone:
		return fmt.Errorf("unexpected end of input inside %s", escapeNames[escape])
	case tb.LocalStart != -1 || tb.Position > 0:
		return errors.New("unexpected end of input inside tag")
	case tb.StartTags > 0:
		return fmt.Errorf("unexpected end of input, %d element(s) not closed", tb.StartTags)
	}
	return nil
}

// todo: clean up!
func TagHandler(nodeContent []byte, tb *tagbuffer.TagBuffer, cb *contentbuffer.ContentBuffer, matcher *tagmatcher.TagMatcher, emitterData *contentbuffer.EmitterData, isRecoding bool, isInnerXml bool, lineNumber uint64) (bool, bool, error) {
	if len(nodeContent) < 2 {
		return false, isRecoding, errors.New("found empty tag <>")
	}
	if nodeContent[1] == byte('/') {
		if tb.StartTags == 0 {
			return false, isRecoding, errors.New("found end tag before start tag")
		}
		tb.StartTags--
		if isRecoding {
			if matcher.TagNameMatchesLastMatch() {
				if isInnerXml {
					cb.Backup(len(nodeContent) + 1)
				}
				emitterData.NodePath = matcher.GetCurrentPath()
				emitterData.LineEnd = lineNumber + 1
				stop := cb.Emit(emitterData)
				emitterData.Reset()
				if stop {
					return true, false, nil
				}
				cb.Reset()
				matcher.RemoveLast()
				return false, false, nil
			} else {
				matcher.RemoveLast()
				return false, true, nil
			}
		}
		matcher.RemoveLast()
		return false, false, nil
	} else if nodeContent[len(nodeContent)-1] == byte('/') {
		if !isRecoding {
			if err := matcher.AddTag(string(nodeContent[1:])); err != nil {
				return false, false, err
			}
			if matcher.MatchesPath() {
				bufferFullErr := cb.AddArray(nodeContent)
				if bufferFullErr != nil {
					return false, false, bufferFullErr
				}
				bufferFullErr = cb.Add(byte('>'))
				if bufferFullErr != nil {
					return false, false, bufferFullErr
				}
				emitterData.NodePath = matcher.GetCurrentPath()
				emitterData.LineStart = lineNumber + 1
				emitterData.LineEnd = lineNumber + 1
				stop := cb.Emit(emitterData)
				emitterData.Reset()
				if stop {
					return true, false, nil
				}
				cb.Reset()
			}
			matcher.RemoveLast()
		}
		return false, isRecoding, nil
	} else {
		if err := matcher.AddTag(string(nodeContent[1:])); err != nil {
			return false, isRecoding, err
		}
		tb.StartTags++
		if !isRecoding {
			if matcher.MatchesPath() {
				if !isInnerXml {
					bufferFullErr := cb.AddArray(nodeContent)
					if bufferFullErr != nil {
						return false, false, bufferFullErr
					}
					bufferFullErr = cb.Add(byte('>'))
					if bufferFullErr != nil {
						return false, false, bufferFullErr
					}
				}
				emitterData.LineStart = lineNumber + 1
				return false, true, nil
			} else {
				return false, false, nil
			}
		} else {
			return false, true, nil
		}
	}
}
