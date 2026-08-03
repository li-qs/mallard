package vo

import "mallard/internal/model"

type App struct {
	ID          string   `json:"id"`
	AppName     string   `json:"app_name"`
	IPAllowList []string `json:"ip_allow_list"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

func ToApp(m *model.App) App {
	return App{
		ID:          m.ID.Hex(),
		AppName:     m.AppName,
		IPAllowList: m.IPAllowList,
		CreatedAt:   m.CreatedAt.UnixMilli(),
		UpdatedAt:   m.UpdatedAt.UnixMilli(),
	}
}

type AppAdd struct {
	ID          string   `json:"id"`
	AppName     string   `json:"app_name"`
	Secret      string   `json:"secret"`
	IPAllowList []string `json:"ip_allow_list"`
}

type AppSecret struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}
