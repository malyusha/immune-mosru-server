package immune

import (
	"context"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/malyusha/immune-mosru-server/api/http/types"
	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/httputils"
)

func (ir *immuneRouter) certificateByCode(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	params := httprouter.ParamsFromContext(ctx)
	code := params.ByName("code")
	if code == "" {
		return errdefs.InvalidParameter(errors.New("code not provided"))
	}

	cert, err := ir.vaxcert.GetCertByCode(ctx, code)
	if err != nil {
		return err
	}

	return httputils.JSONOk(w, mapCertificate(cert))
}

func (ir *immuneRouter) certificatesList(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	certs, err := ir.vaxcert.ListCertificates(ctx)
	if err != nil {
		return err
	}

	res := make([]types.Certificate, len(certs))
	for i := range certs {
		res[i] = mapCertificate(certs[i])
	}

	return httputils.JSONOk(w, res)
}

func (ir *immuneRouter) createCertificate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var data types.NewCertificateData
	if err := httputils.DecodeJSONRequest(r, &data); err != nil {
		return err
	}

	if err := data.Validate(); err != nil {
		// validation error here
		return err
	}

	certData := vaxcert.NewCertificateData{
		LastName:    data.LastName,
		FirstName:   data.FirstName,
		SecondName:  data.SecondName,
		DateOfBirth: data.DateBirth,
	}

	cert, err := ir.vaxcert.CreateVaxCert(ctx, certData)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusCreated, mapCertificate(cert))
}

func mapCertificate(c *internal.VaxCert) types.Certificate {
	return types.Certificate{
		Name:       c.FullName(),
		DateBirth:  c.DateOfBirth,
		ExpiringAt: c.ExpiringAt.Format("02.01.2006"),
	}
}
