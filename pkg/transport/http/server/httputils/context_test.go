package httputils

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaders(t *testing.T) {
	is := assert.New(t)
	ctx := context.Background()
	h := GetHeaders(ctx)
	is.Equal(h, http.Header{})

	expectedHeaders := make(http.Header)
	expectedHeaders.Set("X-Header", "yes")
	ctx = PassHeadersToContext(ctx, expectedHeaders)
	h = GetHeaders(ctx)
	is.Equal(h, expectedHeaders)
}

