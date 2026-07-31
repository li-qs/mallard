package repository

import (
	"myapi/internal/model"
	"myapi/pkg/mysql"
)

type RefreshToken struct {
	DB *mysql.DB
}

func (r *RefreshToken) Create(userID int64, tokenHash string, expiresAt int64) error {
	_, err := r.DB.Exec(
		"INSERT INTO `refresh_token` (user_id, token_hash, expires_at) VALUES (?, ?, FROM_UNIXTIME(?))",
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *RefreshToken) FindByHash(tokenHash string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.DB.Get(&rt, "SELECT * FROM `refresh_token` WHERE token_hash=?", tokenHash)
	return &rt, err
}

func (r *RefreshToken) Delete(id int64) error {
	_, err := r.DB.Exec("DELETE FROM `refresh_token` WHERE id=?", id)
	return err
}

func (r *RefreshToken) DeleteAllByUserID(userID int64) error {
	_, err := r.DB.Exec("DELETE FROM `refresh_token` WHERE user_id=?", userID)
	return err
}
