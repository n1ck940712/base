package gameloop

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	gameloop_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/fifashootup"
	gameloop_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/fishprawncrab"
	gameloop_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/lolcouple"
	gameloop_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
)

func Run(identifier string) {
	gameloop := newGameLoop(identifier)
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(identifier)+"gameloop", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	gameloop.Start()
}

func newGameLoop(identifier string) gameloop.Gamelooper {
	switch identifier {
	case constants_loltower.Identifier:
		// return gameloop_loltower.NewGameLoopOld()
		return gameloop_loltower.NewGameLoop()
	case constants_lolcouple.Identifier:
		return gameloop_lolcouple.NewGameLoop()
	case constants_fifashootup.Identifier:
		return gameloop_fifashootup.NewGameLoop()
	case constants_fishprawncrab.Identifier:
		return gameloop_fishprawncrab.NewGameLoop()
	default:
		panic("gameloop invalid identifier: \"" + identifier + "\"")
	}
}
