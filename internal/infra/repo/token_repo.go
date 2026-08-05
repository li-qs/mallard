package repo

import (
	"context"
	"mallard/internal/domain/user"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	TokenDB   = "mallard"
	TokenColl = "refresh_token"
)

type TokenRepo struct {
	coll          *mongo.Collection
	expireSeconds int32
}

func NewTokenRepo(client *mongo.Client, expireSeconds int32) *TokenRepo {
	return &TokenRepo{
		coll:          client.Database(TokenDB).Collection(TokenColl),
		expireSeconds: expireSeconds,
	}
}

func (r *TokenRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetExpireAfterSeconds(r.expireSeconds),
		},
	})
	return err
}

func (r *TokenRepo) Create(ctx context.Context, userID bson.ObjectID, tokenHash string, expiresAt time.Time) error {
	_, err := r.coll.InsertOne(ctx, user.TokenEntity{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	return err
}

func (r *TokenRepo) GetByToken(ctx context.Context, tokenHash string) (*user.TokenEntity, error) {
	var rt user.TokenEntity
	err := r.coll.FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&rt)
	return &rt, err
}

func (r *TokenRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *TokenRepo) DeleteByUserID(ctx context.Context, userID bson.ObjectID) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
