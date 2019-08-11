package tagPath

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPathLength0(t *testing.T) {
	tp := NewTagPath()
	currentPath := tp.GetCurrentPath()
	assert.Equal(t, "", currentPath)
}

func TestGetPathLength1(t *testing.T) {
	tp := NewTagPath()
	tp.Path[0].Name = "node1"
	tp.PathPos++
	currentPath := tp.GetCurrentPath()
	assert.Equal(t, "node1", currentPath)
}

func TestGetPathLength2(t *testing.T) {
	tp := NewTagPath()
	tp.Path[0].Name = "node1"
	tp.Path[1].Name = "node2"
	tp.PathPos = 2
	currentPath := tp.GetCurrentPath()
	assert.Equal(t, "node1/node2", currentPath)
}
