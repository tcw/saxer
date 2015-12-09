package saxReader

import (
	"io"
	"github.com/tcw/saxer/histBuffer"
	"github.com/tcw/saxer/contentBuffer"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/tcw/saxer/tagMatcher"
	"github.com/tcw/saxer/tagBuffer"
)

type SaxReader struct {
	ElementBufferSize int
	ContentBufferSize int
	ReaderBufferSize  int
	PathDepthSize     int
	EmitterFn         func(*contentBuffer.EmitterData) bool
	IsInnerXml        bool
}

const ONE_KB  int = 1024
const ONE_MB  int = ONE_KB * ONE_KB

func NewSaxReaderNoEmitter() SaxReader {
	return SaxReader{ONE_KB * 4, ONE_MB * 4, ONE_KB * 4, 1000, nil, false}
}

func (sr *SaxReader) Read(reader io.Reader, query string) error {
	tb := tagBuffer.NewTagBuffer(sr.ElementBufferSize)
	history := histBuffer.NewHistoryBuffer(ONE_KB * 4)
	contentBuf := contentBuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn)
	tagPath := tagMatcher.NewTagMatcher(sr.PathDepthSize,query)
	buffer := make([]byte, sr.ReaderBufferSize)
	emitterData := &contentBuffer.EmitterData{}
	inEscapeMode := false
	isRecoding := false
	stop := false
	var lineNumber uint64 = 0

	for {
		n, err := reader.Read(buffer)
		if n != 0 && err != nil {
			panic("Error while reading xml")
		}
		if n == 0 {
			break
		}
		tb.ResetLocalState()
		for index := 0; index < n; index++ {
			value := buffer[index]
			if isRecoding {
				contentBuf.Add(value)
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
				if tb.LocalStart != -1 || tb.Position > 0 {
					return errors.New(fmt.Sprintf("Validation error found two '<' chars in a row (last on line %d)", lineNumber + 1))
				}
				tb.LocalStart = index
			}
			if value == byte('>') {
				if tb.LocalStart != -1 || tb.Position > 0 {
					tb.LocalEnd = index
				}
			}
			if ((tb.LocalStart != -1 && index != 0 && tb.LocalStart == index - 1) && (value == byte('!') || value == byte('?'))) ||
			(index == 0 && tb.Position == 1 && (value == byte('!') || value == byte('?'))) {
				inEscapeMode = true
				tb.ResetState()
			}else if tb.LocalStart != -1 && tb.LocalEnd != -1 && tb.Position == 0 {
				stop,isRecoding, err = TagHandler(buffer[tb.LocalStart:tb.LocalEnd], &tb, &contentBuf, &tagPath,emitterData, isRecoding, sr.IsInnerXml)
				if stop{
					return nil
				}
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				tb.ResetLocalState()
			}else if tb.LocalEnd != -1 {
				tb.Add(buffer[:tb.LocalEnd])
				stop,isRecoding, err = TagHandler(tb.GetBuffer(), &tb, &contentBuf, &tagPath, emitterData, isRecoding, sr.IsInnerXml)
				if stop {
					return nil
				}
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				tb.ResetState()
			}
		}
		if tb.LocalStart == -1 && tb.LocalEnd == -1 && tb.Position > 0 {
			tb.Add(buffer)
		}else if tb.LocalStart != -1 {
			tb.Add(buffer[tb.LocalStart:n])
		}
	}
	return nil
}

//todo: clean up!
func TagHandler(nodeContent []byte, tb *tagBuffer.TagBuffer, contentBuffer *contentBuffer.ContentBuffer, matcher *tagMatcher.TagMatcher,emitterData *contentBuffer.EmitterData, isRecoding bool, isInnerXml bool) (bool, bool, error) {
	if nodeContent[1] == byte('/') {
		if tb.StartTags == 0{
			return false, isRecoding, errors.New("found end tag before start tag")
		}
		if isRecoding {
			if matcher.TagNameMatchesLastMatch() {
				if isInnerXml {
					contentBuffer.Backup(len(nodeContent) + 1)
				}
				stop := contentBuffer.Emit(emitterData)
				emitterData.Reset()
				if stop{
					return true ,false, nil
				}
				contentBuffer.Reset()
				matcher.RemoveLast()
				return false,false, nil
			}else {
				matcher.RemoveLast()
				return false,true, nil
			}
		}
		tb.StartTags--
		matcher.RemoveLast()
		return false, false, nil
	}else if nodeContent[len(nodeContent) - 1] == byte('/') {
		matcher.AddTag(string(nodeContent[1:]))
		if matcher.MatchesPath() {
			if !isRecoding {
				contentBuffer.AddArray(nodeContent)
				contentBuffer.Add(byte('>'))
				stop := contentBuffer.Emit(emitterData)
				emitterData.Reset()
				if stop{
					return true,false,nil
				}
				contentBuffer.Reset()
			}
		}
		matcher.RemoveLast()
		return false,false, nil
	}else {
		matcher.AddTag(string(nodeContent[1:]))
		tb.StartTags++
		if !isRecoding {
			if matcher.MatchesPath() {
				if !isInnerXml {
					contentBuffer.AddArray(nodeContent)
					contentBuffer.Add(byte('>'))
				}
				return false,true, nil
			}else {
				return false,false, nil
			}
		}else {
			return false,true, nil
		}
	}
}
