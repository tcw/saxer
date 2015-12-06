package tagBuffer

type TagBuffer struct {
	buffer     []byte
	Position   int
	LocalStart int
	LocalEnd   int
	StartTags  int
}

func NewTagBuffer(bufferSize int) TagBuffer {
	return TagBuffer{buffer: make([]byte, bufferSize), Position: 0, LocalStart:-1, LocalEnd:-1, StartTags:0}
}

func (eb *TagBuffer) ResetLocalState() {
	eb.LocalStart = -1
	eb.LocalEnd = -1
}

func (eb *TagBuffer) ResetState() {
	eb.LocalStart = -1
	eb.LocalEnd = -1
	eb.Position = 0
}

func (eb *TagBuffer) Add(b []byte) {
	copy(eb.buffer[eb.Position:], b)
	eb.Position = eb.Position + len(b)
}

func (eb *TagBuffer) GetBuffer() []byte {
	return eb.buffer[:eb.Position]
}
