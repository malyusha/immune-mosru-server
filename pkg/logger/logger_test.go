package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFields_Merge(t *testing.T) {
	fields := Fields{"one": 1, "two": 2}
	expect := Fields{"one": 1, "two": 5, "three": 3}

	merged := fields.Merge(Fields{"two": 5, "three": 3})

	assert.Equal(t, expect, merged)
}
