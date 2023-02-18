package models

import (
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type JInRequest struct {
	Type types.Bytes `json:"type"`
	Data JInData     `json:"data,omitempty"`
}

type JInData struct {
	EventID   types.Bytes `json:"event_id,omitempty"`
	TableID   types.Bytes `json:"table_id,omitempty"`
	Tickets   []JInTicket `json:"tickets,omitempty"`
	Selection types.Bytes `json:"selection,omitempty"`
}

type JInTicket struct {
	Selection   types.Bytes `json:"selection"`
	Amount      types.Bytes `json:"amount"`
	MarketType  types.Bytes `json:"market_type"`
	ReferenceNo types.Bytes `json:"reference_no"`
}
