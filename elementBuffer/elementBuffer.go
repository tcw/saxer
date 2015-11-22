package elementBuffer

type ElementBuffer struct {
	buffer     []byte
	Position   int
	LocalStart int
	LocalEnd   int
}

func NewElementBuffer(bufferSize int) ElementBuffer {
	return ElementBuffer{buffer: make([]byte, bufferSize), Position: 0, LocalStart:-1, LocalEnd:-1}
}

func (eb *ElementBuffer) ResetLocalState() {
	eb.LocalStart = -1
	eb.LocalEnd = -1
}

func (eb *ElementBuffer) ResetState() {
	eb.LocalStart = -1
	eb.LocalEnd = -1
	eb.Position = 0
}

func (eb *ElementBuffer) Add(b []byte) {
	copy(eb.buffer[eb.Position:], b)
	eb.Position = eb.Position + len(b)
}

func (eb *ElementBuffer) GetBuffer() []byte{
	return eb.buffer[:eb.Position]
}
