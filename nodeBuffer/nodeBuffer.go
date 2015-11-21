package nodeBuffer

type NodeBuffer struct {
	buf     []byte
	pos     int
	emitter chan string
}

func NewNodeBuffer(size int, emitterChannel chan string) NodeBuffer {
	i := make([]byte, size)
	return NodeBuffer{buf: i, pos:0, emitter:emitterChannel }
}

func (nb *NodeBuffer)Reset() {
	nb.pos = 0
}

func (nb *NodeBuffer)Add(b byte) {
	nb.buf[nb.pos] = b
	nb.pos++
}

func (nb *NodeBuffer)AddArray(b []byte) {
	copy(nb.buf[nb.pos:], b)
	nb.pos = nb.pos + len(b)
}

func (nb *NodeBuffer) Emit() {
	nb.emitter <- string(nb.buf[:nb.pos])
}


