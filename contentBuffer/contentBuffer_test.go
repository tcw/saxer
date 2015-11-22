package contentBuffer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	emitterOut := make(chan string)
	cb := NewContentBuffer(1024, emitterOut)
	cb.Add(byte('a'))
	cb.Add(byte('b'))
	cb.Add(byte('c'))
	go emitterEquals(t, emitterOut, "abc")
	cb.Emit()
}

func TestReset(t *testing.T) {
	emitterOut := make(chan string)
	cb := NewContentBuffer(1024, emitterOut)
	cb.Add(byte('a'))
	cb.Add(byte('b'))
	cb.Add(byte('c'))
	go emitterEquals(t, emitterOut, "abc")
	cb.Emit()
	cb.Reset()
	cb.Add(byte('d'))
	go emitterEquals(t, emitterOut, "d")
	cb.Emit();
}

func TestAddArray(t *testing.T) {
	emitterOut := make(chan string)
	cb := NewContentBuffer(1024, emitterOut)
	cb.AddArray([]byte{'a', 'b', 'c'})
	go emitterEquals(t, emitterOut, "abc")
	cb.Emit()
}

func emitterEquals(t *testing.T, emitterOut chan string, expected string) {
	assert.Equal(t, <-emitterOut, expected)
}
