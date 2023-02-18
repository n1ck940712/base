package models

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID               string         `gorm:"column:id;primaryKey;" json:"id"`
	Ctime            time.Time      `gorm:"column:ctime;autoCreateTime" json:"ctime"`
	Mtime            time.Time      `gorm:"column:mtime;autoUpdateTime" json:"mtime"`
	MarketType       int16          `gorm:"column:market_type" json:"market_type"`
	Amount           float64        `gorm:"column:amount" json:"amount"`
	EuroOdds         float64        `gorm:"column:euro_odds" json:"euro_odds"`
	Selection        string         `gorm:"column:selection" json:"selection"`
	Status           int16          `gorm:"column:status" json:"status"`
	Result           *int16         `gorm:"column:result" json:"result,omitempty"`
	EventID          int64          `gorm:"column:event_id" json:"event_id"`
	UserID           int64          `gorm:"column:user_id" json:"user_id"`
	ReferenceNo      string         `gorm:"column:reference_no" json:"reference_no"`
	SelectionData    string         `gorm:"column:selection_data" json:"selection_data"`
	ESDataVersion    int32          `gorm:"column:es_data_version" json:"es_data_version"`
	LocalDataVersion int            `gorm:"column:local_data_version;default:1" json:"local_data_version"`
	StatusMtime      time.Time      `gorm:"column:status_mtime" json:"status_mtime"`
	SyncStatus       int16          `gorm:"column:sync_status" json:"sync_status"`
	Odds             float64        `gorm:"column:odds" json:"odds"`
	WinLossAmount    *float64       `gorm:"column:win_loss_amount" json:"win_loss_amount,omitempty"`
	IpAddress        string         `gorm:"column:ip_address" json:"ip_address"`
	RequestSource    *string        `gorm:"column:request_source" json:"request_source"`
	RequestUserAgent *string        `gorm:"column:request_user_agent" json:"request_user_agent"`
	PayoutAmount     *float64       `gorm:"column:payout_amount" json:"payout_amount,omitempty"`
	ExchangeRate     *float64       `gorm:"column:exchange_rate" json:"exchange_rate"`
	OriginalOdds     *float64       `gorm:"column:original_odds" json:"original_odds"`
	TableID          int64          `gorm:"column:mini_game_table_id" json:"mini_game_table_id"`
	ComboTickets     *[]ComboTicket `gorm:"foreignKey:TicketID;references:ID" json:"combo_tickets"`
}

func (Ticket) TableName() string {
	return "mini_game_ticket"
}

type ComboTicket struct {
	ID                     string         `gorm:"column:id;size:128;primaryKey" json:"id"`
	TicketID               string         `gorm:"column:ticket_id;size:128" json:"ticket_id"`
	Ctime                  time.Time      `gorm:"column:ctime;autoCreateTime;index:mini_game_combo_ticket_ctime_845ba22d;index:mini_game_cticket_user_ctime;index:mini_game_cticket_ticket_event_status_ctime_idx" json:"ctime"`
	Mtime                  time.Time      `gorm:"column:mtime;autoUpdateTime;index:mini_game_combo_ticket_mtime_5008f51b" json:"mtime"`
	MarketType             int16          `gorm:"column:market_type;index:mini_game_combo_ticket_market_type_38902c29;check:market_type >= 0" json:"market_type"`
	Amount                 float64        `gorm:"column:amount" json:"amount"`
	EuroOdds               float64        `gorm:"column:euro_odds" json:"euro_odds"`
	Selection              string         `gorm:"column:selection" json:"selection"`
	Status                 int16          `gorm:"column:status;index:mini_game_combo_ticket_status_c11c9533;index:mini_game_cticket_ticket_event_status_ctime_idx;index:mini_game_cticket_ticket_event_status_idx;check:status >= 0" json:"status"`
	Result                 *int16         `gorm:"column:result;check:result >= 0" json:"result,omitempty"`
	EventID                int64          `gorm:"column:event_id;index:mini_game_combo_ticket_event_id_9f854927;index:mini_game_cticket_ticket_event_status_ctime_idx;index:mini_game_cticket_ticket_event_status_idx" json:"event_id"`
	UserID                 int64          `gorm:"column:user_id;index:mini_game_cticket_user_id_7469c8a1;index:mini_game_cticket_user_ctime" json:"user_id"`
	ReferenceNo            string         `gorm:"column:reference_no" json:"reference_no"`
	SelectionData          string         `gorm:"column:selection_data;type:jsonb" json:"selection_data"`
	ESDataVersion          int32          `gorm:"column:es_data_version;check:es_data_version >= 0" json:"es_data_version"`
	ESID                   *string        `gorm:"column:es_id;size:128" json:"es_id"`
	LocalDataVersion       int8           `gorm:"column:local_data_version;check:local_data_version >= 0;default:1" json:"local_data_version"`
	StatusMtime            time.Time      `gorm:"column:status_mtime" json:"status_mtime"`
	SyncStatus             int8           `gorm:"column:sync_status;index:mini_game_combo_ticket_sync_status_e76debc5;check:sync_status >= 0" json:"sync_status"`
	AutoPlayID             *uuid.UUID     `gorm:"column:auto_play_id;type:uuid;index:mini_game_combo_ticket_auto_play_id_b15fc53e" json:"auto_play_id"`
	Odds                   float64        `gorm:"column:odds" json:"odds"`
	WinLossAmount          *float64       `gorm:"column:win_loss_amount" json:"win_loss_amount,omitempty"`
	IpAddress              string         `gorm:"column:ip_address" json:"ip_address"`
	Country                *string        `gorm:"column:country" json:"country"`
	Payload                *string        `gorm:"column:payload" json:"payload"`
	RequestSource          *string        `gorm:"column:request_source" json:"request_source"`
	RequestUserAgent       *string        `gorm:"column:request_user_agent" json:"request_user_agent"`
	PayoutAmount           *float64       `gorm:"column:payout_amount" json:"payout_amount,omitempty"`
	ExchangeRate           *float64       `gorm:"column:exchange_rate" json:"exchange_rate"`
	OriginalOdds           *float64       `gorm:"column:original_odds" json:"original_odds"`
	PossibleWinningsAmount *float64       `gorm:"column:possible_winnings_amount" json:"possible_winnings_amount"`
	TableID                int64          `gorm:"column:mini_game_table_id;index:mini_game_combo_ticket_mini_game_table_id_8ec279af" json:"mini_game_table_id"`
	Results                *[]EventResult `gorm:"foreignKey:EventID;references:EventID" json:"event_result"`
}

func (ComboTicket) TableName() string {
	return "mini_game_combo_ticket"
}

type TicketsForSettlement struct {
	TicketID            string    `json:"ticket_id"`
	LatestComboTicketID string    `json:"latest_combo_ticket_id"`
	ComboCtime          time.Time `json:"combo_ctime"`
	Count               int64     `json:"count"`
	EventID             int64     `json:"event_id"`
}

type OldTicketsForSettlement struct {
	ID           string         `json:"id"`
	ComboTickets *[]ComboTicket `gorm:"foreignKey:TicketID;references:ID" json:"combo_tickets"`
}
