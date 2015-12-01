package saxReader

import (
	"io"
	"github.com/tcw/saxer/histBuffer"
	"github.com/tcw/saxer/contentBuffer"
	"bytes"
	"github.com/tcw/saxer/elementBuffer"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/tcw/saxer/tagMatcher"
)

type SaxReader struct {
	ElementBufferSize int
	ContentBufferSize int
	ReaderBufferSize  int
	PathDepthSize     int
	EmitterFn         func(string)
	IsInnerXml        bool
}

const ONE_KB  int = 1024
const ONE_MB  int = ONE_KB * ONE_KB

func NewSaxReader(emitterFn func(element string), isInnerXml bool) SaxReader {
	return SaxReader{ONE_KB * 4, ONE_MB * 4, ONE_KB * 4, 1000, emitterFn, isInnerXml}
}

func NewSaxReaderNoEmitter() SaxReader {
	return SaxReader{ONE_KB * 4, ONE_MB * 4, ONE_KB * 4, 1000, nil, false}
}

func (sr *SaxReader) Read(reader io.Reader, query string) error {
	eb := elementBuffer.NewElementBuffer(sr.ElementBufferSize)
	history := histBuffer.NewHistoryBuffer(ONE_KB * 4)
	contentBuffer := contentBuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn)
	tagPath := tagMatcher.NewTagMatcher(sr.PathDepthSize,query)
	buffer := make([]byte, sr.ReaderBufferSize)
	inEscapeMode := false
	isRecoding := false
	var lineNumber uint64 = 0

	for {
		n, err := reader.Read(buffer)
		if n != 0 && err != nil {
			panic("Error while reading xml")
		}
		if n == 0 {
			break
		}
		eb.ResetLocalState()
		for index := 0; index < n; index++ {
			value := buffer[index]
			if isRecoding {
				contentBuffer.Add(value)
			}
			if value == 0x0A {
				lineNumber++
			}
			if inEscapeMode {
				history.Add(value)
				if value == byte('>') {
					if history.HasLast([]byte{'-', '-', '>'}) {
						inEscapeMode = false
						continue
					}
					if history.HasLast([]byte{']', ']', '>'}) {
						inEscapeMode = false
						continue
					}
					if history.HasLast([]byte{'?', '>'}) {
						inEscapeMode = false
						continue
					}
				}
				continue
			}
			if value == byte('<') {
				if eb.LocalStart != -1 || eb.Position > 0 {
					return errors.New(fmt.Sprintf("Validation error found two '<' chars in a row (last on line %d)", lineNumber + 1))
				}
				eb.LocalStart = index
			}
			if value == byte('>') {
				if eb.LocalStart != -1 || eb.Position > 0 {
					eb.LocalEnd = index
				}
			}
			if ((eb.LocalStart != -1 && index != 0 && eb.LocalStart == index - 1) && (value == byte('!') || value == byte('?'))) ||
			(index == 0 && eb.Position == 1 && (value == byte('!') || value == byte('?'))) {
				inEscapeMode = true
				eb.ResetState()
			}else if eb.LocalStart != -1 && eb.LocalEnd != -1 && eb.Position == 0 {
				isRecoding, err = ElementType(buffer[eb.LocalStart:eb.LocalEnd], &eb, &contentBuffer, &tagPath, isRecoding, sr.IsInnerXml)
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				eb.ResetLocalState()
			}else if eb.LocalEnd != -1 {
				eb.Add(buffer[:eb.LocalEnd])
				isRecoding, err = ElementType(eb.GetBuffer(), &eb, &contentBuffer, &tagPath, isRecoding, sr.IsInnerXml)
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				eb.ResetState()
			}
		}
		if eb.LocalStart == -1 && eb.LocalEnd == -1 && eb.Position > 0 {
			eb.Add(buffer)
		}else if eb.LocalStart != -1 {
			eb.Add(buffer[eb.LocalStart:n])
		}
	}
	return nil
}

func ElementType(nodeContent []byte, eb *elementBuffer.ElementBuffer, contentBuffer *contentBuffer.ContentBuffer, matcher *tagMatcher.TagMatcher, isRecoding bool, isInnerXml bool) (bool, error) {
	if nodeContent[1] == byte('/') {
		if eb.StartTags == 0{
			return isRecoding, errors.New("found end tag before start tag")
		}
		if isRecoding {
			if matcher.TagNameMatchesLastMatch() {
				if isInnerXml {
					contentBuffer.Backup(len(nodeContent) + 1)
				}
				contentBuffer.Emit()
				contentBuffer.Reset()
				matcher.RemoveLast()
				return false, nil
			}else {
				matcher.RemoveLast()
				return true, nil
			}
		}
		eb.StartTags--
		matcher.RemoveLast()
		return false, nil
	}else if nodeContent[len(nodeContent) - 1] == byte('/') {
		matcher.AddTag(getNodeName(nodeContent))
		if matcher.MatchesPath() {
			if !isRecoding {
				contentBuffer.AddArray(nodeContent)
				contentBuffer.Add(byte('>'))
				contentBuffer.Emit()
				contentBuffer.Reset()
			}
		}
		matcher.RemoveLast()
		return false, nil
	}else {
		nodename := getNodeName(nodeContent)
		matcher.AddTag(nodename)
		eb.StartTags++
		if !isRecoding {
			if matcher.MatchesPath() {
				if !isInnerXml {
					contentBuffer.AddArray(nodeContent)
					contentBuffer.Add(byte('>'))
				}
				return true, nil
			}else {
				return false, nil
			}
		}else {
			return true, nil
		}
	}
}

func getNodeName(nodeContent []byte) string {
	idx := bytes.IndexByte(nodeContent, byte(' '))
	if idx == -1 {
		return string(nodeContent[1:])
	}else {
		return string(nodeContent[1:idx])
	}
}
