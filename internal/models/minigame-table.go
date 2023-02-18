package models

import "time"

type GameTable struct {
	ID                                    int64         `gorm:"column:id;primaryKey;autoIncrement:true"`
	Ctime                                 time.Time     `gorm:"column:ctime;autoCreateTime"`
	Mtime                                 time.Time     `gorm:"column:mtime;autoUpdateTime"`
	Name                                  string        `gorm:"column:name"`
	GameID                                int64         `gorm:"column:game_id"`
	MinBetAmount                          float64       `gorm:"column:min_bet_amount"`
	MaxBetAmount                          float64       `gorm:"column:max_bet_amount"`
	MaxPayoutAmount                       float64       `gorm:"column:max_payout_amount"`
	IsEnabled                             bool          `gorm:"column:is_enabled"`
	DisableOnDailyMaxWinnings             *float64      `gorm:"column:disable_on_daily_max_winnings"`
	DisabledByMgTrigger                   bool          `gorm:"column:disabled_by_mg_trigger"`
	LastDailyMaxWinningsTriggeredDatetime *time.Time    `gorm:"column:last_daily_max_winnings_triggered_datetime"`
	BetChipsets                           *[]BetChipset `gorm:"foreignKey:TableID;references:ID"`
}

func (GameTable) TableName() string {
	return "mini_game_minigametable"
}
