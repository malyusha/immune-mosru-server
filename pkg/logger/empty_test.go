package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyLogger(t *testing.T) {
	var logger Logger
	assert.Nil(t, logger)
	logger = Empty{}
	assert.NotNil(t, logger)
}
