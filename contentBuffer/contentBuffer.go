package contentBuffer

type ContentBuffer struct {
	buf     []byte
	pos     int
	emitter chan string
}

func NewContentBuffer(size int, emitterChannel chan string) ContentBuffer {
	i := make([]byte, size)
	return ContentBuffer{buf: i, pos:0, emitter:emitterChannel }
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
	cb.emitter <- string(cb.buf[:cb.pos])
}


