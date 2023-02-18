package models

import (
	"time"
)

type MemberConfig struct {
	ID     int64     `json:"id"`
	Ctime  time.Time `gorm:"autoCreateTime" json:"ctime"`
	Mtime  time.Time `gorm:"autoUpdateTime" json:"mtime"`
	Name   string    `json:"name"`
	Value  string    `json:"value"`
	GameId int64     `json:"game_id"`
	UserID int64     `json:"user_id"`
}

func (MemberConfig) TableName() string {
	return "mini_game_memberconfig"
}
