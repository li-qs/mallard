package repository

import (
	"context"
	"myapi/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type User struct {
	coll *mongo.Collection
}

func NewUser(client *mongo.Client) *User {
	return &User{
		coll: client.Database(model.RefreshTokenDB).Collection(model.RefreshTokenColl),
	}
}

func (r *User) GetByID(id bson.ObjectID) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (r *User) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	return &user, err
}

func (r *User) UpdatePasswordHash(id bson.ObjectID, passwordHash string) error {
	_, err := r.coll.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"password_hash": passwordHash})
	return err
}
