package fields

import "go.mongodb.org/mongo-driver/bson"

// ID field is just simple variable to predefine "_id" field.
const ID = "_id"

// Date fields of default document
const CreatedAt = "created_at"
const UpdatedAt = "updated_at"
const DeletedAt = "deleted_at"

// Empty is predefined empty map.
var Empty = bson.M{}
