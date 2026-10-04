package contentbuffer

import "errors"

var ErrFull = errors.New("content buffer is full, use --cont-buf to increase buffer")

type ContentBuffer struct {
	buf       []byte
	pos       int
	emitterFn func(*EmitterData) bool
}

type EmitterData struct {
	Content   string
	LineStart uint64
	LineEnd   uint64
	NodePath  string
}

func (ed *EmitterData) Reset() {
	ed.Content = ""
	ed.LineStart = 0
	ed.LineEnd = 0
	ed.NodePath = ""
}

func NewContentBuffer(bufferSize int, emitter func(*EmitterData) bool) ContentBuffer {
	return ContentBuffer{buf: make([]byte, bufferSize), emitterFn: emitter}
}

func (cb *ContentBuffer) Reset() {
	cb.pos = 0
}

func (cb *ContentBuffer) Add(b byte) error {
	if cb.pos >= len(cb.buf) {
		return ErrFull
	}
	cb.buf[cb.pos] = b
	cb.pos++
	return nil
}

func (cb *ContentBuffer) AddArray(b []byte) error {
	if cb.pos+len(b) > len(cb.buf) {
		return ErrFull
	}
	copy(cb.buf[cb.pos:], b)
	cb.pos += len(b)
	return nil
}

func (cb *ContentBuffer) Backup(step int) {
	cb.pos -= step
}

func (cb *ContentBuffer) Emit(ed *EmitterData) bool {
	ed.Content = string(cb.buf[:cb.pos])
	return cb.emitterFn(ed)
}
