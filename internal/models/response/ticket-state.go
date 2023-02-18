package response

type LOLTowerTicketState struct {
	BetAmount      *float64 `json:"bet_amount"`
	Level          int      `json:"level"`
	Skip           int      `json:"skip"`
	Selection      *string  `json:"selection"`
	MaxPayoutLevel *int8    `json:"max_payout_level"`
}

func (LOLTowerTicketState) Description() string {
	return "response loltower ticket state only for json responses"
}

type TicketState struct {
	Tickets ResponseData `json:"tickets"`
}

func (TicketState) Description() string {
	return "response ticket state only for json responses"
}

type LolCoupleTicket struct {
	MarketType    int16    `json:"market_type"`
	Amount        float64  `json:"amount"`
	Selection     string   `json:"selection"`
	WinLossAmount *float64 `json:"win_loss_amount"`
}

type LolCoupleTickets []LolCoupleTicket

func (LolCoupleTickets) Description() string {
	return "response lolcouple ticket only for json responses"
}

type FifaShootupTicket struct {
	Amount       float64  `json:"amount"`
	PayoutAmount *float64 `json:"payout_amount"`
	Ball1        *string  `json:"ball1"`
	Ball2        *string  `json:"ball2"`
	Ball3        *string  `json:"ball3"`
	Ball4        *string  `json:"ball4"`
	Ball5        *string  `json:"ball5"`
	Selection    *string  `json:"selection"`
}

type FifaShootupTickets []FifaShootupTicket

func (FifaShootupTickets) Description() string {
	return "response  ticket FifaShootup only for json responses"
}

type FishPrawnCrabTicket struct {
	MarketType    int16    `json:"market_type"`
	Amount        float64  `json:"amount"`
	Selection     string   `json:"selection"`
	WinLossAmount *float64 `json:"win_loss_amount"`
}

type FishPrawnCrabTickets []FishPrawnCrabTicket

func (FishPrawnCrabTickets) Description() string {
	return "response FishPrawnCrab ticket only for json responses"
}
