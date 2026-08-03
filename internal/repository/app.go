package repository

import (
	"context"
	"fmt"
	"mallard/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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

func (r *App) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.D{})
}

func (r *App) List(ctx context.Context, page, pageSize int64) ([]model.App, error) {
	skip := (page - 1) * pageSize
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	findOptions.SetSkip(skip)
	findOptions.SetLimit(pageSize)

	c, err := r.coll.Find(ctx, bson.D{}, findOptions)
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
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"secret_hash": secretHash})
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
		"$set": bson.M{"ip_allow_list": ipAllowList},
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
