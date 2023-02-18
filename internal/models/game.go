package models

import (
	"time"
)

type Game struct {
	ID               int64        `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime            time.Time    `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime            time.Time    `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	LocalDataVersion int32        `gorm:"column:local_data_version;default:1" json:"local_data_version"`
	ESID             int64        `gorm:"column:es_id" json:"es_id"`
	ESDataVersion    int32        `gorm:"column:es_data_version" json:"es_data_version"`
	SyncStatus       int16        `gorm:"column:sync_status" json:"sync_status"`
	Name             string       `gorm:"column:name;size:255" json:"name"`
	ShortName        string       `gorm:"column:short_name;size:64" json:"short_name"`
	Type             string       `gorm:"column:type;size:25" json:"type"`
	GameTables       *[]GameTable `gorm:"foreignKey:GameID;references:ID" json:"game_tables"`
}

func (Game) TableName() string {
	return "mini_game_game"
}
