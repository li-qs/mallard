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
	UserDB   = "mallard"
	UserColl = "user"
)

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Username     string        `bson:"username"`
	PasswordHash string        `bson:"password_hash"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

type UserRepo struct {
	coll *mongo.Collection
}

func NewUserRepo(client *mongo.Client) *UserRepo {
	return &UserRepo{
		coll: client.Database(UserDB).Collection(UserColl),
	}
}

func (r *UserRepo) EnsureIndexes(ctx context.Context) error {
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

func (r *UserRepo) GetByID(ctx context.Context, id bson.ObjectID) (*user.UserEntity, error) {
	var user user.UserEntity
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*user.UserEntity, error) {
	var user user.UserEntity
	err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	return &user, err
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id bson.ObjectID, passwordHash string) error {
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
