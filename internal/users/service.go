package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

const (
	defaultInviteCodeLen = 4
	// defaultInvitesPerUser is the default number of invites, that are available per user.
	defaultInvitesPerUser      = 1
	defaultInvitesPerAdmin     = 10
	defaultGenerationsPerAdmin = 10000
)

type SingleUserFilter struct {
	TelegramID int
}

type CreateUserData struct {
	Username  string
	FirstName string
	LastName  string
}

type Service interface {
	// FindOrCreateUser finds user by given filter or creates new one, using given data merged with filter data.
	FindOrCreateUser(ctx context.Context, filter SingleUserFilter, data CreateUserData) (*internal.User, error)

	// ActivateInvite activates invite on user with given ID.
	ActivateInvite(ctx context.Context, userID string, inviteCode string) (*internal.User, error)

	// GetUserInvites returns active user invites.
	GetUserInvites(ctx context.Context, userID string) ([]internal.Invite, error)

	// AddInvitesToUser adds more invites for given user.
	AddInvitesToUser(ctx context.Context, user *internal.User) (*internal.User, error)
}

type service struct {
	users            internal.UsersStorage
	uniqueInviteCode string
	inviteCodeLen    int
	invitesPerUser   int
	telegramAdminId  *int
	invitesForAdmin  int
}

func NewService(users internal.UsersStorage, opts ...Option) (*service, error) {
	s := &service{
		users:           users,
		inviteCodeLen:   defaultInviteCodeLen,
		invitesPerUser:  defaultInvitesPerUser,
		invitesForAdmin: defaultInvitesPerAdmin,
	}

	for _, o := range opts {
		o(s)
	}

	return s, nil
}

func (s *service) FindOrCreateUser(ctx context.Context, filter SingleUserFilter, data CreateUserData) (*internal.User, error) {
	user, err := s.users.FindUser(ctx, internal.FindUserFilter{TelegramID: &filter.TelegramID})
	if err != nil && err != internal.ErrNoUser {
		return nil, errdefs.Unknown(err)
	}

	if err == internal.ErrNoUser {
		logger.WithContext(ctx).Infof("creating new user for tg id %d", filter.TelegramID)
		return s.createUser(ctx, filter.TelegramID, data)
	}

	return user, nil
}

func (s *service) createUser(ctx context.Context, tgId int, data CreateUserData) (*internal.User, error) {
	fullName := strings.Trim(fmt.Sprintf("%s %s", data.LastName, data.FirstName), " ")
	u := internal.NewUser(data.Username, fullName, &internal.TelegramData{Username: data.Username, ID: tgId})
	if s.telegramAdminId != nil && tgId == *s.telegramAdminId {
		u.QRGenerationsLeft = defaultGenerationsPerAdmin
	}

	user, err := s.users.CreateUser(ctx, u)

	if err != nil {
		return nil, errdefs.Unknown(fmt.Errorf("failed to create new user: %w", err))
	}

	return user, nil
}

var (
	ErrNoInvite = errors.New("no such invite")
)
