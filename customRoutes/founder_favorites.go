package customRoutes

import (
	"backend/constants"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"net/http"
)

func FounderFavorites(app core.App, c echo.Context) error {
	// TODO: actually get the 3 highest rated teas for the three founding users
	var err error
	var teasColl *models.Collection
	var gavynTea, carissaTea, kenzieTea *models.Record

	if teasColl, err = app.Dao().FindCollectionByNameOrId(constants.TeasCollId); err != nil {
		return err
	}

	gavynTea = models.NewRecord(teasColl)
	gavynTea.Set("name", "Raspberry Mint")
	gavynTea.Set("Description", "A unique blend of raspberries, blackberry leaves, peppermint, and apple bits gives a sweet, refreshing aroma with a lightly tart, fruity flavor and a beautiful, light, ruby-red color.")

	carissaTea = models.NewRecord(teasColl)
	carissaTea.Set("name", "Strawberry Valley")
	carissaTea.Set("description", "Dried strawberries with a unique honey-sweet note. Great for an afternoon dessert tea option.")

	kenzieTea = models.NewRecord(teasColl)
	kenzieTea.Set("name", "Hazelnut Chai")
	kenzieTea.Set("description", "Toasted hazelnut aroma with a rich, creamy, and full-bodied Assam tea.")

	var teas = []*models.Record{gavynTea, carissaTea, kenzieTea}
	if err = apis.EnrichRecords(c, app.Dao(), teas); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, teas)
}
