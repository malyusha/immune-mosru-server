package types

import (
	"errors"

	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
)

type Certificate struct {
	Name       string `json:"name"`
	DateBirth  string `json:"dateBirth"`
	ExpiringAt string `json:"expiringAt"`
}

func (c *NewCertificateData) Validate() error {
	err := errdefs.Validation(errors.New("invalid certificate data"))
	vErr := err.(errdefs.ErrValidation)

	if c.FirstName == "" {
		vErr.Fields().Add("firstName", "empty first name given")
	}

	if c.LastName == "" {
		vErr.Fields().Add("lastName", "empty last name given")
	}

	if c.SecondName == "" {
		vErr.Fields().Add("secondName", "empty second name given")
	}

	if c.DateBirth == "" {
		vErr.Fields().Add("dateBirth", "empty date of birth")
	}

	if len(vErr.Fields()) == 0 {
		return nil
	}

	return err
}

type NewCertificateData struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	SecondName string `json:"secondName"`
	DateBirth  string `json:"dateBirth"`
}
