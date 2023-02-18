package main

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	gameloop_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
)

func main() {
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_fifashootup.Identifier)+"gameloop", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	gameloop_fifashootup.NewGameLoop().Start()
}
