package models

import (
	"encoding/json"
	"time"
)

type BetChipset struct {
	ID            string    `gorm:"column:id;primaryKey;autoIncrement:true"`
	Ctime         time.Time `gorm:"column:ctime;autoCreateTime"`
	Mtime         time.Time `gorm:"column:mtime;autoUpdateTime"`
	Currency      string    `gorm:"column:currency"`
	CurrencyRatio float64   `gorm:"column:currency_ratio"`
	Default       bool      `gorm:"column:default"`
	TableID       int64     `gorm:"column:mini_game_table_id"`
	Value         string    `gorm:"column:value"`
}

func (BetChipset) TableName() string {
	return "mini_game_betchipset"
}

func (bc *BetChipset) GetBetChips() *[]float64 {
	betChips := []float64{}

	json.Unmarshal([]byte(bc.Value), &betChips)
	return &betChips
}
