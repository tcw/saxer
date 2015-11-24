package htmlConverter
import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNoMatch(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]byte, 10)

	res := hc.Translate(arr, byte('q'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], byte('q'))
}

func TestQuot(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]byte, 10)

	res := hc.Translate(arr, byte('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('q'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('u'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('o'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], byte('"'))
}

func TestAmp(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]byte, 10)

	res := hc.Translate(arr, byte('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('a'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('m'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('p'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], byte('&'))
}

func TestLt(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]byte, 10)

	res := hc.Translate(arr, byte('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('l'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], byte('<'))
}

func TestGt(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]byte, 10)

	res := hc.Translate(arr, byte('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('g'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], byte('>'))
}

func TestPartMatch(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]byte, 10)

	res := hc.Translate(arr, byte('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('g'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, byte('f'))
	assert.Equal(t, res, 4)
	assert.Equal(t, arr[0], byte('&'))
	assert.Equal(t, arr[1], byte('g'))
	assert.Equal(t, arr[2], byte('t'))
	assert.Equal(t, arr[3], byte('f'))
}