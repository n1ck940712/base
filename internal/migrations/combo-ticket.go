package migrations

import "bitbucket.org/esportsph/minigame-backend-golang/internal/models"

type MiniGameComboTicket = models.ComboTicket

func (m *Migrations) migrateComboTicket() {
	m.handleTableCreation("mini_game_combo_ticket", &MiniGameComboTicket{})

	m.indexes = []string{
		"mini_game_combo_ticket_auto_play_id_b15fc53e",
		"mini_game_combo_ticket_ctime_845ba22d",
		"mini_game_combo_ticket_event_id_9f854927",
		"mini_game_combo_ticket_market_type_38902c29",
		"mini_game_combo_ticket_mini_game_table_id_8ec279af",
		"mini_game_combo_ticket_mtime_5008f51b",
		"mini_game_combo_ticket_status_c11c9533",
		"mini_game_combo_ticket_sync_status_e76debc5",

		"mini_game_cticket_id_c13196f4",
		"mini_game_cticket_user_id_7469c8a1",
		"mini_game_cticket_user_ctime",
		"mini_game_cticket_ticket_event_status_ctime_idx",
		"mini_game_cticket_ticket_event_status_idx",
	}
	m.handleIndexCreation(&MiniGameComboTicket{})
}
