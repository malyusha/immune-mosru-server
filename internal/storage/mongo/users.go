package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/pkg/database/mongodb"
)

type usersStorage struct {
	*mongodb.Collection
}

func (s *usersStorage) FindByInvite(ctx context.Context, code string) (*internal.User, error) {
	filter := bson.M{
		"invites": bson.M{
			"$elemMatch": bson.M{
				"usedby": bson.M{"$eq": nil},
				"code":   bson.M{"$eq": code},
			},
		},
	}
	// select only invites

	var dst User
	res := s.FindOne(ctx, filter)
	err := res.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, internal.ErrNoInvite
		}

		return nil, err
	}

	if err := res.Decode(&dst); err != nil {
		return nil, fmt.Errorf("failed to decode invite document: %w", err)
	}

	return mapUser(dst), nil
}

func (s *usersStorage) FindUser(ctx context.Context, f internal.FindUserFilter) (*internal.User, error) {
	filter := parseFindUserFilter(f)
	var mongoUser User

	if err := s.First(ctx, filter, &mongoUser); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, internal.ErrNoUser
		}

		return nil, fmt.Errorf("failed to execute First on users: %w", err)
	}

	return mapUser(mongoUser), nil
}

func (s *usersStorage) CreateUser(ctx context.Context, user internal.User) (*internal.User, error) {
	mongoUser := mapUserToMongo(&user)
	if err := s.Create(ctx, mongoUser); err != nil {
		return nil, fmt.Errorf("failed to execute Create on users: %w", err)
	}

	return mapUser(*mongoUser), nil
}

func (s *usersStorage) UpdateUser(ctx context.Context, id string, user internal.User) error {
	mongoUser := mapUserToMongo(&user)
	if err := s.Update(ctx, mongoUser); err != nil {
		if err == mongo.ErrNoDocuments {
			return internal.ErrNoUser
		}

		return err
	}

	return nil
}

type User struct {
	mongodb.DefaultDocument `bson:",inline"`
	Login                   string                 `bson:"login"`
	Name                    string                 `bson:"name"`
	Invites                 []internal.Invite      `bson:"invites"`
	InvitedBy               string                 `bson:"invited_by"` // ID of user, that invited current user into system
	QRGenerationsLeft       int                    `bson:"qr_generations_left"`
	TelegramData            *internal.TelegramData `bson:"telegram"`  // Data, provided by telegram chat
	IsActive                bool                   `bson:"is_active"` // IsActive means user has entered invite code
}

func (s *usersStorage) ListIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.M{"telegram.id": 1},
			Options: options.Index().SetName("telegram_id_1"),
		},
		{
			Keys:    bson.M{"invites.code": 1},
			Options: options.Index().SetName("invites_code_1"),
		},
		{
			Keys:    bson.D{{"invites.code", 1}, {"is_active", 1}},
			Options: options.Index().SetName("invites_code_1_is_active_1"),
		},
	}
}

func mapUser(u User) *internal.User {
	return &internal.User{
		ID:                u.ID.Hex(),
		Login:             u.Login,
		Name:              u.Name,
		Invites:           u.Invites,
		InvitedBy:         u.InvitedBy,
		QRGenerationsLeft: u.QRGenerationsLeft,
		TelegramData:      u.TelegramData,
		IsActive:          u.IsActive,
		CreatedAt:         u.CreatedAt.Time(),
		UpdatedAt:         u.UpdatedAt.Time(),
	}
}

func mapUserToMongo(u *internal.User) *User {
	user := &User{
		Login:             u.Login,
		Name:              u.Name,
		InvitedBy:         u.InvitedBy,
		Invites:           u.Invites,
		QRGenerationsLeft: u.QRGenerationsLeft,
		TelegramData:      u.TelegramData,
		IsActive:          u.IsActive,
	}

	if u.ID != "" {
		id, _ := user.PrepareID(u.ID)
		user.SetID(id)
	}

	return user
}

func parseFindUserFilter(filter internal.FindUserFilter) bson.M {
	f := bson.M{}
	if filter.ID != nil {
		id, _ := primitive.ObjectIDFromHex(*filter.ID)
		f["_id"] = id
	}

	if filter.TelegramID != nil {
		f["telegram.id"] = *filter.TelegramID
	}

	if filter.IsActive != nil {
		f["is_active"] = *filter.IsActive
	}

	if filter.InviteCode != nil {
		inviteFilter := bson.M{
			"code": bson.M{"$eq": *filter.InviteCode},
		}
		if filter.InviteActive {
			inviteFilter["usedby"] = bson.M{"$eq": nil}
		}

		f["invites"] = bson.M{
			"$elemMatch": inviteFilter,
		}
	}

	return f
}

func NewUsersStorage(ctx context.Context, cfg mongodb.CollectionConfig) (*usersStorage, error) {
	storage := &usersStorage{}

	cfg.CollectionName = "users"
	coll, err := mongodb.NewCollection(cfg, storage)
	if err != nil {
		return nil, err
	}

	storage.Collection = coll

	return storage, nil
}
