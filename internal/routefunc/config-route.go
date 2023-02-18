package routefunc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func UpdateConfig(ctx *gin.Context) {
	if err := ValidateTableID(ctx); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	user, ok := ctx.MustGet("user").(*models.User)

	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "internal user was not passed, please check validate token",
		})
		return
	}
	gameID := GetGameID(ctx)
	request := models.Request{}

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	requestMap := map[string]any{}
	jsonRequestStr := utils.JSON(request)
	configs := []models.MemberConfig{}

	json.Unmarshal([]byte(jsonRequestStr), &requestMap)
	for k, v := range requestMap {
		configs = append(configs, models.MemberConfig{
			Name:   k,
			Value:  fmt.Sprintf("%v", v),
			UserID: user.ID,
			GameId: gameID,
		})
	}

	if err := db.Shared().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "game_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "ctime"}),
	}).Create(&configs).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, request)
}
