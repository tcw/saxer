package htmlConverter
import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNoMatch(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]rune, 10)

	res := hc.Translate(arr, rune('q'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], rune('q'))
}

func TestQuot(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]rune, 10)

	res := hc.Translate(arr, rune('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('q'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('u'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('o'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], rune('"'))
}

func TestAmp(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]rune, 10)

	res := hc.Translate(arr, rune('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('a'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('m'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('p'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], rune('&'))
}

func TestLt(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]rune, 10)

	res := hc.Translate(arr, rune('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('l'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], rune('<'))
}

func TestGt(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]rune, 10)

	res := hc.Translate(arr, rune('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('g'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune(';'))
	assert.Equal(t, res, 1)
	assert.Equal(t, arr[0], rune('>'))
}

func TestPartMatch(t *testing.T) {
	hc := NewHtmlConverter()
	arr := make([]rune, 10)

	res := hc.Translate(arr, rune('&'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('g'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('t'))
	assert.Equal(t, res, 0)
	res = hc.Translate(arr, rune('f'))
	assert.Equal(t, res, 4)
	assert.Equal(t, arr[0], rune('&'))
	assert.Equal(t, arr[1], rune('g'))
	assert.Equal(t, arr[2], rune('t'))
	assert.Equal(t, arr[3], rune('f'))
}

func BenchmarkConverterWorstCase(b *testing.B) {
	dest := make([]rune, 100)
	hc := NewHtmlConverter()
	for i := 0; i < b.N; i++ {
		hc.Translate(dest, rune('&'))
		hc.Translate(dest, rune('l'))
		hc.Translate(dest, rune('t'))
		hc.Translate(dest, rune(';'))
	}
}

func BenchmarkConverterBestCase(b *testing.B) {
	dest := make([]rune, 100)
	hc := NewHtmlConverter()
	for i := 0; i < b.N; i++ {
		hc.Translate(dest, rune('a'))
	}
}