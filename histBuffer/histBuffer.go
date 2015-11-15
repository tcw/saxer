package histBuffer

import "math"

type HistoryBuffer struct {
	buf []byte
	pos int
}

func NewHistoryBuffer(size int) HistoryBuffer {
	i := make([]byte, size)
	return HistoryBuffer{buf: i, pos:0}
}

func (hb *HistoryBuffer)Add(b byte) {
	hb.buf[hb.pos] = b
	if hb.pos == len(hb.buf) - 1 {
		hb.pos = 0
	}else {
		hb.pos++
	}
}

func (hb *HistoryBuffer)GetLast(n int) []byte {
	if hb.pos - n < 0 {
		if hb.pos == 0 {
			return hb.buf[len(hb.buf) - n:len(hb.buf)]
		}else {
			fromEnd := int(math.Abs(float64(hb.pos - n)))
			return append(hb.buf[len(hb.buf) - fromEnd:], hb.buf[:hb.pos]...)
		}
	}else {
		return hb.buf[hb.pos - n:hb.pos]
	}
}

func (hb *HistoryBuffer)HasLast(bytes []byte) bool {
	byteFromBuffer := hb.GetLast(len(bytes))
	for i := 0; i < len(bytes); i++ {
		if bytes[i] != byteFromBuffer[i] {
			return false
		}
	}
	return true
}

