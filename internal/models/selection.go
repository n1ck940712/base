package models

import "time"

type SelectionHeader struct {
	ID             int64            `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime          time.Time        `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime          time.Time        `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	Name           string           `gorm:"column:name;size:128" json:"name"`
	Status         string           `gorm:"column:status;size:16" json:"status"`
	GameID         int64            `gorm:"column:game_id" json:"game_id"`
	SelectionLines *[]SelectionLine `gorm:"foreignKey:HeaderID;references:ID" json:"selections"`
}

func (SelectionHeader) TableName() string {
	return "mini_game_minigameselectionheader"
}

type SelectionLine struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime      time.Time `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime      time.Time `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	Name       string    `gorm:"column:name" json:"name"`
	Logo       string    `gorm:"column:logo" json:"logo"`
	Attributes string    `gorm:"column:attributes;type:jsonb" json:"attributes"`
	HeaderID   int64     `gorm:"column:mini_game_selection_header_id" json:"mini_game_selection_header_id"`
}

func (SelectionLine) TableName() string {
	return "mini_game_minigameselectionline"
}
