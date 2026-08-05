package app

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AppRepoImpl interface {
	Add(ctx context.Context, app *AppEntity) error
	ListApps(ctx context.Context, filter AppFilter, page int64, pageSize int64) ([]AppEntity, error)
	CountApps(ctx context.Context, filter AppFilter) (int64, error)
	UpdateSecret(ctx context.Context, id bson.ObjectID, secretHash string) error
	GetByID(ctx context.Context, id bson.ObjectID) (*AppEntity, error)
	UpdateIPAllowList(ctx context.Context, id bson.ObjectID, ipAllowList []string) error
	Delete(ctx context.Context, id bson.ObjectID) error
}
