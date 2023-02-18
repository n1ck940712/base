package main

import (
	"encoding/json"
	"net/http"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/controller"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/middlewares"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/routefunc"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsconsumer"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	go slack.SendPayload(slack.NewLootboxNotification("api", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	server := gin.Default()
	server.Use(
		gin.Recovery(),
		cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"DELETE", "GET", "OPTIONS", "PATCH", "POST", "PUT"},
			AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With", "user-agent"},
			AllowCredentials: true,
			MaxAge:           24 * time.Hour,
		}),
	)

	// public endpoints
	pub := server.Group("/api/game-client/mini-game-go")
	{
		pv1 := pub.Group("v1")
		{
			pv1.GET("/events/generate/:game_id", routefunc.CreateEvents)
			pv1.GET("/hash", routefunc.GetHash)
			pv1.GET("/hash/generate/:game_id", routefunc.GenerateHash)
			pv1.POST("/test", func(ctx *gin.Context) {
				var message map[string]interface{}
				ctx.BindJSON(&message)

				payload, err := json.Marshal(message)
				if err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"message": "unexpected JSON format"})
				}

				wsconsumer.NewMessagePublishBroker().Publish(string(payload))
			})
		}
	}

	apiRoutes := server.Group("/api/game-client/mini-game-go", middlewares.Authenticate(middlewares.TokenTypeMember))
	{
		v1 := apiRoutes.Group("v1")
		{
			v1.GET("/tables/:table_id/state/", controller.NewProcessController().Process(process.StateType))
			v1.GET("/tables/:table_id/config/", controller.NewProcessController().Process(process.ConfigType))
			v1.GET("/tables/:table_id/odds/", controller.NewProcessController().Process(process.OddsType))
			v1.GET("/tables/:table_id/ticket-state/", controller.NewProcessController().Process(process.TicketType))
			v1.POST("/tables/:table_id/tickets/", controller.NewProcessController().Process(process.BetType))
			v1.POST("/tables/:table_id/selection/", controller.NewProcessController().Process(process.SelectionType))
			v1.GET("/tables/:table_id/member-list/", controller.NewProcessController().Process(process.MemberListType))
			v1.GET("/tables/:table_id/game-data/", controller.NewProcessController().Process(process.GameDataType))
			v1.PATCH("/tables/:table_id/config/", routefunc.UpdateConfig)
		}
	}

	apiAdminRoutes := server.Group("/api/admin/mini-game-go", middlewares.Authenticate(middlewares.TokenTypeServer))
	{
		adminV1 := apiAdminRoutes.Group("v1")
		{
			adminV1.GET("/tables/:table_id/fake-member/:fake_id", routefunc.GetTableFakeMember)
			adminV1.GET("/tables/:table_id/fake-member/", routefunc.GetTableFakeMembers)
			adminV1.POST("/tables/:table_id/fake-member/", routefunc.CreateTableFakeMember)
			adminV1.PATCH("/tables/:table_id/fake-member/:fake_id", routefunc.UpdateTableFakeMember)
			adminV1.DELETE("/tables/:table_id/fake-member/:fake_id", routefunc.DeleteTableFakeMember)
		}
	}

	server.Run(":" + settings.PORT)
}
