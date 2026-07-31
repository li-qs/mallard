package repository

import (
	"myapi/internal/model"
	"myapi/pkg/mysql"
)

type User struct {
	DB *mysql.DB
}

func (r *User) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.DB.Get(&user, "SELECT * FROM `user` WHERE id=?", id)
	return &user, err
}

func (r *User) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Get(&user, "SELECT * FROM `user` WHERE username=?", username)
	return &user, err
}

func (r *User) UpdateUsername(id int64, newUsername string) error {
	_, err := r.DB.Exec("UPDATE `user` SET username=? WHERE id=?", newUsername, id)
	return err
}

func (r *User) UpdatePasswordHash(id int64, passwordHash string) error {
	_, err := r.DB.Exec("UPDATE `user` SET password_hash=? WHERE id=?", passwordHash, id)
	return err
}
