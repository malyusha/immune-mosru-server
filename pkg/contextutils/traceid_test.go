package contextutils

import (
	"context"
	"testing"

	assertion "github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	assert := assertion.New(t)

	ctx := context.Background()

	rid := GetTraceID(ctx)
	assert.Equal(rid, "")

	requestID := "trace-id"
	ctx = SetTraceID(ctx, requestID)

	rid = GetTraceID(ctx)
	assert.Equal(rid, requestID)
}
