package gamemanager_fishprawncrab

import (
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
)

func NewGameManager() gamemanager.GameManager {
	return gamemanager.NewGameManagerV2(NewDatasource(constants_fishprawncrab.Identifier))
}
