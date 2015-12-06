package saxReader

import (
	"io"
	"github.com/tcw/saxer/histBuffer"
	"github.com/tcw/saxer/contentBuffer"
	"github.com/tcw/saxer/nodePath"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/tcw/saxer/tagBuffer"
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
	tb := tagBuffer.NewTagBuffer(sr.ElementBufferSize)
	history := histBuffer.NewHistoryBuffer(ONE_KB * 4)
	contentBuffer := contentBuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn)
	nodePath := nodePath.NewNodePath(sr.PathDepthSize, query)
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
		tb.ResetLocalState()
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
				isRecoding, err = ElementType(buffer[tb.LocalStart:tb.LocalEnd], &tb, &contentBuffer, &nodePath, isRecoding, sr.IsInnerXml)
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				tb.ResetLocalState()
			}else if tb.LocalEnd != -1 {
				tb.Add(buffer[:tb.LocalEnd])
				isRecoding, err = ElementType(tb.GetBuffer(), &tb, &contentBuffer, &nodePath, isRecoding, sr.IsInnerXml)
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

func ElementType(nodeContent []byte, tb *tagBuffer.TagBuffer, contentBuffer *contentBuffer.ContentBuffer, nodePath *nodePath.NodePath, isRecoding bool, isInnerXml bool) (bool, error) {
	if nodeContent[1] == byte('/') {
		if tb.StartTags == 0{
			return isRecoding, errors.New("found end tag before start tag")
		}
		if isRecoding {
			if nodePath.MatchesLastMatch() {
				if isInnerXml {
					contentBuffer.Backup(len(nodeContent) + 1)
				}
				contentBuffer.Emit()
				contentBuffer.Reset()
				nodePath.RemoveLast()
				return false, nil
			}else {
				nodePath.RemoveLast()
				return true, nil
			}
		}
		tb.StartTags--
		nodePath.RemoveLast()
		return false, nil
	}else if nodeContent[len(nodeContent) - 1] == byte('/') {
		nodePath.Add(getNodeName(nodeContent))
		if nodePath.MatchesPath() {
			if !isRecoding {
				contentBuffer.AddArray(nodeContent)
				contentBuffer.Add(byte('>'))
				contentBuffer.Emit()
				contentBuffer.Reset()
			}
		}
		nodePath.RemoveLast()
		return false, nil
	}else {
		nodename := getNodeName(nodeContent)
		nodePath.Add(nodename)
		tb.StartTags++
		if !isRecoding {
			if nodePath.MatchesPath() {
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
