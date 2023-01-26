package helpers

import (
	"backend/constants"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"math"
)

func CalculateRankChange(winnerRank, loserRank float64) float64 {
	var rankDifference = loserRank - winnerRank
	var rankChange = math.Round(1 + 39/(1+math.Pow(math.E, -rankDifference/20)))
	return rankChange
}

func CalculateAverageRank(teaId string, dao *daos.Dao) (float64, error) {
	var err error
	var localRanksColl *models.Collection

	if localRanksColl, err = dao.FindCollectionByNameOrId(constants.LocalRanksCollId); err != nil {
		return 0, err
	}

	var averageRank = struct {
		AverageRank float64 `db:"avg_rank"`
	}{}

	if err = dao.RecordQuery(localRanksColl).
		Select("AVG(rank) as avg_rank").
		Where(dbx.HashExp{
			"tea": teaId,
		}).One(&averageRank); err != nil {
		return 0, err
	}

	return math.Round(averageRank.AverageRank), nil
}
