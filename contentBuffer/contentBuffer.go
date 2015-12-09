package contentBuffer

type ContentBuffer struct {
	buf           []byte
	pos           int
	emitterFn     func(*EmitterData) bool
}

type EmitterData struct {
	Content   string
	LineStart uint64
	LineEnd   uint64
	NodePath  string
}

func (ed *EmitterData)Reset() {
	ed.Content = ""
	ed.LineStart = 0
	ed.LineEnd = 0
	ed.NodePath = ""
}

func NewContentBuffer(size int, emitter func(*EmitterData) bool) ContentBuffer {
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

func (cb *ContentBuffer) Emit(ed *EmitterData) bool {
	ed.Content = string(cb.buf[:cb.pos])
	return cb.emitterFn(ed)
}


