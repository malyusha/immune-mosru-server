package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb"
	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
)

type certsStorage struct {
	*mongodb.Collection
}

func (c *certsStorage) ListIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{"code", 1}, {"created_at", -1}},
			Options: options.Index().SetName("code_1_created_at_-1"),
		},
	}
}

type Cert struct {
	mongodb.DefaultDocument `bson:",inline"`

	OwnerID     string `json:"owner_id" bson:"owner_id"`
	Code        string `json:"code" bson:"code"`
	FirstName   string `json:"first_name" bson:"first_name"`
	LastName    string `json:"last_name" bson:"last_name"`
	SecondName  string `json:"second_name" bson:"second_name"`
	DateOfBirth string `json:"date_birth" bson:"date_birth"`

	ExpiringAt primitive.DateTime `json:"expiring_at" bson:"expiring_at"`
}

func (c *certsStorage) ListCertificates(ctx context.Context) ([]*internal.VaxCert, error) {
	dst := make([]Cert, 0)
	err := c.SimpleFind(ctx, &dst, bson.M{})
	if err != nil {
		return nil, errdefs.Unknown(err)
	}

	res := make([]*internal.VaxCert, len(dst))
	for i, c := range dst {
		res[i] = mapCert(c)
	}

	return res, nil
}

func (c *certsStorage) DeleteCert(ctx context.Context, code string) error {
	return errdefs.NotImplemented(errors.New("not implemented"))
}

func (c *certsStorage) UpdateCert(ctx context.Context, code string, cert internal.UpdateCertData) (*internal.VaxCert, error) {
	return nil, errdefs.NotImplemented(errors.New("not implemented"))
}

func (c *certsStorage) GetByCode(ctx context.Context, code string) (*internal.VaxCert, error) {
	dst := Cert{}
	if err := c.First(ctx, bson.M{"code": code}, &dst); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, internal.ErrCertNotFound
		}

		return nil, fmt.Errorf("failed to execut First on certificates storage")
	}

	return mapCert(dst), nil
}

func (c *certsStorage) Exists(ctx context.Context, code string) (bool, error) {
	dst := new(Cert)
	if err := c.First(ctx, bson.M{"code": code}, dst); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *certsStorage) CreateCert(ctx context.Context, cert internal.VaxCert) (*internal.VaxCert, error) {
	if cert.Code == "" {
		return nil, errdefs.InvalidParameter(errors.New("certificate missing code"))
	}

	doc := mapCertToMongo(&cert)
	err := c.Create(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create new certificate: %w", err)
	}

	return mapCert(*doc), nil
}

func mapCert(c Cert) *internal.VaxCert {
	return &internal.VaxCert{
		ID:          c.ID.Hex(),
		OwnerID:     c.OwnerID,
		Code:        c.Code,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		SecondName:  c.SecondName,
		DateOfBirth: c.DateOfBirth,
		CreatedAt:   c.CreatedAt.Time(),
		UpdatedAt:   c.UpdatedAt.Time(),
		ExpiringAt:  c.ExpiringAt.Time(),
	}
}

func mapCertToMongo(c *internal.VaxCert) *Cert {
	cert := &Cert{
		Code:        c.Code,
		OwnerID:     c.OwnerID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		SecondName:  c.SecondName,
		DateOfBirth: c.DateOfBirth,
		ExpiringAt:  primitive.NewDateTimeFromTime(c.ExpiringAt),
	}

	if c.ID != "" {
		id, _ := cert.PrepareID(c.ID)
		cert.SetID(id)
	}

	return cert
}

func NewCertsStorage(ctx context.Context, cfg mongodb.CollectionConfig) (*certsStorage, error) {
	storage := &certsStorage{}

	cfg.CollectionName = "certs"
	coll, err := mongodb.NewCollection(cfg, storage)
	if err != nil {
		return nil, err
	}

	storage.Collection = coll

	return storage, nil
}
