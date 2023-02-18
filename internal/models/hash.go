package models

import "time"

type Hash struct {
	ID        int64           `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime     time.Time       `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime     time.Time       `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	Seed      string          `gorm:"column:seed" json:"seed"`
	GameID    int64           `gorm:"column:game_id" json:"game_id"`
	TableID   int64           `gorm:"column:mini_game_table_id" json:"mini_game_table_id"`
	Status    string          `gorm:"column:status" json:"status"`
	Sequences *[]HashSequence `gorm:"foreignKey:HashID" json:"sequences"`
}

func (Hash) TableName() string {
	return "mini_game_minigamehash"
}

type HashSequence struct {
	ID       int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime    time.Time `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime    time.Time `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	Value    string    `gorm:"column:value" json:"value"`
	HashID   int64     `gorm:"column:mini_game_hash_id" json:"mini_game_hash_id"`
	Sequence int       `gorm:"column:sequence" json:"sequence"`
}

func (HashSequence) TableName() string {
	return "mini_game_minigamehashsequence"
}
