package models

import "time"

type LolTowerMemberLevel struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime         time.Time `gorm:"column:ctime;autoCreateTime;index:mini_game_lol_tower_member_level_idx_ctime_ticket_id_combo_id;index:mini_game_lol_tower_member_level_idx_ctime,sort:desc" json:"ctime"`
	Mtime         time.Time `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	UserID        int64     `gorm:"column:user_id;index:mini_game_lol_tower_member_level_idx_user_id" json:"user_id"`
	TicketID      string    `gorm:"column:ticket_id;index:mini_game_lol_tower_member_level_idx_ticket_id;index:mini_game_lol_tower_member_level_idx_ctime_ticket_id_combo_id" json:"ticket_id"`
	ComboTicketID string    `gorm:"column:combo_ticket_id;index:mini_game_lol_tower_member_level_idx_combo_ticket_id;index:mini_game_lol_tower_member_level_idx_ctime_ticket_id_combo_id" json:"combo_ticket_id"`
	Level         int8      `gorm:"column:level" json:"level"`
	Skip          int8      `gorm:"column:skip" json:"skip"`
	NextLevelOdds float64   `gorm:"column:next_level_odds;type:numeric(10,2)" json:"next_level_odds,omitempty"`
}

func (LolTowerMemberLevel) TableName() string {
	return "mini_game_lol_tower_member_level"
}
