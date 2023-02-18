package migrations

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
)

type MiniGameFakeMember = models.FakeMember

func (m *Migrations) MigrateMiniGameFakeMember() {
	m.handleTableCreation("mini_game_fake_member", &MiniGameFakeMember{})
}
