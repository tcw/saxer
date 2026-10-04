package contentBuffer

import "errors"

type ContentBuffer struct {
	size      int
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
	i := make([]byte, bufferSize)
	return ContentBuffer{size: bufferSize, buf: i, pos: 0, emitterFn: emitter}
}

func (cb *ContentBuffer) Reset() {
	cb.pos = 0
}

func (cb *ContentBuffer) Add(b byte) error {
	if cb.size-1 <= cb.pos {
		return errors.New("ContentBuffer is full, use --cont-buf to increase buffer!")
	}
	cb.buf[cb.pos] = b
	cb.pos++
	return nil
}

func (cb *ContentBuffer) AddArray(b []byte) error {
	if cb.size-1 <= cb.pos+len(b) {
		return errors.New("ContentBuffer is full, use --cont-buf to increase buffer!")
	}
	copy(cb.buf[cb.pos:], b)
	cb.pos = cb.pos + len(b)
	return nil
}

func (cb *ContentBuffer) Backup(step int) {
	cb.pos = cb.pos - step
}

func (cb *ContentBuffer) Emit(ed *EmitterData) bool {
	ed.Content = string(cb.buf[:cb.pos])
	return cb.emitterFn(ed)
}
