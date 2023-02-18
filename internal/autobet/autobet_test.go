package autobet

import (
	"testing"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type loltowerDatasource struct {
}

func (tds *loltowerDatasource) GetIdentifier() string {
	return *types.String("test-loltower-bets").Ptr()
}

func (tds *loltowerDatasource) GetTableID() int64 {
	return constants_loltower.TableID
}

func TestLOLTowerBets(t *testing.T) {
	db.UseLocalhost()
	gamelooploltower := gameloop.NewGameLoop(&loltowerDatasource{})

	gamelooploltower.SetPrepareCallback(func(elapsedTime gameloop.Milliseconds) {
		println(gamelooploltower.GetIdentifier(), " SetPrepareCallback called!")
	})
	gamelooploltower.CreatePhase(gameloop.NewGamePhase(
		"BETTING",
		0,
		constants_loltower.StartBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			remainingTime := constants_loltower.StartBetMS - elapsedTime

			if event := gamelooploltower.GetCurrentEvent(); event != nil {
				logger.Info(gamelooploltower.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Info(gamelooploltower.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " event: ", nil)
			}
			phase.SetEndMSOffset(gamelooploltower.ElapsedTimeOffset(gamelooploltower.ElapsedTime()))
		}))
	gamelooploltower.CreatePhase(gameloop.NewGamePhase(
		"STOP_BETTING",
		constants_loltower.StartBetMS,
		constants_loltower.StartBetMS+constants_loltower.StopBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			remainingTime := (constants_loltower.StartBetMS + constants_loltower.StopBetMS) - elapsedTime

			if event := gamelooploltower.GetCurrentEvent(); event != nil {
				logger.Info(gamelooploltower.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Info(gamelooploltower.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " event: ", nil)
			}
			phase.SetEndMSOffset(gamelooploltower.ElapsedTimeOffset(gamelooploltower.ElapsedTime()))
		}))
	gamelooploltower.CreatePhase(gameloop.NewGamePhase(
		"SHOW_RESULT",
		constants_loltower.StartBetMS+constants_loltower.StopBetMS,
		constants_loltower.StartBetMS+constants_loltower.StopBetMS+constants_loltower.ShowResultMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			remainingTime := (constants_loltower.StartBetMS + constants_loltower.StopBetMS + constants_loltower.ShowResultMS) - elapsedTime

			if event := gamelooploltower.GetCurrentEvent(); event != nil {
				logger.Info(gamelooploltower.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Info(gamelooploltower.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, " remainingTime: ", remainingTime, " event: ", nil)
			}
			phase.SetEndMSOffset(gamelooploltower.ElapsedTimeOffset(gamelooploltower.ElapsedTime()))
		}))
	gamelooploltower.Start()
}
