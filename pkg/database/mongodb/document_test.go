package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Doc struct {
	DefaultDocument `bson:",inline"`

	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

func TestPrepareInvalidId(t *testing.T) {
	d := &Doc{}

	_, err := d.PrepareID("test")
	assert.Error(t, err, "Expected get error on invalid id value")
}

func TestPrepareId(t *testing.T) {
	d := &Doc{}

	hexId := "5df7fb2b1fff9ee374b6bd2a"
	val, err := d.PrepareID(hexId)
	id, _ := primitive.ObjectIDFromHex(hexId)
	assert.Equal(t, val.(primitive.ObjectID), id)
	assert.NoError(t, err)
}
