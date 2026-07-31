package repository

import (
	"context"
	"myapi/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RefreshToken struct {
	coll *mongo.Collection
}

func NewRefreshToken(client *mongo.Client) *RefreshToken {
	return &RefreshToken{
		coll: client.Database(model.RefreshTokenDB).Collection(model.RefreshTokenColl),
	}
}

func (r *RefreshToken) Create(userID bson.ObjectID, tokenHash string, expiresAt time.Time) error {
	_, err := r.coll.InsertOne(context.Background(), model.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	return err
}

func (r *RefreshToken) FindByHash(tokenHash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.coll.FindOne(context.Background(), bson.M{"token_hash": tokenHash}).Decode(&rt)
	return &rt, err
}

func (r *RefreshToken) Delete(id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

func (r *RefreshToken) DeleteAllByUserID(userID bson.ObjectID) error {
	_, err := r.coll.DeleteMany(context.Background(), bson.M{"user_id": userID})
	return err
}
