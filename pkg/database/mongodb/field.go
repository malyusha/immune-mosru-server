package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDField struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
}

// DateFields struct contain `created_at` and `updated_at`
// fields that autofill on insert/update document.
type DateFields struct {
	CreatedAt primitive.DateTime `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt primitive.DateTime `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt primitive.DateTime `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// PrepareID method prepare id value to using it as id in filtering,...
// e.g convert hex-string id value to bson.ObjectId
func (f *IDField) PrepareID(id interface{}) (interface{}, error) {
	if idStr, ok := id.(string); ok {
		return primitive.ObjectIDFromHex(idStr)
	}

	// Otherwise id must be ObjectId
	return id, nil
}

// String returns HEX representation of ObjectID
func (f *IDField) String() string {
	return f.ID.Hex()
}

// GetID method return document's id
func (f *IDField) GetID() interface{} {
	return f.ID
}

// SetID set id value of document's id field.
func (f *IDField) SetID(id interface{}) {
	f.ID = id.(primitive.ObjectID)
}

// Deleting hook used here to set `deleted_at` field
// value on deleting new document from datebase.
func (f *DateFields) Deleting() error {
	f.DeletedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	return nil
}

// Creating hook used here to set `created_at` field
// value on inserting new document into database.
func (f *DateFields) Creating() error {
	f.CreatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	return nil
}

// Saving hook used here to set `updated_at` field value
// on create/update document.
func (f *DateFields) Saving() error {
	f.UpdatedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	return nil
}
