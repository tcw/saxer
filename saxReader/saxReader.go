package saxReader

import (
	"io"
	"github.com/tcw/saxer/histBuffer"
	"github.com/tcw/saxer/contentBuffer"
	"github.com/tcw/saxer/nodePath"
	"bytes"
	"github.com/tcw/saxer/elementBuffer"
)

type SaxReader struct {
	ElementBufferSize int
	ContentBufferSize int
	ReaderBufferSize  int
	PathDepthSize     int
	Reader            io.Reader
	EmitterFn         func(string)
	PathQuery         string
}

const FOUR_KB  int = 1024 * 4

func NewSaxReader(reader io.Reader, emitterFn func(element string), pathQuery string) SaxReader {
	return SaxReader{FOUR_KB, 1024 * 1024 * 4, FOUR_KB, 1000, reader, emitterFn, pathQuery}
}

func (sr *SaxReader) Read() {
	eb := elementBuffer.NewElementBuffer(sr.ElementBufferSize)
	history := histBuffer.NewHistoryBuffer(FOUR_KB)
	contentBuffer := contentBuffer.NewContentBuffer(sr.ContentBufferSize, sr.EmitterFn)
	nodePath := nodePath.NewNodePath(sr.PathDepthSize, sr.PathQuery)
	buffer := make([]byte, sr.ReaderBufferSize)

	inEscapeMode := false
	isRecoding := false
	var lineNumber uint64 = 0

	for {
		n, err := sr.Reader.Read(buffer)
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
				}
				continue
			}
			if value == byte('<') {
				eb.LocalStart = index
			}
			if value == byte('>') {
				eb.LocalEnd = index
			}
			if ((eb.LocalStart != -1 && index != 0 && eb.LocalStart == index - 1) && value == byte('!')) ||
			(index == 0 && eb.Position == 1 && value == byte('!')) {
				inEscapeMode = true
				eb.ResetState()
			}else if eb.LocalStart != -1 && eb.LocalEnd != -1 && eb.Position == 0 {
				isRecoding = ElementType(buffer[eb.LocalStart:eb.LocalEnd], &contentBuffer, &nodePath, isRecoding)
				eb.ResetLocalState()
			}else if eb.LocalEnd != -1 {
				eb.Add(buffer[:eb.LocalEnd])
				isRecoding = ElementType(eb.GetBuffer(), &contentBuffer, &nodePath, isRecoding)
				eb.ResetState()
			}
		}
		if eb.LocalStart == -1 && eb.LocalEnd == -1 && eb.Position > 0 {
			eb.Add(buffer)
		}else if eb.LocalStart != -1 {
			eb.Add(buffer[eb.LocalStart:n])
		}
	}
}

func ElementType(nodeContent []byte, nodeBuffer *contentBuffer.ContentBuffer, nodePath *nodePath.NodePath, isRecoding bool) bool {
	if nodeContent[1] == byte('/') {
		if isRecoding {
			if nodePath.MatchesLastMatch() {
				nodeBuffer.Emit()
				nodeBuffer.Reset()
				nodePath.RemoveLast()
				return false
			}else {
				nodePath.RemoveLast()
				return true
			}
		}
		nodePath.RemoveLast()
		return false
	}else if nodeContent[len(nodeContent) - 1] == byte('/') {
		nodePath.Add(getNodeName(nodeContent))
		if nodePath.MatchesPath() {
			if !isRecoding {
				nodeBuffer.AddArray(nodeContent)
				nodeBuffer.Add(byte('>'))
				nodeBuffer.Emit()
				nodeBuffer.Reset()
			}
		}
		nodePath.RemoveLast()
		return false
	}else {
		nodename := getNodeName(nodeContent)
		nodePath.Add(nodename)
		if !isRecoding {
			if nodePath.MatchesPath() {
				nodeBuffer.AddArray(nodeContent)
				nodeBuffer.Add(byte('>'))
				return true
			}else {
				return false
			}
		}else {
			return true
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