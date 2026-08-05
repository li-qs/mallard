package repo

import (
	"context"
	"fmt"
	"mallard/internal/domain/app"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	AppDB   = "mallard_tracing"
	AppColl = "app"
)

type AppRepo struct {
	coll *mongo.Collection
}

func NewAppRepo(client *mongo.Client) *AppRepo {
	return &AppRepo{coll: client.Database(AppDB).Collection(AppColl)}
}

func (r *AppRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "app_name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	})
	return err
}

func (r *AppRepo) GetByID(ctx context.Context, id bson.ObjectID) (*app.AppEntity, error) {
	var app app.AppEntity
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&app)
	return &app, err
}

func (r *AppRepo) CountApps(ctx context.Context, filter app.AppFilter) (int64, error) {
	match, err := buildAppMatch(filter)
	if err != nil {
		return 0, err
	}
	return r.coll.CountDocuments(ctx, match)
}

func (r *AppRepo) ListApps(ctx context.Context, filter app.AppFilter, page, pageSize int64) ([]app.AppEntity, error) {
	match, err := buildAppMatch(filter)
	if err != nil {
		return nil, err
	}

	skip := (page - 1) * pageSize
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(pageSize)

	c, err := r.coll.Find(ctx, match, findOptions)
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	var apps []app.AppEntity
	if err := c.All(ctx, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *AppRepo) Add(ctx context.Context, appEntity *app.AppEntity) error {
	res, err := r.coll.InsertOne(ctx, appEntity)
	if err != nil {
		return err
	}

	if IsDuplicateKey(err) {
		return app.ErrAppExists
	}

	id, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return fmt.Errorf("unexpected inserted id type: %T", res.InsertedID)
	}
	appEntity.ID = id
	return nil
}

func (r *AppRepo) UpdateSecret(ctx context.Context, id bson.ObjectID, secretHash string) error {
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"secret_hash": secretHash,
			"updated_at":  time.Now(),
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

func (r *AppRepo) UpdateIPAllowList(ctx context.Context, id bson.ObjectID, ipAllowList []string) error {
	res, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{
			"ip_allow_list": ipAllowList,
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

func (r *AppRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func buildAppMatch(filter app.AppFilter) (bson.D, error) {
	match := bson.D{}
	if filter.AppName != "" {
		match = append(match, bson.E{Key: "app_name", Value: bson.Regex{Pattern: regexp.QuoteMeta(filter.AppName), Options: "i"}})
	}
	if filter.ID != "" {
		id, err := bson.ObjectIDFromHex(strings.ToLower(strings.TrimPrefix(filter.ID, "0x")))
		if err != nil {
			return nil, err
		}
		match = append(match, bson.E{Key: "_id", Value: id})
	}
	return match, nil
}
