package histBuffer
import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetLast2(t *testing.T) {
	hb := NewHistoryBuffer(5)
	hb.Add(byte('a'))
	hb.Add(byte('b'))
	hb.Add(byte('c'))
	hb.Add(byte('d'))
	last := hb.GetLast(2)
	assert.Equal(t,last,[]byte{'c','d'})
}


func TestGetLastSwich(t *testing.T) {
	hb := NewHistoryBuffer(3)
	hb.Add(byte('a'))
	hb.Add(byte('b'))
	hb.Add(byte('c'))
	hb.Add(byte('d'))
	last := hb.GetLast(2)
	assert.Equal(t,last,[]byte{'c','d'})
}

func TestGetLastOnZeroPos(t *testing.T) {
	hb := NewHistoryBuffer(3)
	hb.Add(byte('a'))
	hb.Add(byte('b'))
	hb.Add(byte('c'))
	last := hb.GetLast(2)
	assert.Equal(t,last,[]byte{'b','c'})
}

func TestHasLast(t *testing.T) {
	hb := NewHistoryBuffer(5)
	hb.Add(byte('a'))
	hb.Add(byte('b'))
	last := hb.HasLast([]byte{'b'})
	assert.True(t,last)
}

func TestHasLast2(t *testing.T) {
	hb := NewHistoryBuffer(5)
	hb.Add(byte('a'))
	hb.Add(byte('b'))
	hb.Add(byte('c'))
	last := hb.HasLast([]byte{'b','c'})
	assert.True(t,last)
}

func BenchmarkAdd(b *testing.B) {
	hb := NewHistoryBuffer(4096)
	for i := 0; i < b.N; i++ {
		hb.Add(byte('a'))
	}
}

func BenchmarkGetLast(b *testing.B) {
	hb := NewHistoryBuffer(4096)
	for i := 0; i < 4096; i++  {
		hb.Add(byte('a'))
	}
	for i := 0; i < b.N; i++ {
		hb.GetLast(5)
	}
}