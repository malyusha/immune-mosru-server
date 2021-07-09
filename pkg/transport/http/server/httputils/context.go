package httputils

import (
	"context"
	"net/http"

	"github.com/malyusha/immune-mosru-server/pkg/contextutils"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

type userIdCtxKeyType string
type requestHeadersCtxKeyType string

const (
	headersCtxKey requestHeadersCtxKeyType = "headers"
	userIdCtxKey  userIdCtxKeyType         = "userID"
)

// GetHeaders returns headers from context.
func GetHeaders(ctx context.Context) http.Header {
	if headers, ok := ctx.Value(headersCtxKey).(http.Header); ok {
		return headers
	}
	return http.Header{}
}

// PassHeadersToContext providers given headers to context and returns new parent context.
func PassHeadersToContext(ctx context.Context, h http.Header) context.Context {
	return context.WithValue(ctx, headersCtxKey, h)
}

// GetUserID returns previously set user id from given context.
// To set user id to ctx use SetUserID func below.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIdCtxKey).(string); ok {
		return id
	}

	return ""
}

// SetUserID passes given ID to ctx and returns parent with value.
func SetUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIdCtxKey, id)
}

// ExtractLogFields returns logger fields from given context for further usage inside logger.
func ExtractLogFields(ctx context.Context) logger.Fields {
	fields := logger.Fields{}
	// trace ID from context
	reqID := contextutils.GetTraceID(ctx)
	if reqID != "" {
		fields["request_id"] = reqID
	}

	return fields
}
