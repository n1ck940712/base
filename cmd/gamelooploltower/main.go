package main

import (
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	gameloop_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
)

func main() {
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_loltower.Identifier)+"gameloop", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	gameloop_loltower.NewGameLoop().Start()
}
