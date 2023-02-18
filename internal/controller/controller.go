package controller

import (
	"errors"
	"time"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/gin-gonic/gin"
)

const timeout = time.Duration(10_000 * time.Millisecond)

func Lock(forKey string) bool {
	val, _ := redis.Cache().Get(forKey)

	if val == "true" {
		return true
	}
	redis.Cache().Set(forKey, "true", timeout)
	return false
}

func Unlock(forKey string) {
	redis.Cache().Set(forKey, "", timeout)
}

func ValidateTableID(ctx *gin.Context) error {
	tableID := GetTableID(ctx)

	switch tableID {
	case constants_loltower.TableID,
		constants_lolcouple.TableID,
		constants_fifashootup.TableID,
		constants_fishprawncrab.TableID:
		return nil
	default:
		return errors.New("table id: " + string(types.Int(tableID).String()) + " is not supported")
	}
}

func GetTableID(ctx *gin.Context) int64 {
	return types.String(ctx.Param("table_id")).Int().Int64()
}

func GetGameID(ctx *gin.Context) int64 {
	tableID := GetTableID(ctx)

	switch tableID {
	case constants_loltower.TableID:
		return constants_loltower.GameID
	case constants_lolcouple.TableID:
		return constants_lolcouple.GameID
	case constants_fifashootup.TableID:
		return constants_fifashootup.GameID
	case constants_fishprawncrab.TableID:
		return constants_fishprawncrab.GameID
	default:
		panic("GetGameID table id: " + string(types.Int(tableID).String()) + " is not supported")
	}
}

func GetIdentifier(ctx *gin.Context) string {
	tableID := GetTableID(ctx)

	switch tableID {
	case constants_loltower.TableID:
		return constants_loltower.Identifier
	case constants_lolcouple.TableID:
		return constants_lolcouple.Identifier
	case constants_fifashootup.TableID:
		return constants_fifashootup.Identifier
	case constants_fishprawncrab.TableID:
		return constants_fishprawncrab.Identifier
	default:
		panic("GetIdentifier table id: " + string(types.Int(tableID).String()) + " is not supported")
	}
}
