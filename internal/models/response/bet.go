package response

import (
	"time"
)

type BetData struct {
	UserID  string           `json:"user_id"` //encrypted user_id
	Tickets *[]BetTicketData `json:"tickets"`
}

type BetTicketData struct {
	ID            string                 `json:"id"`
	ComboTicketID string                 `json:"combo_ticket_id,omitempty"`
	Ctime         time.Time              `json:"ctime"`
	EventID       int64                  `json:"event_id"`
	GameID        int64                  `json:"game_id"`
	Selection     string                 `json:"selection,omitempty"`
	Amount        string                 `json:"amount"`
	MarketType    int16                  `json:"market_type"`
	ReferenceNo   string                 `json:"reference_no"`
	Result        string                 `json:"result,omitempty"`
	WinLossAmount string                 `json:"win_loss_amount,omitempty"`
	PayoutAmount  string                 `json:"payout_amount,omitempty"`
	Status        int16                  `json:"status"`
	TableID       int64                  `json:"table_id"`
	SelectionData BetTicketSelectionData `json:"selection_data"`
	Odds          string                 `json:"odds"`
	Level         *int8                  `json:"level,omitempty"`
	Skip          *int8                  `json:"skip,omitempty"`
	NextLevelOdds *string                `json:"next_level_odds,omitempty"`
}

type BetTicketSelectionData interface {
}

func (BetData) Description() string {
	return "fox jumps over"
}
