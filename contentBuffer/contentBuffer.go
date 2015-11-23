package contentBuffer

type ContentBuffer struct {
	buf     []byte
	pos     int
	emitterFn func(string)
}

func NewContentBuffer(size int, emitter func(string)) ContentBuffer {
	i := make([]byte, size)
	return ContentBuffer{buf: i, pos:0, emitterFn:emitter }
}

func (cb *ContentBuffer)Reset() {
	cb.pos = 0
}

func (cb *ContentBuffer)Add(b byte) {
	cb.buf[cb.pos] = b
	cb.pos++
}

func (cb *ContentBuffer)AddArray(b []byte) {
	copy(cb.buf[cb.pos:], b)
	cb.pos = cb.pos + len(b)
}

func (cb *ContentBuffer) Emit() {
	cb.emitterFn(string(cb.buf[:cb.pos]))
}

