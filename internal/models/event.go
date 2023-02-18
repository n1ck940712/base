package models

import (
	"time"
)

type Event struct {
	ID                    *int64         `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	LocalDataVersion      int            `gorm:"column:local_data_version;default:1" json:"local_data_version"`
	ESID                  *int64         `gorm:"column:es_id" json:"es_id"`
	ESDataVersion         int32          `gorm:"column:es_data_version" json:"es_data_version"`
	SyncStatus            int16          `gorm:"column:sync_status" json:"sync_status"`
	Ctime                 time.Time      `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime                 time.Time      `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	Name                  string         `gorm:"column:name" json:"name"`
	StartDatetime         time.Time      `gorm:"column:start_datetime" json:"start_datetime"`
	Status                int16          `gorm:"column:status" json:"status"`
	MaxBet                float64        `gorm:"column:max_bet" json:"max_bet"`
	MaxPayout             float64        `gorm:"column:max_payout" json:"max_payout"`
	GameID                int64          `gorm:"column:game_id" json:"game_id"`
	HashSequenceID        int64          `gorm:"column:mini_game_hash_sequence_id" json:"mini_game_hash_sequence_id"`
	SelectionLineResultID *int64         `gorm:"column:mini_game_selection_line_result_id" json:"mini_game_selection_line_result_id"`
	SelectionHeaderID     int64          `gorm:"column:selection_header_id" json:"selection_header_id"`
	GroundType            string         `gorm:"column:ground_type" json:"ground_type"`
	IsAutoPlayExecuted    bool           `gorm:"column:is_auto_play_executed" json:"is_auto_play_executed"`
	PrevEventID           *int64         `gorm:"column:prev_event_id" json:"prev_event_id"`
	SettlementDate        *time.Time     `gorm:"column:settlement_date" json:"settlement_date"`
	TableID               int64          `gorm:"column:mini_game_table_id" json:"mini_game_table_id"`
	Results               *[]EventResult `gorm:"foreignKey:EventID;references:ID" json:"event_result"`
	Tickets               *[]Ticket      `gorm:"foreignKey:EventID;references:ID" json:"tickets"`
}

func (Event) TableName() string {
	return "mini_game_event"
}

type EventResult struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Ctime      time.Time `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime      time.Time `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	ResultType int16     `gorm:"column:result_type" json:"result_type"`
	Value      string    `gorm:"column:value" json:"value"`
	EventID    *int64    `gorm:"column:event_id" json:"event_id"`
	InstanceID *int16    `gorm:"column:instance_id" json:"instance_id"`
	Stage      *int16    `gorm:"column:stage" json:"stage"`
}

func (EventResult) TableName() string {
	return "mini_game_eventresult"
}

type CacheEvent struct {
	EventID                *int64 `json:"event_id"`
	Event                  *Event `json:"event"`
	IsBroadcastBetting     bool   `json:"is_broadcast_betting"`
	IsBroadcastStopBetting bool   `json:"is_broadcast_stop_betting"`
	IsBroadcastShowResult  bool   `json:"is_broadcast_show_result"`
	IsBroadcastShowBomb    bool   `json:"is_broadcast_bomb"`
}
