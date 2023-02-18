package request

import (
	"encoding/json"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type BetData struct {
	EventID types.JSONRaw `json:"event_id"`
	TableID types.JSONRaw `json:"table_id"`
	Tickets BetTickets    `json:"tickets"`
}

type BetTicketData struct {
	Selection         types.JSONRaw `json:"selection"`
	Amount            types.JSONRaw `json:"amount"`
	MarketType        types.JSONRaw `json:"market_type"`
	SelectionData     types.JSONRaw `json:"selection_data"`
	Odds              types.JSONRaw `json:"odds"`
	ReferenceNo       types.JSONRaw `json:"reference_no"`
	HongkongOdds      float64       //overriding hongkong odds - set only depending on requirement
	EuroOdds          float64       //overriding euro odds - set only depending on requirement
	MalayOdds         float64       //overriding malay odds - set only depending on requirement
	MaxPayoutEuroOdds float64       //used only for calculating max payout
}

type BetTickets []BetTicketData

func (bt *BetTickets) TotalAmount() float64 {
	total := 0.0

	for i := 0; i < len(*bt); i++ {
		total += (*bt)[i].Amount.ToFloat64()
	}
	return total
}

func (btd *BetTicketData) GetSelectionData() string {
	selectionDataStr := btd.SelectionData.String()

	if selectionDataStr == "" {
		return "{}"
	}
	return btd.SelectionData.RawString()
}

// json mirror of bet data
type JSONBetData struct {
	EventID json.RawMessage     `json:"event_id"`
	TableID json.RawMessage     `json:"table_id"`
	Tickets []JSONBetTicketData `json:"tickets"`
}

func (jbd *JSONBetData) GetBetData() *BetData {
	betData := BetData{}
	betData.EventID = types.JSONRaw(jbd.EventID)
	betData.TableID = types.JSONRaw(jbd.TableID)
	for _, jBetTicket := range jbd.Tickets {
		betData.Tickets = append(betData.Tickets, *jBetTicket.GetBetTicketData())
	}
	return &betData
}

type JSONBetTicketData struct {
	Selection     json.RawMessage `json:"selection"`
	Amount        json.RawMessage `json:"amount"`
	MarketType    json.RawMessage `json:"market_type"`
	SelectionData json.RawMessage `json:"selection_data"`
	Odds          json.RawMessage `json:"odds"`
	ReferenceNo   json.RawMessage `json:"reference_no"`
}

func (jbtd *JSONBetTicketData) GetBetTicketData() *BetTicketData {
	return &BetTicketData{
		Selection:     types.JSONRaw(jbtd.Selection),
		Amount:        types.JSONRaw(jbtd.Amount),
		MarketType:    types.JSONRaw(jbtd.MarketType),
		SelectionData: types.JSONRaw(jbtd.SelectionData),
		Odds:          types.JSONRaw(jbtd.Odds),
		ReferenceNo:   types.JSONRaw(jbtd.ReferenceNo),
	}
}

// request extension func
func (r *Request) GetBetData() *BetData {
	jsonBetData := JSONBetData{}

	json.Unmarshal(r.RawJSONData, &jsonBetData)
	return jsonBetData.GetBetData()
}
