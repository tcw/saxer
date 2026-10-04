package tagbuffer

import "errors"

var ErrFull = errors.New("tag is larger than the tag buffer, use --tag-buf to increase buffer")

// TagBuffer holds the bytes of the tag being read, up to a fixed size.
type TagBuffer struct {
	buffer []byte
	pos    int
}

func NewTagBuffer(bufferSize int) TagBuffer {
	return TagBuffer{buffer: make([]byte, bufferSize)}
}

func (tb *TagBuffer) Reset() {
	tb.pos = 0
}

func (tb *TagBuffer) Add(b byte) error {
	if tb.pos >= len(tb.buffer) {
		return ErrFull
	}
	tb.buffer[tb.pos] = b
	tb.pos++
	return nil
}

func (tb *TagBuffer) Len() int {
	return tb.pos
}

func (tb *TagBuffer) Bytes() []byte {
	return tb.buffer[:tb.pos]
}
