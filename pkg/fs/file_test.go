package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	assert.True(t, FileExists("testdata/exists.dummy"), "Expected FileExist to return true")
	assert.False(t, FileExists("testdata/not-exists.dummy"), "Expected FileExist to return false")
}
