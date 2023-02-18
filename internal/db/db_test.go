package db

import (
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
)

func TestDB(t *testing.T) {
	UseLocalhost()
	logger.Info("db: ", utils.PrettyJSON(Shared()))
}
