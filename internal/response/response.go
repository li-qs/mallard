package response

import (
	"net/http"

	"mallard/internal/vo"

	"github.com/labstack/echo/v5"
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type ListData[T any] struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	List     T   `json:"list"`
}

func JsonData[T any](c *echo.Context, data T) error {
	return c.JSON(http.StatusOK, &Response[T]{
		Code: 0,
		Data: data,
	})
}

func JsonList[T any](c *echo.Context, l T, p vo.Pagination) error {
	return c.JSON(http.StatusOK, &Response[ListData[T]]{
		Code: 0,
		Data: ListData[T]{
			List:     l,
			Page:     p.Page,
			PageSize: p.PageSize,
			Total:    p.Total,
		},
	})
}

func JsonError(c *echo.Context, code int, msg string) error {
	return c.JSON(http.StatusOK, &Response[any]{
		Code:    code,
		Message: msg,
	})
}
