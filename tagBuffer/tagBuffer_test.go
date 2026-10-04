package tagBuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	tb := NewTagBuffer(8)
	require.NoError(t, tb.Add([]byte("<ab")))
	require.NoError(t, tb.Add([]byte("cd")))
	assert.Equal(t, []byte("<abcd"), tb.GetBuffer())
}

func TestAddFillsBufferExactly(t *testing.T) {
	tb := NewTagBuffer(4)
	require.NoError(t, tb.Add([]byte("<abc")))
	assert.Equal(t, []byte("<abc"), tb.GetBuffer())
}

func TestAddOverflowReturnsError(t *testing.T) {
	tb := NewTagBuffer(4)
	require.NoError(t, tb.Add([]byte("<ab")))
	assert.Error(t, tb.Add([]byte("cd")))
	assert.Equal(t, []byte("<ab"), tb.GetBuffer())
}

func TestResetState(t *testing.T) {
	tb := NewTagBuffer(8)
	tb.LocalStart, tb.LocalEnd = 1, 2
	require.NoError(t, tb.Add([]byte("<ab")))

	tb.ResetLocalState()
	assert.Equal(t, -1, tb.LocalStart)
	assert.Equal(t, -1, tb.LocalEnd)
	assert.Equal(t, []byte("<ab"), tb.GetBuffer())

	tb.ResetState()
	assert.Empty(t, tb.GetBuffer())
}
