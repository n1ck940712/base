package gamemanager_lolcouple

import (
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
)

func NewGameManager() gamemanager.GameManager {
	return gamemanager.NewGameManagerV2(NewDatasource(constants_lolcouple.Identifier))
}
