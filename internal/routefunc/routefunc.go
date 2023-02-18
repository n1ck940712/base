package routefunc

import (
	"errors"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/gin-gonic/gin"
)

func ValidateTableID(ctx *gin.Context) error {
	tableID := GetTableID(ctx)

	switch tableID {
	case constants_loltower.TableID, constants_lolcouple.TableID, constants_fifashootup.TableID:
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
	}
	panic("must call ValidateTableID")
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
	}
	panic("must call ValidateTableID")
}
