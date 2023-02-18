package models

import (
	"time"
)

type MemberTable struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime             time.Time `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime             time.Time `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	UserID            int64     `gorm:"column:user_id" json:"user_id"`
	TableID           int64     `gorm:"column:mini_game_table_id" json:"mini_game_table_id"`
	IsEnabled         bool      `gorm:"column:is_enabled;default:true" json:"is_enabled"`
	IsAutoPlayEnabled bool      `gorm:"column:is_auto_play_enabled;default:false" json:"is_auto_play_enabled"`
	MaxBetAmount      float64   `gorm:"column:max_bet_amount;default:2000" json:"max_bet_amount"`
	MinBetAmount      float64   `gorm:"column:min_bet_amount;default:10" json:"min_bet_amount"`
	MaxPayoutAmount   float64   `gorm:"column:max_payout_amount;default:5000" json:"max_payout_amount"`
}

func (MemberTable) TableName() string {
	return "mini_game_memberminigametable"
}
