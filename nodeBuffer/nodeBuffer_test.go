package nodeBuffer
import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	emitterOut := make(chan string)
	nb := NewNodeBuffer(1024, emitterOut)
	nb.Add(byte('a'))
	nb.Add(byte('b'))
	nb.Add(byte('c'))
	go emitterEquals(t, emitterOut, "abc")
	nb.Emit()
}

func TestReset(t *testing.T) {
	emitterOut := make(chan string)
	nb := NewNodeBuffer(1024, emitterOut)
	nb.Add(byte('a'))
	nb.Add(byte('b'))
	nb.Add(byte('c'))
	go emitterEquals(t, emitterOut, "abc")
	nb.Emit()
	nb.Reset()
	nb.Add(byte('d'))
	go emitterEquals(t, emitterOut, "d")
	nb.Emit();
}

func TestAddArray(t *testing.T) {
	emitterOut := make(chan string)
	nb := NewNodeBuffer(1024, emitterOut)
	nb.AddArray([]byte{'a', 'b', 'c'})
	go emitterEquals(t, emitterOut, "abc")
	nb.Emit()
}

func emitterEquals(t *testing.T, emitterOut chan string, expected string) {
	assert.Equal(t, <-emitterOut, expected)
}
