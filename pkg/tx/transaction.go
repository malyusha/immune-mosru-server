package tx

// Package tx provides interface with Transactable interface to support transactionable operations
// on service logic.

import (
	"context"
)

type Transactable interface {
	Transaction(ctx context.Context, fn Transaction) error
}

type Context interface {
	context.Context
	Abort() error
}

type Transaction func(ctx Context) error
