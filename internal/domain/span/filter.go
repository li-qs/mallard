package span

import "go.mongodb.org/mongo-driver/v2/bson"

type TraceFilter struct {
	AppID       *bson.ObjectID
	Operation   string
	Status      *int
	TraceID     string
	StartTimeGT int64
	StartTimeLT int64
}
