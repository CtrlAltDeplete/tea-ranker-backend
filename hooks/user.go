package hooks

import (
	"backend/constants"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

func BeforeCreateUser(app core.App, e *core.RecordCreateEvent) error {
	return app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		var err error
		var teasColl, notesColl *models.Collection
		var teas []*models.Record
		var userId = e.Record.GetId()

		// Retrieve all the teas
		if teasColl, err = txDao.FindCollectionByNameOrId(constants.TeasCollId); err != nil {
			return err
		}

		if err = txDao.RecordQuery(teasColl).All(&teas); err != nil {
			return err
		}

		// Create a Note for each tea
		if notesColl, err = txDao.FindCollectionByNameOrId(constants.NotesCollId); err != nil {
			return err
		}

		for _, tea := range teas {
			record := models.NewRecord(notesColl)
			record.Set("user", userId)
			record.Set("notes", "")
			record.Set("tea", tea.GetId())
			if err = txDao.SaveRecord(record); err != nil {
				return err
			}
		}

		return nil
	})
}
