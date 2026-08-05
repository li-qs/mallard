package user

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserRepoImpl interface {
	GetByID(ctx context.Context, userID bson.ObjectID) (*UserEntity, error)
	GetByUsername(ctx context.Context, username string) (*UserEntity, error)
	UpdatePasswordHash(ctx context.Context, userID bson.ObjectID, passwordHash string) error
}

type TokenRepoImpl interface {
	Create(ctx context.Context, userID bson.ObjectID, tokenHash string, expiresAt time.Time) error
	GetByToken(ctx context.Context, tokenHash string) (*TokenEntity, error)
	Delete(ctx context.Context, userID bson.ObjectID) error
	DeleteByUserID(ctx context.Context, userID bson.ObjectID) error
}
