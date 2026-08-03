package repository

import (
	"context"
	"mallard/internal/model"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Span struct {
	coll          *mongo.Collection
	expireSeconds int32
}

func NewSpan(client *mongo.Client, expireSeconds int32) *Span {
	return &Span{
		coll:          client.Database(model.SpanDB).Collection(model.SpanColl),
		expireSeconds: expireSeconds,
	}
}

func (r *Span) InsertMany(ctx context.Context, spans []model.Span) error {
	if len(spans) == 0 {
		return nil
	}

	docs := make([]any, len(spans))
	for i := range spans {
		docs[i] = spans[i]
	}

	_, err := r.coll.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	if err != nil && !IsDuplicateKey(err) {
		return err
	}
	return nil
}

func (r *Span) FindByTraceID(ctx context.Context, traceID string) ([]model.Span, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "start_time", Value: 1}})

	c, err := r.coll.Find(ctx, bson.M{"trace_id": traceID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	var spans []model.Span
	if err := c.All(ctx, &spans); err != nil {
		return nil, err
	}
	return spans, nil
}

type TraceFilter struct {
	AppID       *bson.ObjectID
	Operation   string
	Status      *int
	TraceID     string
	StartTimeGT int64
	StartTimeLT int64
}

func (r *Span) ListTraces(ctx context.Context, filter TraceFilter, skip, limit int64) ([]model.TraceSummary, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: buildTraceMatch(filter)}},
		{{Key: "$sort", Value: bson.D{{Key: "start_time", Value: 1}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$trace_id"},
			{Key: "app_ids", Value: bson.M{"$addToSet": "$app_id"}},
			{Key: "operation", Value: bson.M{"$first": "$operation"}},
			{Key: "start_time", Value: bson.M{"$min": "$start_time"}},
			{Key: "end_time", Value: bson.M{"$max": bson.M{"$add": bson.A{"$start_time", "$duration"}}}},
			{Key: "span_count", Value: bson.M{"$sum": 1}},
			{Key: "error_count", Value: errorCountExpr},
		}}},
	}

	if m := buildErrorMatch(filter); m != nil {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: m}})
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "trace_id", Value: "$_id"},
			{Key: "app_ids", Value: 1},
			{Key: "operation", Value: 1},
			{Key: "start_time", Value: 1},
			{Key: "duration", Value: bson.M{"$subtract": bson.A{"$end_time", "$start_time"}}},
			{Key: "span_count", Value: 1},
			{Key: "error_count", Value: 1},
			{Key: "_id", Value: 0},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "start_time", Value: -1}}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	c, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	var traces []model.TraceSummary
	if err := c.All(ctx, &traces); err != nil {
		return nil, err
	}
	return traces, nil
}

func (r *Span) CountTraces(ctx context.Context, filter TraceFilter) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: buildTraceMatch(filter)}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$trace_id"},
			{Key: "error_count", Value: errorCountExpr},
		}}},
	}

	if m := buildErrorMatch(filter); m != nil {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: m}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$count", Value: "total"}})

	c, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer c.Close(ctx)

	var result []struct {
		Total int64 `bson:"total"`
	}
	if err := c.All(ctx, &result); err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Total, nil
}

var errorCountExpr = bson.M{
	"$sum": bson.M{"$cond": bson.A{bson.M{"$ne": bson.A{"$status", 0}}, 1, 0}},
}

func buildErrorMatch(filter TraceFilter) bson.M {
	if filter.Status == nil {
		return nil
	}
	switch *filter.Status {
	case 1: // 成功：无错误 span
		return bson.M{"error_count": 0}
	case 2: // 错误：存在错误 span
		return bson.M{"error_count": bson.M{"$gt": 0}}
	default:
		return nil
	}
}

func buildTraceMatch(filter TraceFilter) bson.D {
	match := bson.D{}
	if filter.AppID != nil {
		match = append(match, bson.E{Key: "app_id", Value: *filter.AppID})
	}
	if filter.Operation != "" {
		match = append(match, bson.E{Key: "operation", Value: filter.Operation})
	}
	if filter.TraceID != "" {
		match = append(match, bson.E{Key: "trace_id", Value: bson.Regex{Pattern: "^" + regexp.QuoteMeta(filter.TraceID)}})
	}

	timeConds := bson.A{}
	if filter.StartTimeGT > 0 {
		timeConds = append(timeConds, bson.D{{Key: "start_time", Value: bson.M{"$gte": filter.StartTimeGT}}})
	}
	if filter.StartTimeLT > 0 {
		timeConds = append(timeConds, bson.D{{Key: "start_time", Value: bson.M{"$lt": filter.StartTimeLT}}})
	}
	if len(timeConds) > 0 {
		match = append(match, bson.E{Key: "$and", Value: timeConds})
	}
	return match
}

func (r *Span) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "trace_id", Value: 1}, {Key: "start_time", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "span_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "parent_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "app_id", Value: 1}, {Key: "start_time", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "start_time", Value: -1}},
		},
		{
			Keys:    bson.D{{Key: "reported_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(r.expireSeconds),
		},
	})
	return err
}
