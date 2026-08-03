package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
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
