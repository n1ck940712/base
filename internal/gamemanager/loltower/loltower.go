package gamemanager_loltower

import (
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
)

func NewGameManager() gamemanager.GameManager {
	return gamemanager.NewGameManagerV2(NewDatasource(constants_loltower.Identifier))
}
