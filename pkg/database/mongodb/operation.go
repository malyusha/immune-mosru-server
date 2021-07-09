package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb/fields"
)

func create(ctx context.Context, c *Collection, document Document, opts ...*options.InsertOneOptions) error {
	// Call to saving hook
	if err := callToBeforeCreateHooks(document); err != nil {
		return err
	}

	res, err := c.InsertOne(ctx, document, opts...)

	if err != nil {
		return err
	}

	// Set new id
	document.SetID(res.InsertedID)

	return callToAfterCreateHooks(document)
}

func first(ctx context.Context, c *Collection, filter interface{}, document Document, opts ...*options.FindOneOptions) error {
	return c.FindOne(ctx, filter, opts...).Decode(document)
}

func update(ctx context.Context, c *Collection, document Document, opts ...*options.UpdateOptions) error {
	// Call to saving hook
	if err := callToBeforeUpdateHooks(document); err != nil {
		return err
	}

	res, err := c.UpdateOne(ctx, bson.M{fields.ID: document.GetID()}, bson.M{"$set": document}, opts...)

	if err != nil {
		return err
	}

	return callToAfterUpdateHooks(res, document)
}

func del(ctx context.Context, c *Collection, document Document) error {
	if err := callToBeforeDeleteHooks(document); err != nil {
		return err
	}

	if c.softDeletes {
		update := bson.M{fields.DeletedAt: primitive.NewDateTimeFromTime(time.Now().UTC())}
		res, err := c.UpdateOne(ctx, bson.M{fields.ID: document.GetID()}, bson.M{"$set": update})
		if err != nil {
			return err
		}

		return callToAfterUpdateHooks(res, document)
	}

	res, err := c.DeleteOne(ctx, bson.M{fields.ID: document.GetID()})
	if err != nil {
		return err
	}

	return callToAfterDeleteHooks(res, document)
}
