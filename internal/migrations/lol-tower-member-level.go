package migrations

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
)

type MiniGameLolTowerMemberLevel = models.LolTowerMemberLevel

func (m *Migrations) migrateLolTowerMemberLevel() {
	m.handleTableCreation("mini_game_lol_tower_member_level", &MiniGameLolTowerMemberLevel{})

	m.indexes = []string{
		"mini_game_lol_tower_member_level_idx_ctime_ticket_id_combo_id",
		"mini_game_lol_tower_member_level_idx_user_id",
		"mini_game_lol_tower_member_level_idx_combo_ticket_id",
	}
	m.handleIndexCreation(&MiniGameLolTowerMemberLevel{})
}
