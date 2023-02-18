package redis

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	r "github.com/go-redis/redis/v8"
)

const (
	gameDataBaseKey = "gameData_"
)

func SetFIFAShootupGameData(identfier string, gameData *response.FIFAShootupGameData) error {
	return Cache().Set(gameDataBaseKey+identfier, utils.JSON(gameData), r.KeepTTL)
}

func GetFIFAShootupGameData(identfier string) (*response.FIFAShootupGameData, error) {
	if cGameData, err := Cache().Get(gameDataBaseKey + identfier); err != nil {
		return nil, err
	} else {
		gameData := (*response.FIFAShootupGameData)(nil)

		if err := json.Unmarshal([]byte(cGameData), &gameData); err != nil {
			return gameData, err
		}
		return gameData, nil
	}
}
