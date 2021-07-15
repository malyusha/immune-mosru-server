package internal

import (
	"context"
	"errors"
	"time"
)

const (
	// DefaultQRGenerationsLimit is the limit of possible generations of QR codes per user.
	DefaultQRGenerationsLimit = 1
)

type UsersStorage interface {
	FindByInvite(ctx context.Context, code string) (*User, error)
	FindUser(ctx context.Context, filter FindUserFilter) (*User, error)

	CreateUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, id string, user User) error
}

type User struct {
	ID                string
	Login             string
	Name              string
	Invites           []Invite
	InvitedBy         string // ID of user, that invited current user into system
	QRGenerationsLeft int
	TelegramData      *TelegramData // Data, provided by telegram chat
	IsActive          bool          // IsActive means user has entered invite code
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type FindUserFilter struct {
	TelegramID   *int
	InviteActive bool
	InviteCode   *string
	ID           *string
	IsActive     *bool
}

type Invite struct {
	Code   string     // unique code of invite
	UsedBy *string    // UserID that have used this invite. If filled, then invite is not active anymore.
	UsedAt *time.Time // time when invite has been used.
}

// IsUsed returns flag whether invite has been used by somebody.
func (i *Invite) IsUsed() bool {
	return i.UsedAt != nil || i.UsedBy != nil
}

type TelegramData struct {
	ID       int
	Username string
}

func NewUser(login, name string, tgData *TelegramData) User {
	return User{
		Login:             login,
		Name:              name,
		QRGenerationsLeft: DefaultQRGenerationsLimit,
		TelegramData:      tgData,
		IsActive:          false, // inactive by default
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func NewInvite(code string) Invite {
	return Invite{
		Code: code,
	}
}

var (
	ErrNoInvite = errors.New("invite not found")
	ErrNoUser   = errors.New("user not found")
)
