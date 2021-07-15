package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/util"
)

func (s *service) AddInvitesToUser(ctx context.Context, user *internal.User) (*internal.User, error) {
	logger.Infof("adding more invites to user %s", user.ID)

	if err := s.generateInvitesForUser(ctx, user); err != nil {
		return nil, err
	}

	if err := s.users.UpdateUser(ctx, user.ID, *user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) generateInvitesForUser(ctx context.Context, user *internal.User) error {
	// generate invite for activated user
	num := s.invitesPerUser
	if s.telegramAdminId != nil && user.TelegramData.ID == *s.telegramAdminId {
		num = s.invitesForAdmin
	}

	for num != 0 {
		invite, err := s.createInvite(ctx)
		if err != nil {
			return errdefs.Unknown(fmt.Errorf("failed to generate invite: %w", err))
		}

		// add invite to user
		user.Invites = append(user.Invites, invite)
		num--
	}

	return nil
}

// createInvite returns new Invite instance.
func (s *service) createInvite(ctx context.Context) (internal.Invite, error) {
	invite := internal.Invite{Code: util.GenerateCode(s.inviteCodeLen)}
	errRetries := 3
	for errRetries != 0 {
		_, err := s.users.FindByInvite(ctx, invite.Code)
		if err != nil {
			if err == internal.ErrNoInvite {
				return invite, nil
			}

			return invite, err
		}

		errRetries--
	}

	return invite, errors.New("unable to generate invite within 3 tries")
}

func (s *service) GetUserInvites(ctx context.Context, userID string) ([]internal.Invite, error) {
	log := logger.WithContext(ctx)
	log.Debugf("loading invites of user %s", userID)

	user, err := s.users.FindUser(ctx, internal.FindUserFilter{ID: &userID})
	if err != nil {
		if err == internal.ErrNoUser {
			return nil, errdefs.NotFound(err)
		}

		return nil, errdefs.Unknown(err)
	}

	activeInvites := make([]internal.Invite, 0)
	for _, inv := range user.Invites {
		if !inv.IsUsed() {
			activeInvites = append(activeInvites, inv)
		}
	}

	return activeInvites, nil
}

func (s *service) ActivateInvite(ctx context.Context, userID string, inviteCode string) (*internal.User, error) {
	log := logger.WithContext(ctx)
	log.Debugf("activating invite for user %s", userID)
	currentUser, err := s.users.FindUser(ctx, internal.FindUserFilter{ID: &userID})
	if err != nil {
		return nil, errdefs.NotFound(err)
	}

	inviteOwner, err := s.getOwnerFromCode(ctx, inviteCode)
	if err != nil {
		return nil, fmt.Errorf("failed to receive invite-code owner: %w", err)
	}

	// ok, invite is active, now let's activate user and mark invite as used
	currentUser.IsActive = true
	if inviteOwner != nil {
		if inviteOwner.ID == userID {
			return nil, errors.New("unknown behaviour: you can't invite yourself")
		}

		// Update invite owner's code and mark that current user has used it.
		for i, inv := range inviteOwner.Invites {
			if inv.Code == inviteCode {
				inviteOwner.Invites[i].UsedBy = &currentUser.ID
				now := time.Now()
				inviteOwner.Invites[i].UsedAt = &now
			}
		}
		// set invited by only when owner exists
		currentUser.InvitedBy = inviteOwner.ID
	}

	log.Debug("generating invites for user")
	if err := s.generateInvitesForUser(ctx, currentUser); err != nil {
		return nil, fmt.Errorf("failed to generate invites for user: %w", err)
	}

	log.Debugf("invites generated: %d", len(currentUser.Invites))
	log.Debug("updating user with new info")
	if err := s.users.UpdateUser(ctx, currentUser.ID, *currentUser); err != nil {
		return nil, fmt.Errorf("failed to update given user: %w", err)
	}

	if inviteOwner != nil {
		log.Debugf("updating invite owner")
		if err := s.users.UpdateUser(ctx, inviteOwner.ID, *inviteOwner); err != nil {
			return nil, fmt.Errorf("failed to update invite owner: %w", err)
		}
	}

	return currentUser, nil
}

func (s *service) getOwnerFromCode(ctx context.Context, code string) (*internal.User, error) {
	if s.uniqueInviteCode != "" && code == s.uniqueInviteCode {
		logger.WithContext(ctx).Info("received unique invite code")
		return nil, nil
	}

	filter := internal.FindUserFilter{
		InviteActive: true,
		InviteCode:   &code,
	}

	inviteOwner, err := s.users.FindUser(ctx, filter)
	if err != nil {
		if err == internal.ErrNoUser {
			return nil, ErrNoInvite
		}

		return nil, errdefs.Unknown(err)
	}

	return inviteOwner, nil
}
