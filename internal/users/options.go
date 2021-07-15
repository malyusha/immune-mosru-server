package users

type Option func(s *service)

func noopOpt(s *service) {}

func WithUniqueCode(code string) Option {
	return func(s *service) {
		s.uniqueInviteCode = code
	}
}

func WithDefaultInviteCodeLen(l int) Option {
	if l == 0 {
		return noopOpt
	}
	return func(s *service) {
		s.inviteCodeLen = l
	}
}

func WithTelegramAdminId(id *int) Option {
	return func(s *service) {
		s.telegramAdminId = id
	}
}

func WithAdminInvites(num int) Option {
	return func(s *service) {
		s.invitesForAdmin = num
	}
}

func WithInvitesPerUser(num int) Option {
	return func(s *service) {
		s.invitesPerUser = num
	}
}


