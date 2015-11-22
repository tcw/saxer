package contentBuffer

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var actual string = ""


func emitterTestFn(element string) {
	actual = element
}

func TestAdd(t *testing.T) {
	cb := NewContentBuffer(1024, emitterTestFn)
	cb.Add(byte('a'))
	cb.Add(byte('b'))
	cb.Add(byte('c'))
	cb.Emit()
	assert.Equal(t, actual, "abc")
}

func TestReset(t *testing.T) {
	cb := NewContentBuffer(1024, emitterTestFn)
	cb.Add(byte('a'))
	cb.Add(byte('b'))
	cb.Add(byte('c'))
	cb.Emit()
	assert.Equal(t, actual, "abc")
	cb.Reset()
	cb.Add(byte('d'))
	cb.Emit();
	assert.Equal(t, actual, "d")
}

func TestAddArray(t *testing.T) {
	cb := NewContentBuffer(1024, emitterTestFn)
	cb.AddArray([]byte{'a', 'b', 'c'})
	cb.Emit()
	assert.Equal(t, actual, "abc")
}

