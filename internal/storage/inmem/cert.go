package inmem

import (
	"context"
	"errors"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/pkg/database/inmem"
)

type storage struct {
	table *inmem.Table
}

func (s *storage) Exists(ctx context.Context, code string) (bool, error) {
	_, err := s.table.Get(code)
	if err != nil {
		if err == inmem.ErrRowNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *storage) ListCertificates(ctx context.Context) ([]*internal.VaxCert, error) {
	rows := s.table.List()
	list := make([]*internal.VaxCert, len(rows))
	for i, row := range rows {
		list[i] = row.(*internal.VaxCert)
	}

	return list, nil
}

func (s *storage) GetByCode(ctx context.Context, code string) (*internal.VaxCert, error) {
	row, err := s.table.Get(code)
	if err != nil {
		if err == inmem.ErrRowNotFound {
			return nil, internal.ErrCertNotFound
		}

		return nil, err
	}

	return row.(*internal.VaxCert), nil
}

func (s *storage) CreateCert(ctx context.Context, cert internal.VaxCert) (*internal.VaxCert, error) {
	if cert.Code == "" {
		return nil, errors.New("empty code provided")
	}

	if c, _ := s.GetByCode(ctx, cert.Code); c != nil {
		return nil, internal.ErrCertAlreadyExists
	}

	if err := s.table.Write(cert.Code, &cert); err != nil {
		return nil, err
	}

	return &cert, nil
}

// DeleteCert deletes certificate with given code from list.
func (s *storage) DeleteCert(ctx context.Context, code string) error {
	s.table.Delete(code)
	return nil
}

// Updates certificate with given code with given data.
func (s *storage) UpdateCert(ctx context.Context, code string, data internal.UpdateCertData) (*internal.VaxCert, error) {
	row, err := s.table.Get(code)
	if err != nil {
		return nil, err
	}

	cert := row.(*internal.VaxCert)
	cert.FirstName = data.FirstName
	cert.LastName = data.LastName
	cert.SecondName = data.SecondName
	cert.DateOfBirth = data.DateOfBirth

	return cert, nil
}

func NewCertificatesStorage(db *inmem.DB) *storage {
	return &storage{db.Table("certificates")}
}
