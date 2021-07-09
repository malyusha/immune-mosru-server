package immune

import (
	"github.com/malyusha/immune-mosru-server/internal/vaxcert"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/router"
)

type immuneRouter struct {
	vaxcert vaxcert.Service
}

func (ir *immuneRouter) Routes() []router.Route {
	return []router.Route{
		router.NewGetRoute("/certs", ir.certificatesList),
		router.NewPostRoute("/certs", ir.createCertificate),
		router.NewGetRoute("/certs/:code", ir.certificateByCode),
	}
}

func NewRouter(certs vaxcert.Service) *immuneRouter {
	return &immuneRouter{
		vaxcert: certs,
	}
}
