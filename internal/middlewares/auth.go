package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/bind"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/gin-gonic/gin"
)

type TokenType string

const (
	TokenTypeUser   = "user"
	TokenTypeServer = "server"
	TokenTypeMember = "member"
)

func Authenticate(tokenTypes ...TokenType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := bind.Header{}

		if qErr := ctx.ShouldBindQuery(&header); qErr != nil {
			logger.Warning("unable to bind query error: ", qErr.Error())
		}
		if header.Token == "" {
			if hErr := ctx.ShouldBindHeader(&header); hErr != nil {
				logger.Warning("unable to bind header error: ", hErr.Error())
			}
		}
		if header.Token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		header.Token = strings.ReplaceAll(header.Token, "Token ", "") //replace Token keyword
		validateToken := bind.ValidateToken{}

		if err := api.NewAPI(settings.GetEBOAPI().String() + "/v1/validate-token/").
			SetIdentifier("api middleware authenticate").
			AddHeaders(map[string]string{
				"User-Agent":    settings.GetUserAgent().String(),
				"Authorization": settings.GetServerToken().String(),
				"Content-Type":  "application/json",
			}).
			AddBody(map[string]string{
				"token": header.Token,
			}).
			Post(&validateToken); err != nil {
			if err.GetResponse() != nil && err.GetResponse().StatusCode >= 500 {
				go slack.SendPayload(slack.NewLootboxNotification(
					"api middleware authenticate",
					fmt.Sprint("*Status Code:* \n> *", err.GetResponse().StatusCode, "*\n*Response Body:* \n> *", err.GetResponseBody(), "*"),
				), slack.LootboxHealthCheck)
			}
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		if len(tokenTypes) > 0 && !types.Array[TokenType](tokenTypes).Constains(TokenType(validateToken.Type)) {
			logger.Warning("supported token types are ", tokenTypes, " token is ", validateToken.Type)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		switch validateToken.Type {
		case TokenTypeMember:
			user := models.User{
				EsportsID:       validateToken.ID,
				IsAccountFrozen: false,
			}

			if err := db.Shared().Where(user).First(&user).Error; err != nil {
				logger.Warning("get user error: ", err.Error())
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "User not found",
				})
				return
			}
			if !user.IsActive {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "User is not active",
				})
				return
			}
			user.SetRequest(ctx.Request)
			user.AuthToken = &header.Token
			ctx.Set("user", &user)
		case TokenTypeServer:
			break
		case TokenTypeUser:
			break
		default:
			logger.Warning(validateToken.Type, " token is not supported")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		ctx.Next()
	}
}
