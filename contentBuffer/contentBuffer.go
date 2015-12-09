package contentBuffer

type ContentBuffer struct {
	buf       []byte
	pos       int
	emitterFn func(string,uint64,string) bool
}

func NewContentBuffer(size int, emitter func(string,uint64,string) bool) ContentBuffer {
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

func (cb *ContentBuffer)Backup(step int) {
	cb.pos = cb.pos - step
}

func (cb *ContentBuffer) Emit(lineNumber uint64, path string) bool {
	return cb.emitterFn(string(cb.buf[:cb.pos]), lineNumber, path)
}


