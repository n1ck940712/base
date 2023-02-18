package gameloop

import (
	"testing"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type testDatasource struct {
}

func (tds *testDatasource) GetIdentifier() string {
	return *types.String("testGameloop").Ptr()
}

func (tds *testDatasource) GetTableID() int64 {
	return 11
}

func TestGameLoop(t *testing.T) {
	db.UseLocalhost()
	gameloop := NewGameLoop(&testDatasource{})

	gameloop.SetPrepareCallback(func(elapsedTime Milliseconds) {
		logger.Info(gameloop.GetIdentifier(), " Prepare called elapsedTime: ", elapsedTime)
	})
	gameloop.CreatePhase(NewGamePhase(
		"open-betting",
		0,
		constants_loltower.StartBetMS,
		func(elapsedTime Milliseconds, phase GamePhase) {
			remainingTime := constants_loltower.StartBetMS - elapsedTime

			if event := gameloop.GetCurrentEvent(); event != nil {
				logger.Info(gameloop.GetIdentifier(), " elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " - OPEN-BETTING: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Info(gameloop.GetIdentifier(), " elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " - OPEN-BETTING: ", nil)
			}
		}))
	gameloop.CreatePhase(NewGamePhase(
		"close-betting",
		constants_loltower.StartBetMS,
		constants_loltower.StartBetMS+constants_loltower.StopBetMS,
		func(elapsedTime Milliseconds, phase GamePhase) {
			remainingTime := (constants_loltower.StartBetMS + constants_loltower.StopBetMS) - elapsedTime

			if event := gameloop.GetCurrentEvent(); event != nil {
				logger.Info(gameloop.GetIdentifier(), " elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " - CLOSE-BETTING: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Info(gameloop.GetIdentifier(), " elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " - CLOSE-BETTING: ", nil)
			}
		}))
	gameloop.CreatePhase(NewGamePhase(
		"on-resulting",
		constants_loltower.StartBetMS+constants_loltower.StopBetMS,
		constants_loltower.StartBetMS+constants_loltower.StopBetMS+constants_loltower.ShowResultMS,
		func(elapsedTime Milliseconds, phase GamePhase) {
			remainingTime := (constants_loltower.StartBetMS + constants_loltower.StopBetMS + constants_loltower.ShowResultMS) - elapsedTime

			if event := gameloop.GetCurrentEvent(); event != nil {
				logger.Info(gameloop.GetIdentifier(), " elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " - ON-RESULTING: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Info(gameloop.GetIdentifier(), " elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " - ON-RESULTING: ", nil)
			}
		}))
	gameloop.Start()
}
