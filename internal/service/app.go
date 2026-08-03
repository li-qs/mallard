package service

import (
	"context"
	"crypto/subtle"
	"fmt"
	"mallard/internal/model"
	"mallard/internal/repository"
	"mallard/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type App struct {
	appRepo *repository.App
	cache   *ttlCache[string, model.App]
}

var (
	ErrInvalidSecret = fmt.Errorf("invalid secret")
	ErrAppExists     = fmt.Errorf("app already exists")
)

func NewApp(appRepo *repository.App) *App {
	return &App{
		appRepo: appRepo,
		cache:   newTTLCache[string, model.App](time.Minute),
	}
}

func (s *App) Add(ctx context.Context, appName string, ipAllowList []string) (bson.ObjectID, string, error) {
	secret, secretHash, err := s.generateSecret()
	if err != nil {
		return bson.NilObjectID, "", err
	}

	now := time.Now()
	app := model.App{
		AppName:     appName,
		IPAllowList: ipAllowList,
		SecretHash:  secretHash,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.appRepo.Add(ctx, &app); err != nil {
		if repository.IsDuplicateKey(err) {
			return bson.NilObjectID, "", ErrAppExists
		}
		return bson.NilObjectID, "", err
	}

	return app.ID, secret, nil
}

func (s *App) List(ctx context.Context, page, pageSize int64) ([]model.App, int64, error) {
	apps, err := s.appRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.appRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

func (s *App) generateSecret() (secret string, secretHash string, err error) {
	secret, err = utils.GenerateRandomString(32)
	if err != nil {
		return "", "", err
	}
	secretHash = utils.SHA256Hex(secret)
	return
}

func (s *App) UpdateSecret(ctx context.Context, id bson.ObjectID) (string, error) {
	secret, secretHash, err := s.generateSecret()
	if err != nil {
		return "", err
	}

	if err := s.appRepo.UpdateSecretHash(ctx, id, secretHash); err != nil {
		return "", err
	}
	s.cache.Delete(id.Hex())

	return secret, nil
}

func (s *App) ValidSecret(ctx context.Context, id bson.ObjectID, secret string) (*model.App, bool, error) {
	key := id.Hex()

	app, ok := s.cache.Get(key)
	if !ok {
		cached, err := s.appRepo.GetByID(ctx, id)
		if err != nil {
			return nil, false, err
		}
		app = *cached
		s.cache.Set(key, app)
	}

	sum := utils.SHA256Hex(secret)
	if subtle.ConstantTimeCompare([]byte(sum), []byte(app.SecretHash)) != 1 {
		return nil, false, ErrInvalidSecret
	}
	return &app, true, nil
}

func (s *App) UpdateIPAllowList(ctx context.Context, id bson.ObjectID, ipAllowList []string) error {
	if err := s.appRepo.UpdateIPAllowList(ctx, id, ipAllowList); err != nil {
		return err
	}
	s.cache.Delete(id.Hex())
	return nil
}

func (s *App) Delete(ctx context.Context, id bson.ObjectID) error {
	if err := s.appRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.cache.Delete(id.Hex())
	return nil
}
