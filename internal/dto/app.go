package dto

type AddApp struct {
	AppName     string   `json:"app_name" validate:"required"`
	IPAllowList []string `json:"ip_allow_list"`
}

type UpdateApp struct {
	IPAllowList []string `json:"ip_allow_list"`
}
