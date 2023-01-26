package hooks

import (
	"backend/constants"
	"backend/helpers"
	"database/sql"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"math"
)

func BeforeCreateMatch(app core.App, e *core.RecordCreateEvent) error {
	return app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		var userId = e.Record.GetString("user")
		var winnerId = e.Record.GetString("winner")
		var loserId = e.Record.GetString("loser")

		// Load winner and loser
		var err error
		var winnerTea, loserTea *models.Record
		if winnerTea, err = txDao.FindRecordById(constants.TeasCollId, winnerId); err != nil {
			return err
		}

		if loserTea, err = txDao.FindRecordById(constants.TeasCollId, loserId); err != nil {
			return err
		}

		// Find/Create LocalRanks for winner and loser
		var localRanksColl *models.Collection
		var winnerLocalRankRow, loserLocalRankRow dbx.NullStringMap
		var winnerLocalRank, loserLocalRank *models.Record

		if localRanksColl, err = txDao.FindCollectionByNameOrId(constants.LocalRanksCollId); err != nil {
			return err
		}

		// winner
		if err = txDao.RecordQuery(localRanksColl).
			Where(dbx.HashExp{
				"user": userId,
				"tea":  winnerId,
			}).OrderBy("created ASC").One(&winnerLocalRankRow); err != nil {
			if err == sql.ErrNoRows {
				winnerLocalRank = models.NewRecord(localRanksColl)
				winnerLocalRank.Set("user", userId)
				winnerLocalRank.Set("rank", 1000)
				winnerLocalRank.Set("tea", winnerId)
				if err = txDao.SaveRecord(winnerLocalRank); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			winnerLocalRank = models.NewRecordFromNullStringMap(localRanksColl, winnerLocalRankRow)
		}

		// loser
		if err = txDao.RecordQuery(localRanksColl).
			Where(dbx.HashExp{
				"user": userId,
				"tea":  loserId,
			}).OrderBy("created ASC").One(&loserLocalRankRow); err != nil {
			if err == sql.ErrNoRows {
				loserLocalRank = models.NewRecord(localRanksColl)
				loserLocalRank.Set("user", userId)
				loserLocalRank.Set("rank", 1000)
				loserLocalRank.Set("tea", loserId)
				if err = txDao.SaveRecord(loserLocalRank); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			loserLocalRank = models.NewRecordFromNullStringMap(localRanksColl, loserLocalRankRow)
		}

		// Calculate RankChange for this match
		var oldWinnerRankVal = winnerLocalRank.GetFloat("rank")
		var oldLoserRankVal = loserLocalRank.GetFloat("rank")
		var rankChange = helpers.CalculateRankChange(oldWinnerRankVal, oldLoserRankVal)
		var newWinnerRankVal = math.Round(oldWinnerRankVal + rankChange)
		var newLoserRankVal = math.Round(oldLoserRankVal - rankChange)
		e.Record.Set("rank_change", rankChange)

		// Apply RankChange to winner and loser LocalRanks
		winnerLocalRank.Set("rank", newWinnerRankVal)
		loserLocalRank.Set("rank", newLoserRankVal)

		if err = txDao.SaveRecord(winnerLocalRank); err != nil {
			return err
		}

		if err = txDao.SaveRecord(loserLocalRank); err != nil {
			return err
		}

		// Update/Create GlobalRank for winner and loser
		var winnerGlobalRankId = winnerTea.GetString("global_rank")
		var loserGlobalRankId = loserTea.GetString("global_rank")
		var winnerGlobalRank, loserGlobalRank *models.Record

		if winnerGlobalRankId == "" {
			// Create GlobalRank for winner
			var globalRankColl *models.Collection

			if globalRankColl, err = txDao.FindCollectionByNameOrId(constants.GlobalRanksCollId); err != nil {
				return err
			}

			winnerGlobalRank = models.NewRecord(globalRankColl)
			winnerGlobalRank.Set("rank", newWinnerRankVal)
			if err = txDao.SaveRecord(winnerGlobalRank); err != nil {
				return err
			}

			// Update winner to point to new GlobalRank
			winnerTea.Set("global_rank", winnerGlobalRank.GetId())
			if err = txDao.SaveRecord(winnerTea); err != nil {
				return err
			}
		} else {
			// Update GlobalRank for winner
			var winnerAvgRank float64

			if winnerAvgRank, err = helpers.CalculateAverageRank(winnerId, txDao); err != nil {
				return err
			}

			if winnerGlobalRank, err = txDao.FindRecordById(constants.GlobalRanksCollId, winnerGlobalRankId); err != nil {
				return err
			}

			winnerGlobalRank.Set("rank", winnerAvgRank)
			if err = txDao.SaveRecord(winnerGlobalRank); err != nil {
				return err
			}
		}

		if loserGlobalRankId == "" {
			// Create GlobalRank for loser
			var globalRankColl *models.Collection

			if globalRankColl, err = txDao.FindCollectionByNameOrId(constants.GlobalRanksCollId); err != nil {
				return err
			}

			loserGlobalRank = models.NewRecord(globalRankColl)
			loserGlobalRank.Set("rank", newLoserRankVal)
			if err = txDao.SaveRecord(loserGlobalRank); err != nil {
				return err
			}

			// Update loser to point to new GlobalRank
			loserTea.Set("global_rank", loserGlobalRank.GetId())
			if err = txDao.SaveRecord(loserTea); err != nil {
				return err
			}
		} else {
			// Update GlobalRank for loser
			var loserAvgRank float64

			if loserAvgRank, err = helpers.CalculateAverageRank(loserId, txDao); err != nil {
				return err
			}

			if loserGlobalRank, err = txDao.FindRecordById(constants.GlobalRanksCollId, loserGlobalRankId); err != nil {
				return err
			}

			loserGlobalRank.Set("rank", loserAvgRank)
			if err = txDao.SaveRecord(loserGlobalRank); err != nil {
				return err
			}
		}

		return nil
	})
}
