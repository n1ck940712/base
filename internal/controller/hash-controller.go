package controller

import (
	"crypto/sha256"
	"fmt"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"

	"github.com/gin-gonic/gin"
)

const LolTurretGameID = 8
const DefaultSecquenceCount = 100
const HashDefaultVal = "queued"

type HashController interface {
	HashFindAll(ctx *gin.Context) []models.Hash
	HashSave(ctx *gin.Context) ([]models.Hash, error)
}

type hController struct {
	service  service.HashService
	gService service.GameService
	eService service.EventService
}

func NewHashController(service service.HashService, gService service.GameService, eService service.EventService) HashController {
	return &hController{
		service:  service,
		gService: gService,
		eService: eService,
	}
}

func (c *hController) HashFindAll(ctx *gin.Context) []models.Hash {
	return c.service.HashFindAll(ctx)
}

func (c *hController) HashSave(ctx *gin.Context) ([]models.Hash, error) {
	var gameID int64
	fmt.Sscan(ctx.Param("game_id"), &gameID)

	var hash []models.Hash

	games, _ := c.gService.GetGames(gameID)
	maxHQ := settings.DEF_MAX_SEQUENCE

	for _, game := range *games.GameTables {
		temp := map[string]interface{}{}
		temp["mini_game_table_id"] = game.ID

		eHash := c.eService.GetLatestHashSeq(temp)
		hStatusCnt := c.service.GetActiveQueuedHQ(temp)

		logger.Info(gameID, " latest hash: ", utils.PrettyJSON(eHash))
		logger.Info(gameID, " hStatusCnt: ", utils.PrettyJSON(hStatusCnt))
		if (eHash.ID == 0 && eHash.Sequence == 0) || //hash is empty TODO: update to nil checking
			(eHash.Sequence >= maxHQ && hStatusCnt.Queued != 0) || // create hash base on max hash sequence
			(hStatusCnt.Active == 0 && hStatusCnt.Queued == 0) || // create hash if there's no active hash header
			(hStatusCnt.Active > 1 && hStatusCnt.Queued == 0) { // makes sure there's a queued hash sequence
			hashes := models.Hash{
				Seed:    generateHash(),
				GameID:  gameID,
				TableID: game.ID,
				Status:  HashDefaultVal,
			}

			res := c.service.HashSave(hashes)

			hashSequence := hashSeqSave(res)
			res.Sequences = &hashSequence
			res = c.service.HashSeqSave(res)

			hash = append(hash, res)
		}
	}
	logger.Info(gameID, " created hashes: ", len(hash))
	return hash, nil
}

func hashSeqSave(res models.Hash) []models.HashSequence {

	hashSequence := []models.HashSequence{}
	defHQ := settings.DEF_HASH_SEQUENCE_COUNT
	for i := 1; i <= defHQ; i++ {
		temp := models.HashSequence{
			Value:    generateHash(),
			HashID:   res.ID,
			Sequence: i,
		}

		hashSequence = append(hashSequence, temp)
	}

	return hashSequence
}

func generateHash() string {
	dt := time.Now()
	str := dt.String() + "h@sh-go-mg--r4pid"
	hash := sha256.Sum256([]byte(str))
	strHash := fmt.Sprintf("%x", hash)

	return strHash
}
