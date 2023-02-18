package routefunc

import (
	"net/http"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/controller"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	hashService    service.HashService       = service.HashNew()
	hashController controller.HashController = controller.NewHashController(hashService, gameService, eventService)
)

func GetHash(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, hashController.HashFindAll(ctx))
}

func GenerateHash(ctx *gin.Context) {
	res, err := hashController.HashSave(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		ctx.JSON(http.StatusCreated, res)
	}
}
