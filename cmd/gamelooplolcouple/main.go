package main

import (
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	gameloop_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
)

func main() {
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_lolcouple.Identifier)+"gameloop", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	gameloop_lolcouple.NewGameLoop().Start()
}
