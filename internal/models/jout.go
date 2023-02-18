package models

import (
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
)

type JsonResponse struct {
	Type        string        `json:"type,omitempty"`
	Data        *ResponseData `json:"data,omitempty"`
	Cts         float64       `json:"cts,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
	IsAnonymous *bool         `json:"is_anonymous,omitempty"`
	MemberCode  string        `json:"member_code,omitempty"`
	QueryParams *QueryParams  `json:"query_params,omitempty"`
}

type ResponseData struct {
	Tickets []TicketResponse    `json:"tickets,omitempty"`
	UserID  string              `json:"user_id,omitempty"`
	Level   string              `json:"level,omitempty"`
	Users   []ChampionResponse  `json:"users,omitempty"`
	Ied     errors.GenericError `json:"ied,omitempty"`
	Ieid    errors.IEIDError    `json:"ieid,omitempty"`
	Mid     string              `json:"mid,omitempty"`
	Message string              `json:"msg,omitempty"`
	Type    string              `json:"type,omitempty"`
}

type TicketResponse struct {
	ID            string    `json:"id"`
	ComboTicketID string    `json:"combo_ticket_id,omitempty"`
	Ctime         time.Time `json:"ctime,omitempty"`
	EventID       int64     `json:"event_id"`
	GameID        int16     `json:"game_id"`
	Selection     string    `json:"selection,omitempty"`
	Amount        string    `json:"amount,omitempty"`
	MarketType    int16     `json:"market_type,omitempty"`
	ReferenceNo   string    `json:"reference_no,omitempty"`
	Result        string    `json:"result,omitempty"`
	WinLossAmount string    `json:"win_loss_amount,omitempty"`
	Status        int16     `json:"status,omitempty"`
	TableID       int16     `json:"table_id,omitempty"`
	SelectionData string    `json:"selection_data,omitempty"`
	Odds          float32   `json:"odds,omitempty"`
}

type ChampionResponse struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type QueryParams struct {
	AuthToken []string `json:"auth_token,omitempty"`
}
