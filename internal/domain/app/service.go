package app

import (
	"context"
	"crypto/subtle"
	"mallard/pkg/utils"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	appRepo   AppRepoImpl
	options   *ServiceOptions
	appsCache *ristretto.Cache[string, AppEntity]
}

type ServiceOptions struct {
}

func NewService(appRepo AppRepoImpl, appsCache *ristretto.Cache[string, AppEntity], opts *ServiceOptions) *Service {
	return &Service{
		appRepo:   appRepo,
		options:   opts,
		appsCache: appsCache,
	}
}

func (s *Service) Add(ctx context.Context, appName string, ipAllowList []string) (*AppEntity, string, error) {
	secret, secretHash, err := s.generateSecret()
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	app := AppEntity{
		AppName:     appName,
		IPAllowList: ipAllowList,
		SecretHash:  secretHash,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.appRepo.Add(ctx, &app); err != nil {
		return nil, "", err
	}

	return &app, secret, nil
}

func (s *Service) List(ctx context.Context, filter AppFilter, page, pageSize int64) ([]AppEntity, int64, error) {
	apps, err := s.appRepo.ListApps(ctx, filter, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.appRepo.CountApps(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

func (s *Service) generateSecret() (secret string, secretHash string, err error) {
	secret, err = utils.GenerateRandomString(32)
	if err != nil {
		return "", "", err
	}
	secretHash = utils.SHA256Hex(secret)
	return
}

func (s *Service) UpdateSecret(ctx context.Context, id bson.ObjectID) (string, error) {
	secret, secretHash, err := s.generateSecret()
	if err != nil {
		return "", err
	}

	if err := s.appRepo.UpdateSecret(ctx, id, secretHash); err != nil {
		return "", err
	}
	s.appsCache.Del(id.Hex())

	return secret, nil
}

func (s *Service) ValidSecret(ctx context.Context, id bson.ObjectID, secret string) (*AppEntity, bool, error) {
	app, ok := s.appsCache.Get(id.Hex())
	if !ok {
		cached, err := s.appRepo.GetByID(ctx, id)
		if err != nil {
			return nil, false, err
		}
		app = *cached
		s.appsCache.Set(id.Hex(), app, 1)
	}

	sum := utils.SHA256Hex(secret)
	if subtle.ConstantTimeCompare([]byte(sum), []byte(app.SecretHash)) != 1 {
		return nil, false, ErrInvalidSecret
	}
	return &app, true, nil
}

func (s *Service) UpdateIPAllowList(ctx context.Context, id bson.ObjectID, ipAllowList []string) error {
	if err := s.appRepo.UpdateIPAllowList(ctx, id, ipAllowList); err != nil {
		return err
	}
	s.appsCache.Del(id.Hex())
	return nil
}

func (s *Service) Delete(ctx context.Context, id bson.ObjectID) error {
	if err := s.appRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.appsCache.Del(id.Hex())
	return nil
}
