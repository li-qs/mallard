package service

import (
	"context"
	"mallard/internal/model"
	"mallard/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserRepo *repository.User
}

func (s *User) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	return s.UserRepo.GetByID(ctx, id)
}

func (s *User) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.UserRepo.GetByUsername(ctx, username)
}

func (s *User) UpdatePassword(ctx context.Context, id bson.ObjectID, pwd string) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	if err != nil {
		return err
	}
	return s.UserRepo.UpdatePasswordHash(ctx, id, string(pwdHash))
}
