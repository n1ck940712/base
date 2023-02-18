package scheduler

import (
	"time"

	gamemanager_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
)

func Init() {
	_ = godotenv.Load(".env")
	s := gocron.NewScheduler(time.UTC)

	// run generate hash if environment = local
	gmLOLTower := gamemanager_loltower.NewGameManager()
	s.Every(14).Seconds().Do(func() {
		if err := gmLOLTower.CreateFutureHashes(); err != nil {
			logger.Error("gamemanager_loltower CreateFutureHashes error: ", err.Error())
		}
		if err := gmLOLTower.CreateFutureEvents(); err != nil {
			logger.Error("gamemanager_loltower CreateFutureEvents error: ", err.Error())
		}
	})
	s.StartAsync()
}
