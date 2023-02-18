package gameloop_lolcouple

import (
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
)

func TestGameLoop(t *testing.T) {
	db.UseLocalhost()
	redis.UseLocalhost()
	gameloop := NewGameLoop()

	gameloop.Start()
}
