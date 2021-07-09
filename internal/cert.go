package internal

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
)

type VaxCert struct {
	ID string
	Code,
	FirstName,
	LastName,
	SecondName,
	DateOfBirth string

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiringAt time.Time
}

// For Иванов Иван Иванович returns И***** И*** И*******
func (c *VaxCert) FullName() string {
	l, f, s := hideString(c.LastName), hideString(c.FirstName), hideString(c.SecondName)
	return strings.TrimRight(fmt.Sprintf("%s %s %s", l, f, s), " ")
}

func hideString(s string) string {
	// total length of string in runes
	length := utf8.RuneCountInString(s)
	// get only first char
	firstChar, _ := utf8.DecodeRuneInString(s)
	return fmt.Sprintf("%s%s", string(firstChar), strings.Repeat("*", length-1))
}

type UpdateCertData struct {
	FirstName,
	LastName,
	SecondName,
	DateOfBirth string
}

// CertificatesStorage represents storage interface for certificates storage.
type CertificatesStorage interface {
	ListCertificates(ctx context.Context) ([]*VaxCert, error)
	GetByCode(ctx context.Context, code string) (*VaxCert, error)
	Exists(ctx context.Context, code string) (bool, error)
	CreateCert(ctx context.Context, cert VaxCert) (*VaxCert, error)
	DeleteCert(ctx context.Context, code string) error
	UpdateCert(ctx context.Context, code string, cert UpdateCertData) (*VaxCert, error)
}

func CreateVaxCert(code, firstName, lastName, secondName, dateOfBirth string) *VaxCert {
	// random number of days with limit of 30
	extractDays := -1 * (time.Duration(rand.Intn(30)) * 24 * time.Hour)

	return &VaxCert{
		Code:        code,
		FirstName:   firstName,
		LastName:    lastName,
		SecondName:  secondName,
		DateOfBirth: dateOfBirth,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiringAt:  time.Now().Add(365 * 24 * time.Hour).Add(extractDays),
	}
}

var (
	ErrCertNotFound      = errdefs.NotFound(errors.New("certificate not found"))
	ErrCertAlreadyExists = errdefs.Validation(errors.New("certificate already exists"))
)
