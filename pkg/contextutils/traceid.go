package contextutils

import (
	"context"

	"github.com/rs/xid"
)

type traceIDKey struct{}

// NewTraceID returns new ID, generated using xid library.
func NewTraceID() string {
	return xid.New().String()
}

// EnsureTraceID checks whether given context contains trace-id. If doesn't it writes new trace-id
// into context.
func EnsureTraceID(ctx context.Context) context.Context {
	if GetTraceID(ctx) == "" {
		return SetTraceID(ctx, NewTraceID())
	}

	return ctx
}

// SetTraceID provides given id to context and returns new context with value.
func SetTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, id)
}

// GetTraceID returns trace ID from context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey{}).(string); ok {
		return id
	}
	return ""
}
