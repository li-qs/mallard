package handler

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

func parsePagination(c *echo.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	pageSize, _ = strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return
}
