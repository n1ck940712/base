package main

import (
	"os"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/build/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/build/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/build/websocket"
)

func main() {
	switch getServerType() {
	case "api":
		api.Run(getServerIdentifier(), getPort())
	case "gameloop":
		gameloop.Run(getServerIdentifier())
	case "websocket":
		websocket.Run(getServerIdentifier(), getPort())
	default:
		panic("unsupported server type: \"" + getServerType() + "\"")
	}
}

func getServerType() string {
	return os.Getenv("SERVER_TYPE")
}

func getServerIdentifier() string {
	return os.Getenv("SERVER_IDENTIFIER")
}

func getPort() string {
	return os.Getenv("SERVER_PORT")
}
