package web

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
)

func SPA(fsys fs.FS) echo.HandlerFunc {
	fileServer := http.StripPrefix("/", http.FileServer(http.FS(fsys)))

	return func(c *echo.Context) error {
		p := strings.TrimPrefix(c.Param("*"), "/")

		if strings.HasPrefix(p, "api/") {
			return echo.ErrNotFound
		}

		if p != "" {
			if f, err := fsys.Open(path.Clean(p)); err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Response(), c.Request())
				return nil
			}
		}

		index, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		return c.Blob(http.StatusOK, "text/html; charset=utf-8", index)
	}
}
