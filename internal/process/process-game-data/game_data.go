package process_game_data

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
)

const GameDataType = "game_data"

type GameDataDatasource interface {
	GetIdentifier() string
}

type GameDataProcess interface {
	GetGameData() response.ResponseData
}
