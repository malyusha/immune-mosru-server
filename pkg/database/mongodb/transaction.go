package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/malyusha/immune-mosru-server/pkg/tx"
)

type mongoTx struct {
	client *mongo.Client
}

func NewTransactable(client *mongo.Client) *mongoTx {
	return &mongoTx{client: client}
}

func (c *mongoTx) Transaction(ctx context.Context, fn tx.Transaction) error {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Majority()
	txOpts := options.Transaction().
		SetWriteConcern(wc).
		SetReadConcern(rc).
		// Set read preference to primary for multi-document transactions
		// see https://docs.mongodb.com/manual/core/transactions/#read-concern-write-concern-read-preference
		SetReadPreference(readpref.Primary())

	sess, err := c.client.StartSession()
	if err != nil {
		return err
	}

	// end session at the end of transaction
	defer sess.EndSession(ctx)

	callback := func(sessionCtx mongo.SessionContext) (interface{}, error) {
		ctx := &mongoCtx{
			SessionContext: sessionCtx,
		}

		if err := fn(ctx); err != nil {
			return nil, err
		}

		return nil, nil
	}

	if _, err := sess.WithTransaction(ctx, callback, txOpts); err != nil {
		return err
	}

	return nil
}

type mongoCtx struct {
	mongo.SessionContext
}

func (c *mongoCtx) Abort() error {
	return c.AbortTransaction(c.SessionContext)
}
