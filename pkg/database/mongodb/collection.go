package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb/fields"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

var collections = make(map[string]bool)

type IndexedCollection interface {
	ListIndexes() []mongo.IndexModel
}

type CollectionConfig struct {
	DB             *mongo.Database
	CollectionName string
}

func NewCollectionConfig(db *mongo.Database) CollectionConfig {
	return CollectionConfig{
		DB: db,
	}
}

// Collection performs operations on documents and given Mongodb collection
type Collection struct {
	softDeletes bool
	*mongo.Collection
	logger logger.Logger
}

// FindByID method find a doc and decode it to document, otherwise return error.
// id field can be any value that if passed to `PrepareID` method, it return
// valid id(e.g string,bson.ObjectId).
func (coll *Collection) FindByID(ctx context.Context, id interface{}, document Document) error {
	id, err := document.PrepareID(id)

	if err != nil {
		return err
	}

	return first(ctx, coll, bson.M{fields.ID: id}, document)
}

// First method search and return first document of search result.
func (coll *Collection) First(ctx context.Context, filter interface{}, document Document, opts ...*options.FindOneOptions) error {
	return first(ctx, coll, filter, document, opts...)
}

// CreateVariant method insert new document into database.
func (coll *Collection) Create(ctx context.Context, document Document, opts ...*options.InsertOneOptions) error {
	return create(ctx, coll, document, opts...)
}

// UpdateVariant function update save changed document into database.
// On call to this method also mgm call to document's updating,updated,
// saving,saved hooks.
func (coll *Collection) Update(ctx context.Context, document Document, opts ...*options.UpdateOptions) error {
	return update(ctx, coll, document, opts...)
}

// Delete method delete document (doc) from collection.
// If you want to doing something on deleting some document
// use hooks, don't need to override this method.
func (coll *Collection) Delete(ctx context.Context, document Document) error {
	return del(ctx, coll, document)
}

// SimpleFind find and decode result to results.
func (coll *Collection) SimpleFind(ctx context.Context, results interface{}, filter interface{}, opts ...*options.FindOptions) error {
	cur, err := coll.Find(ctx, filter, opts...)

	if err != nil {
		return err
	}

	return cur.All(ctx, results)
}

func (coll *Collection) ensureIndexes(ctx context.Context, indexes ...mongo.IndexModel) error {
	idxs := coll.Indexes()
	log := coll.logger.With(logger.Fields{"collection": coll.Name()})
	log.Infof("Ensuring indexes")
	indexesPositions := make(map[string]int, len(indexes))

	// fill existing indexes map with keys
	for pos, indexToCreate := range indexes {
		if indexToCreate.Options == nil {
			return errors.New("index options is nil. provide Name for index creation")
		}

		name := indexToCreate.Options.Name
		if name == nil || *name == "" {
			return fmt.Errorf("failed to create index %coll without specifying name explicitly", indexToCreate.Keys)
		}

		indexesPositions[*name] = pos
	}

	cur, err := idxs.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list indexes: %w", err)
	}

	existingIndexes := make(map[string]bool)
	for cur.Next(ctx) {
		d := bsoncore.Document{}
		if err := cur.Decode(&d); err != nil {
			return fmt.Errorf("failed to decode bson document containing index")
		}
		name := d.Lookup("name")
		existingIndexes[name.StringValue()] = true
	}

	createList := make([]mongo.IndexModel, 0)
	for name, pos := range indexesPositions {
		if !existingIndexes[name] {
			createList = append(createList, indexes[pos])
		}
	}

	if len(createList) > 0 {
		log.Infof("Indexes are filtered. Creating %d indexes", len(createList))
		list, err := idxs.CreateMany(ctx, createList)
		if err != nil {
			return fmt.Errorf("failed to create indexes for %q: %w", coll.Name(), err)
		}

		log.With(logger.Fields{"indexes": list}).Info("successfully created indexes")
	} else {
		log.Debug("all indexes are already created")
	}

	return nil
}

// ensureCollection ensures that collection exist, it creates new collection if it
// doesn't exist.
func (coll *Collection) ensureCollection(ctx context.Context) error {
	log := coll.logger.WithContext(ctx)
	log.Debug("ensuring collection exist")

	var err error
	var list []string
	if len(collections) == 0 {
		list, err = coll.Database().ListCollectionNames(ctx, bson.M{})

		for _, name := range list {
			collections[name] = true
		}
	}

	if err != nil {
		return fmt.Errorf("failed to list collection names: %w", err)
	}

	if collections[coll.Name()] {
		log.Debug("collection already created. skipping.")
		return nil
	}

	if err := coll.Database().CreateCollection(ctx, coll.Name()); err != nil {
		return err
	}

	log.Info("collection successfully created")
	return nil
}

// CreateIndex returns new mongo index with name and fields.
func CreateIndex(name string, keys bson.M) mongo.IndexModel {
	return mongo.IndexModel{
		Keys:    keys,
		Options: &options.IndexOptions{Name: &name},
	}
}

func NewCollection(cfg CollectionConfig, concrete interface{}) (*Collection, error) {
	coll := &Collection{
		softDeletes: false,
		Collection:  cfg.DB.Collection(cfg.CollectionName),
		logger: logger.With(logger.Fields{
			"component": "collection",
			"name":      cfg.CollectionName,
		}),
	}

	if err := coll.ensureCollection(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure collection %q: %w", coll.Name(), err)
	}

	if iColl, ok := concrete.(IndexedCollection); ok {
		collectionIndexes := iColl.ListIndexes()
		indexes := append(getCollectionIndices(), collectionIndexes...)

		// ensure that concrete collection storage indexes are created.
		if err := coll.ensureIndexes(context.Background(), indexes...); err != nil {
			return nil, fmt.Errorf("failed to ensure defaultIndexes on concrete storage: %w", err)
		}
	}

	return coll, nil
}

func getCollectionIndices() []mongo.IndexModel {
	deletedAt := CreateIndex("deleted_at_-1", bson.M{"deleted_at": -1})
	createdAt := CreateIndex("created_at_-1", bson.M{"created_at": -1})
	updatedAt := CreateIndex("updated_at_-1", bson.M{"updated_at": -1})

	return []mongo.IndexModel{createdAt, updatedAt, deletedAt}
}
