package routefunc

import (
	"net/http"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/controller"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	eventService     service.EventService       = service.NewEvent()
	gameService      service.GameService        = service.NewGame()
	selectionService service.SelectionService   = service.NewSelection()
	eventController  controller.EventController = controller.NewEController(eventService, gameService, hashService, selectionService)
)

func CreateEvents(ctx *gin.Context) {
	eventController.GenerateEvents(ctx)
	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// func TestTask(ctx *gin.Context) {
// 	eventController.UpdateEventState(ctx)
// 	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
// }
