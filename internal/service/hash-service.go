package service

import (
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"github.com/gin-gonic/gin"
)

type ACtiveQueuedHQ struct {
	Active int64
	Queued int64
}

type HashService interface {
	HashSave(models.Hash) models.Hash
	HashFindAll(ctx *gin.Context) []models.Hash
	HashSeqSave(models.Hash) models.Hash
	GetActiveQueuedHQ(map[string]interface{}) ACtiveQueuedHQ
	GetNextQueuedHashSequence(tID int64) models.HashSequence
	SetHashStatus(hashID int64, status string) models.Hash
	GetNextHashSequence(tID int64, lastSeq int, hashID int64) models.HashSequence
	IsActiveHash(hashID int64) (isActive bool)
}

type hashService struct {
	hash []models.Hash
}

func HashNew() HashService {
	return &hashService{}
}

func (service *hashService) HashSave(hash models.Hash) models.Hash {
	result := DB.Table("mini_game_minigamehash").Create(&hash)

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return hash
}

func (service *hashService) HashFindAll(ctx *gin.Context) []models.Hash {
	where := getWhere(ctx)
	result := DB.Table("mini_game_minigamehash").Where(where).Order("ctime desc").Limit(2).Find(&service.hash)

	hashes := []models.Hash{}
	for _, hash := range service.hash {
		DB.Table("mini_game_minigamehashsequence").Where("mini_game_hash_id = ?", hash.ID).Find(hash.Sequences)
		hashes = append(hashes, hash)
	}

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return hashes
}

func getWhere(ctx *gin.Context) map[string]interface{} {
	query := ctx.Request.URL.Query()

	where := map[string]interface{}{}
	if _, ok := query["id"]; ok {
		where["id"] = query["id"]
	}

	if _, ok := query["active"]; ok {
		where["active"] = query["active"]
	}

	return where
}

func (service *hashService) HashSeqSave(hash models.Hash) models.Hash {
	result := DB.Table("mini_game_minigamehashsequence").CreateInBatches(hash.Sequences, len(*hash.Sequences))

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return hash
}

func (service *hashService) GetActiveQueuedHQ(param map[string]interface{}) ACtiveQueuedHQ {
	where := map[string]interface{}{}
	if _, ok := param["mini_game_table_id"]; ok {
		where["mini_game_table_id"] = param["mini_game_table_id"]
	}

	var hq ACtiveQueuedHQ

	result := DB.Raw(`
		SELECT 
			(SELECT hq.SEQUENCE FROM	mini_game_event e
				JOIN mini_game_minigamehashsequence hq ON hq.ID = e.mini_game_hash_sequence_id
				JOIN mini_game_minigamehash h ON h.ID = hq.mini_game_hash_id 
			WHERE	h.mini_game_table_id = ?	AND h.status = 'active' 
			ORDER BY hq.SEQUENCE DESC LIMIT 1) active ,
			COUNT (h.status) FILTER (where h.status = 'queued') as queued
			FROM mini_game_minigamehash h
		LEFT JOIN mini_game_minigamehashsequence hq ON h.id = hq.mini_game_hash_id
		LEFT JOIN mini_game_event e ON e.mini_game_hash_sequence_id = hq.id 
		WHERE h.mini_game_table_id = ?
		GROUP BY h.mini_game_table_id`, where["mini_game_table_id"], where["mini_game_table_id"]).Scan(&hq)

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return hq
}

func (service *hashService) GetNextQueuedHashSequence(tID int64) models.HashSequence {
	var res models.HashSequence
	result := DB.Raw(`
				SELECT
					hs.*
				FROM
					mini_game_minigamehash mh
					JOIN mini_game_minigamehashsequence hs ON mh.id = hs.mini_game_hash_id
				WHERE
					mh.mini_game_table_id = ?
					AND mh.status = 'queued'
				ORDER BY
					mh.id ASC, hs.sequence ASC
					LIMIT 1`, tID).Scan(&res)

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return res
}

func (service *hashService) SetHashStatus(hashID int64, status string) models.Hash {
	var hs models.Hash
	payload := map[string]interface{}{
		"mtime":  time.Now(),
		"status": status,
	}

	DB.Table("mini_game_minigamehash").Model(&hs).Where("id = ?", hashID).Updates(payload)

	return hs
}

func (service *hashService) GetNextHashSequence(tID int64, lastSeq int, hashID int64) models.HashSequence {
	var res models.HashSequence
	result := DB.Raw(`
				SELECT
					hs.*
				FROM
					mini_game_minigamehash mh
					JOIN mini_game_minigamehashsequence hs ON mh.id = hs.mini_game_hash_id
				WHERE
					mh.mini_game_table_id = ?
					AND hs.sequence = ?
					AND hs.mini_game_hash_id = ?
					AND mh.status = 'active'
				ORDER BY
					mh.id DESC, hs.sequence ASC
					LIMIT 1`, tID, (lastSeq + 1), hashID).Scan(&res)

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return res
}

func (service *hashService) IsActiveHash(hashID int64) (isActive bool) {
	var count int
	res := DB.Raw(`
				SELECT COUNT
				( * ) 
			FROM
				mini_game_minigamehash 
			WHERE
				ID = ? 
				AND status = 'active'`, hashID).Scan(&count)
	if res.Error != nil {
		logger.Error(res.Error)
	}

	if count == 1 {
		isActive = true
	} else {
		isActive = false
	}

	return isActive
}
