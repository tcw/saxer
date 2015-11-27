package saxReader

import (
	"io"
	"github.com/tcw/saxer/histBuffer"
	"github.com/tcw/saxer/contentBuffer"
	"github.com/tcw/saxer/nodePath"
	"bytes"
	"github.com/tcw/saxer/elementBuffer"
	"github.com/tcw/saxer/htmlConverter"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"fmt"
	"github.com/hashicorp/errwrap"
)

type SaxReader struct {
	ElementBufferSize int
	ContentBufferSize int
	ReaderBufferSize  int
	PathDepthSize     int
	EmitterFn         func(string)
	IsInnerXml        bool
	FilterEscapeSigns bool
}

const ONE_KB  int = 1024
const ONE_MB  int = ONE_KB * ONE_KB

func NewSaxReader(emitterFn func(element string), isInnerXml bool, filterEscape bool) SaxReader {
	return SaxReader{ONE_KB * 4, ONE_MB * 4, ONE_KB * 4, 1000, emitterFn, isInnerXml, filterEscape}
}

func NewSaxReaderNoEmitter() SaxReader {
	return SaxReader{ONE_KB * 4, ONE_MB * 4, ONE_KB * 4, 1000, nil, false, false}
}

func (sr *SaxReader) Read(reader io.Reader, query string) error {
	eb := elementBuffer.NewElementBuffer(sr.ElementBufferSize)
	history := histBuffer.NewHistoryBuffer(ONE_KB * 4)
	contentBuffer := contentBuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn)
	nodePath := nodePath.NewNodePath(sr.PathDepthSize, query)
	conv := htmlConverter.NewHtmlConverter()
	buffer := make([]byte, sr.ReaderBufferSize)
	convBuffer := make([]byte, ONE_KB)
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
		hidx := 0
		if sr.FilterEscapeSigns {
			for index := 0; index < n; index++ {
				cn := conv.Translate(convBuffer, buffer[index])
				for ic := 0; ic < cn; ic++ {
					buffer[hidx] = convBuffer[ic]
					hidx ++
				}
			}
		}else {
			hidx = n
		}
		eb.ResetLocalState()
		for index := 0; index < hidx; index++ {
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
				isRecoding, err = ElementType(buffer[eb.LocalStart:eb.LocalEnd], &eb, &contentBuffer, &nodePath, isRecoding, sr.IsInnerXml)
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				eb.ResetLocalState()
			}else if eb.LocalEnd != -1 {
				eb.Add(buffer[:eb.LocalEnd])
				isRecoding, err = ElementType(eb.GetBuffer(), &eb, &contentBuffer, &nodePath, isRecoding, sr.IsInnerXml)
				if err != nil {
					return errwrap.Wrapf(fmt.Sprintf("Error on line %d {{err}}", lineNumber+1), err)
				}
				eb.ResetState()
			}
		}
		if eb.LocalStart == -1 && eb.LocalEnd == -1 && eb.Position > 0 {
			eb.Add(buffer)
		}else if eb.LocalStart != -1 {
			eb.Add(buffer[eb.LocalStart:hidx])
		}
	}
	return nil
}

func ElementType(nodeContent []byte, eb *elementBuffer.ElementBuffer, contentBuffer *contentBuffer.ContentBuffer, nodePath *nodePath.NodePath, isRecoding bool, isInnerXml bool) (bool, error) {
	if nodeContent[1] == byte('/') {
		if eb.StartTags == 0{
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
		eb.StartTags--
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
		eb.StartTags++
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
