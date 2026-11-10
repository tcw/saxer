package xmlgen

import (
	"bytes"
	"encoding/xml"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteIsWellFormedAndSized(t *testing.T) {
	var b bytes.Buffer
	require.NoError(t, Write(&b, Options{Size: 200_000, Seed: 1}))
	assert.GreaterOrEqual(t, b.Len(), 200_000)
	assert.Less(t, b.Len(), 210_000)

	d := xml.NewDecoder(bytes.NewReader(b.Bytes()))
	d.Strict = true
	d.Entity = map[string]string{"note": "generated > data"}
	for {
		_, err := d.Token()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}
	for _, want := range []string{"<!DOCTYPE", "<!--", "<![CDATA[", "<?inspection", "<option level=\"39\">", "<info>"} {
		assert.True(t, bytes.Contains(b.Bytes(), []byte(want)), want)
	}
}

func TestWriteIsDeterministic(t *testing.T) {
	var a, b bytes.Buffer
	require.NoError(t, Write(&a, Options{Size: 50_000, Seed: 7}))
	require.NoError(t, Write(&b, Options{Size: 50_000, Seed: 7}))
	assert.Equal(t, a.Bytes(), b.Bytes())
}
