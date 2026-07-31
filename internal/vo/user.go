package vo

import "myapi/internal/model"

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func ToUser(m *model.User) User {
	return User{
		ID:        m.ID.Hex(),
		Username:  m.Username,
		CreatedAt: m.CreatedAt.UnixMilli(),
		UpdatedAt: m.UpdatedAt.UnixMilli(),
	}
}
