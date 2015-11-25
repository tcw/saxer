package htmlConverter
import (
	"math"
)


type HtmlConverter struct {
	state      [6][6]byte
	stateNum   int64
	stateLevel int64
	buffer     []byte
	bufferPos  int
}

func NewHtmlConverter() HtmlConverter {
	var st [6][6]byte
	st[0][0] = '&'
	st[1][0] = 'q'
	st[1][1] = 'a'
	st[1][2] = 'l'
	st[1][3] = 'g'
	st[2][0] = 'u'
	st[2][1] = 'm'
	st[2][2] = 't'
	st[2][3] = 'p'
	st[3][0] = 'o'
	st[3][1] = 'p'
	st[3][2] = ';'
	st[4][0] = 't'
	st[4][1] = ';'
	st[4][2] = 's'
	st[5][0] = ';'

	return HtmlConverter{state:st, stateNum:0,buffer:make([]byte,100), bufferPos:0}
}

func (hc *HtmlConverter)Translate(dest []byte, b byte) int {
	found := false
	for i := 0; i < len(hc.state[hc.stateLevel]); i++ {
		if hc.state[hc.stateLevel][i] == b {
			hc.stateNum = hc.stateNum + ((hc.stateLevel * 10) + int64(i)) * int64(math.Pow10(int(10 - (hc.stateLevel * 2))))
			hc.buffer[hc.bufferPos] = b
			hc.bufferPos++
			hc.stateLevel++
			found = true
			break
		}
	}
	if found == false {
		pos := hc.bufferPos
		if pos != 0{
			copy(dest, hc.buffer[:pos])
		}
		dest[pos] = b
		hc.clear()
		return pos+1
	}
	if hc.stateNum == 1020304050 {
		copy(dest, []byte{'"'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1121314100 {
		copy(dest, []byte{'&'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1222320000 {
		copy(dest, []byte{'<'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1322320000 {
		copy(dest, []byte{'>'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1123304250 {
		copy(dest, []byte{'\''})
		hc.clear()
		return 1
	}
	return 0
}

func (hc *HtmlConverter) clear()  {
	hc.stateNum = 0
	hc.stateLevel = 0
	hc.bufferPos = 0
}


