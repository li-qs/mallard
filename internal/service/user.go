package service

import (
	"myapi/internal/model"
	"myapi/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserRepo *repository.User
}

func (s *User) GetByID(id int64) (*model.User, error) {
	return s.UserRepo.GetByID(id)
}

func (s *User) GetByUsername(username string) (*model.User, error) {
	return s.UserRepo.GetByUsername(username)
}

func (s *User) UpdatePassword(id int64, pwd string) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	if err != nil {
		return err
	}
	return s.UserRepo.UpdatePasswordHash(id, string(pwdHash))
}
