package repository

import (
	"context"
	"mallard/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type RefreshToken struct {
	coll          *mongo.Collection
	expireSeconds int32
}

func NewRefreshToken(client *mongo.Client, expireSeconds int32) *RefreshToken {
	return &RefreshToken{
		coll:          client.Database(model.RefreshTokenDB).Collection(model.RefreshTokenColl),
		expireSeconds: expireSeconds,
	}
}

func (r *RefreshToken) EnsureIndexes(ctx context.Context) error {
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

func (r *RefreshToken) Create(ctx context.Context, userID bson.ObjectID, tokenHash string, expiresAt time.Time) error {
	_, err := r.coll.InsertOne(ctx, model.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	return err
}

func (r *RefreshToken) FindByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.coll.FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&rt)
	return &rt, err
}

func (r *RefreshToken) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *RefreshToken) DeleteAllByUserID(ctx context.Context, userID bson.ObjectID) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
