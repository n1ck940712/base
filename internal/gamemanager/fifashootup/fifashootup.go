package gamemanager_fifashootup

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
)

func NewGameManager() gamemanager.GameManager {
	return gamemanager.NewGameManagerV2(NewDatasource(constants_fifashootup.Identifier))
}
