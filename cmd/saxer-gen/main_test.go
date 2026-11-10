package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSize(t *testing.T) {
	for in, want := range map[string]int64{"1024": 1024, "64K": 64 << 10, "500m": 500 << 20, " 2G ": 2 << 30} {
		got, err := parseSize(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"", "G", "-1M", "1T", "0"} {
		_, err := parseSize(in)
		assert.Error(t, err, in)
	}
}
