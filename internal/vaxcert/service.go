package vaxcert

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"

	"github.com/malyusha/immune-mosru-server/internal"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/redis"
)

const (
	certCodeLen = 12 // max length of generated certificate code

	certsCacheKeyPrefix = "certs"
)

type NewCertificateData struct {
	LastName    string
	FirstName   string
	SecondName  string
	DateOfBirth string
}

type Service interface {
	CreateVaxCert(ctx context.Context, data NewCertificateData) (*internal.VaxCert, error)
	GetCertByCode(ctx context.Context, code string) (*internal.VaxCert, error)
	ListCertificates(ctx context.Context) ([]*internal.VaxCert, error)
	UpdateCertByCode(ctx context.Context, code string, data internal.UpdateCertData) (*internal.VaxCert, error)
}

type service struct {
	cache  *redis.Cache
	logger logger.Logger
	certs  internal.CertificatesStorage
}

func NewService(certs internal.CertificatesStorage, cache *redis.Cache) *service {
	return &service{
		cache:  cache,
		logger: logger.With(logger.Fields{"package": "CERT_SERVICE"}),
		certs:  certs,
	}
}

func (s *service) UpdateCertByCode(ctx context.Context, code string, data internal.UpdateCertData) (*internal.VaxCert, error) {
	cert, err := s.certs.UpdateCert(ctx, code, data)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *service) ListCertificates(ctx context.Context) ([]*internal.VaxCert, error) {
	var res []*internal.VaxCert
	item := &cache.Item{
		Ctx:   ctx,
		Key:   getCertsListCacheKey(),
		Value: &res,
		TTL:   time.Hour,
		Do: func(item *cache.Item) (interface{}, error) {
			return s.certs.ListCertificates(item.Context())
		},
		SkipLocalCache: true,
	}

	if err := s.cache.Once(item); err != nil && err != cache.ErrCacheMiss {
		return nil, err
	}

	return res, nil
}

func (s *service) GetCertByCode(ctx context.Context, code string) (*internal.VaxCert, error) {
	cert := new(internal.VaxCert)
	item := &cache.Item{
		Ctx:   ctx,
		Key:   getCertByCodeCacheKey(code),
		Value: &cert,
		TTL:   time.Hour,
		Do: func(item *cache.Item) (interface{}, error) {
			return s.certs.GetByCode(ctx, code)
		},
		SkipLocalCache: true,
	}

	if err := s.cache.Once(item); err != nil && err != cache.ErrCacheMiss {
		return nil, err
	}

	return cert, nil
}

func getCertByCodeCacheKey(code string) string {
	return redis.WithCachePrefix(fmt.Sprintf("%s:%s", certsCacheKeyPrefix, code))
}

func getCertsListCacheKey() string {
	return redis.WithCachePrefix(certsCacheKeyPrefix + ":list")
}

func (s *service) CreateVaxCert(ctx context.Context, data NewCertificateData) (*internal.VaxCert, error) {
	log := s.logger.WithContext(ctx)
	log.Debug("creating new certificate")
	var cert *internal.VaxCert
	for cert == nil {
		code := generateCertCode(certCodeLen)
		exists, err := s.certs.Exists(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("failed to check whether certificate exists: %w", err)
		}

		if exists {
			continue
		}

		cert = createCertificateFromData(code, data)
		cert, err = s.certs.CreateCert(ctx, *cert)
		if err != nil {
			return nil, fmt.Errorf("failed to create certificate: %w", err)
		}
	}

	if err := s.clearCertsListCache(ctx); err != nil {
		log.Errorf("failed to clear list cache: %s", err)
	}
	log.With(logger.Fields{"code": cert.Code}).Info("certificate created")

	return cert, nil
}

// clears all cached entries for APQ key prefix.
func (s *service) clearCertsListCache(ctx context.Context) error {
	return redis.DeleteForPrefix(ctx, s.cache, getCertsListCacheKey())
}

func createCertificateFromData(code string, data NewCertificateData) *internal.VaxCert {
	return internal.CreateVaxCert(code, data.FirstName, data.LastName, data.SecondName, data.DateOfBirth)
}
