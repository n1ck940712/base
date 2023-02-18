package process_game_data

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_game_data "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-game-data"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
)

type GameDataProcess interface {
	process_game_data.GameDataProcess
}

type gameDataProcess struct {
	datasource process_game_data.GameDataDatasource
}

func NewGameDataProcess(datasource process_game_data.GameDataDatasource) GameDataProcess {
	return &gameDataProcess{datasource: datasource}
}

func (gdp *gameDataProcess) GetGameData() response.ResponseData {
	if gameData, err := redis.GetFIFAShootupGameData(gdp.datasource.GetIdentifier()); err != nil {
		logger.Info(gdp.datasource.GetIdentifier(), " GetGameData error: ", err.Error())
		return nil
	} else {
		return gameData
	}
}
