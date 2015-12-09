package htmlConverter
import (
	"math"
)

type HtmlConverter struct {
	state      [6][6]rune
	stateNum   int64
	stateLevel int64
	buffer     []rune
	bufferPos  int
}

func NewHtmlConverter() HtmlConverter {
	var st [6][6]rune
	st[0][0] = rune('&')
	st[1][0] = rune('q')
	st[1][1] = rune('a')
	st[1][2] = rune('l')
	st[1][3] = rune('g')
	st[2][0] = rune('u')
	st[2][1] = rune('m')
	st[2][2] = rune('t')
	st[2][3] = rune('p')
	st[3][0] = rune('o')
	st[3][1] = rune('p')
	st[3][2] = rune(';')
	st[4][0] = rune('t')
	st[4][1] = rune(';')
	st[4][2] = rune('s')
	st[5][0] = rune(';')

	return HtmlConverter{state:st, stateNum:0,buffer:make([]rune,100), bufferPos:0}
}

func (hc *HtmlConverter)Translate(dest []rune, r rune) int {
	found := false
	for i := 0; i < len(hc.state[hc.stateLevel]); i++ {
		if hc.state[hc.stateLevel][i] == r {
			hc.stateNum = hc.stateNum + ((hc.stateLevel * 10) + int64(i)) * int64(math.Pow10(int(10 - (hc.stateLevel * 2))))
			hc.buffer[hc.bufferPos] = r
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
		dest[pos] = r
		hc.clear()
		return pos+1
	}
	if hc.stateNum == 1020304050 {
		copy(dest, []rune{'"'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1121314100 {
		copy(dest, []rune{'&'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1222320000 {
		copy(dest, []rune{'<'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1322320000 {
		copy(dest, []rune{'>'})
		hc.clear()
		return 1
	}else if hc.stateNum == 1123304250 {
		copy(dest, []rune{'\''})
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


