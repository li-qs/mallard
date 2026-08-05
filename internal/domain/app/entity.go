package app

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AppEntity struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	AppName     string        `bson:"app_name"`
	SecretHash  string        `bson:"secret_hash"`
	IPAllowList []string      `bson:"ip_allow_list"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}
