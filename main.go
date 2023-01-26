package main

import (
	"backend/customRoutes"
	"backend/hooks"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"net/http"
)

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		_, err := e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/founder_favorites",
			Handler: func(context echo.Context) error {
				return customRoutes.FounderFavorites(app, context)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
			Name: "founder_favorites",
		})

		return err
	})

	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		switch e.Record.Collection().Name {
		case "matches":
			return hooks.BeforeCreateMatch(app, e)
		case "users":
			return hooks.BeforeCreateUser(app, e)
		default:
			return nil
		}
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
