package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type IndexEnsurer interface {
	EnsureIndexes(context.Context) error
}

func Migrate(ctx context.Context, migrators ...IndexEnsurer) error {
	for _, m := range migrators {
		if err := m.EnsureIndexes(ctx); err != nil {
			return err
		}
	}
	return nil
}

func IsDuplicateKey(err error) bool {
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, code := range we.ErrorCodes() {
			if code == 11000 {
				return true
			}
		}
	}
	var bwe mongo.BulkWriteException
	if errors.As(err, &bwe) {
		for _, code := range bwe.ErrorCodes() {
			if code == 11000 {
				return true
			}
		}
	}
	return false
}
