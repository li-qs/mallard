package reqctx

import (
	"time"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AppCtx struct {
	ID          bson.ObjectID
	AppName     string
	SecretHash  string
	IPAllowList []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const appKey = "app"

func GetApp(c *echo.Context) (app *AppCtx, ok bool) {
	app, ok = c.Get(appKey).(*AppCtx)
	return
}

func SetApp(c *echo.Context, app *AppCtx) {
	c.Set(appKey, app)
}
