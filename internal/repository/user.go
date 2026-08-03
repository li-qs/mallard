package repository

import (
	"context"
	"mallard/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
	coll *mongo.Collection
}

func NewUser(client *mongo.Client) *User {
	return &User{
		coll: client.Database(model.UserDB).Collection(model.UserColl),
	}
}

func (r *User) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	})
	return err
}

func (r *User) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (r *User) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	return &user, err
}

func (r *User) UpdatePasswordHash(ctx context.Context, id bson.ObjectID, passwordHash string) error {
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    time.Now(),
		},
	})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
