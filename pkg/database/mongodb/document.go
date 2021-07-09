package mongodb

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CollectionGetter interface contain method to return
// document's custom collection.
type CollectionGetter interface {
	// Collection method return collection
	Collection() *Collection
}

// CollectionNameGetter interface contain method to return
// collection name of document.
type CollectionNameGetter interface {
	// CollectionName method return document collection's name.
	CollectionName() string
}

// Document interface is base method that must implement by
// each document, If you're using `DefaultDocument` struct in your document,
// don't need to implement any of those method.
type Document interface {
	// PrepareID convert id value if need, and then
	// return it.(e.g convert string to objectId)
	PrepareID(id interface{}) (interface{}, error)

	GetID() interface{}
	SetID(id interface{})
}

type DefaultDocument struct {
	IDField    `bson:",inline"`
	DateFields `bson:",inline"`
}

func MongoIDFromString(hex string) (id primitive.ObjectID) {
	if hex == "" {
		return primitive.NilObjectID
	}

	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID
	}
	
	return
}

// Creating function call to it's inner fields defined hooks
func (doc *DefaultDocument) Creating() error {
	return doc.DateFields.Creating()
}

// Saving function call to it's inner fields defined hooks
func (doc *DefaultDocument) Saving() error {
	return doc.DateFields.Saving()
}

func (doc *DefaultDocument) Deleting() error {
	return doc.DateFields.Deleting()
}

// IsInitialized checks whether document has id.
func (doc *DefaultDocument) IsInitialized() bool {
	return doc.GetID() != primitive.NilObjectID
}
