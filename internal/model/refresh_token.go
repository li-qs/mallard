package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	RefreshTokenDB   = "mallard"
	RefreshTokenColl = "refresh_token"
)

type RefreshToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id,omitempty"`
	TokenHash string        `bson:"token_hash"`
	ExpiresAt time.Time     `bson:"expires_at"`
	CreatedAt time.Time     `bson:"created_at"`
}
