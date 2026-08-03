package repository

import (
	"context"
	"errors"
	"fmt"
	"mallard/internal/model"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrInvalidIDPrefix = errors.New("invalid id prefix")

type App struct {
	coll *mongo.Collection
}

func NewApp(client *mongo.Client) *App {
	return &App{coll: client.Database(model.AppDB).Collection(model.AppColl)}
}

func (r *App) EnsureIndexes(ctx context.Context) error {
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

func (r *App) GetByID(ctx context.Context, id bson.ObjectID) (*model.App, error) {
	var app model.App
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&app)
	return &app, err
}

type AppFilter struct {
	AppName string
	ID      string
}

func (r *App) Count(ctx context.Context, filter AppFilter) (int64, error) {
	match, err := buildAppMatch(filter)
	if err != nil {
		return 0, err
	}
	return r.coll.CountDocuments(ctx, match)
}

func (r *App) List(ctx context.Context, filter AppFilter, page, pageSize int64) ([]model.App, error) {
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

	var apps []model.App
	if err := c.All(ctx, &apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *App) Add(ctx context.Context, app *model.App) error {
	res, err := r.coll.InsertOne(ctx, app)
	if err != nil {
		return err
	}
	id, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return fmt.Errorf("unexpected inserted id type: %T", res.InsertedID)
	}
	app.ID = id
	return nil
}

func (r *App) UpdateSecretHash(ctx context.Context, id bson.ObjectID, secretHash string) error {
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

func (r *App) UpdateIPAllowList(ctx context.Context, id bson.ObjectID, ipAllowList []string) error {
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

func (r *App) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func buildAppMatch(filter AppFilter) (bson.D, error) {
	match := bson.D{}
	if filter.AppName != "" {
		match = append(match, bson.E{Key: "app_name", Value: bson.Regex{Pattern: regexp.QuoteMeta(filter.AppName), Options: "i"}})
	}
	if filter.ID != "" {
		id, err := bson.ObjectIDFromHex(strings.ToLower(strings.TrimPrefix(filter.ID, "0x")))
		if err != nil {
			return nil, ErrInvalidIDPrefix
		}
		match = append(match, bson.E{Key: "_id", Value: id})
	}
	return match, nil
}
