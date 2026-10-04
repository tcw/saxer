package tagbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addAll(tb *TagBuffer, s string) error {
	for i := 0; i < len(s); i++ {
		if err := tb.Add(s[i]); err != nil {
			return err
		}
	}
	return nil
}

func TestAdd(t *testing.T) {
	tb := NewTagBuffer(8)
	require.NoError(t, addAll(&tb, "<abcd"))
	assert.Equal(t, []byte("<abcd"), tb.Bytes())
	assert.Equal(t, 5, tb.Len())
}

func TestAddFillsBufferExactly(t *testing.T) {
	tb := NewTagBuffer(4)
	require.NoError(t, addAll(&tb, "<abc"))
	assert.Equal(t, []byte("<abc"), tb.Bytes())
}

func TestAddOverflowReturnsError(t *testing.T) {
	tb := NewTagBuffer(4)
	require.NoError(t, addAll(&tb, "<abc"))
	assert.ErrorIs(t, tb.Add('d'), ErrFull)
	assert.Equal(t, []byte("<abc"), tb.Bytes())
}

func TestReset(t *testing.T) {
	tb := NewTagBuffer(8)
	require.NoError(t, addAll(&tb, "<ab"))
	tb.Reset()
	assert.Empty(t, tb.Bytes())
	assert.Equal(t, 0, tb.Len())
}
