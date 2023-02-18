package placebet

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type WTCRequest struct {
	MemberID        int64      `json:"member_id,omitempty"`
	Amount          float64    `json:"amount,omitempty"`
	TransactionType string     `json:"transaction_type,omitempty"`
	RefNo           string     `json:"ref_no,omitempty"`
	AutoRollback    bool       `json:"auto_rollback,omitempty"`
	TicketDetail    WTRCTicket `json:"ticket_details,omitempty"`
}

//combo all odds euro
type WTRCTicket struct {
	ID                 string       `json:"id"`
	Odds               string       `json:"odds"`
	Amount             float64      `json:"amount,omitempty"`
	Result             *string      `json:"result,omitempty"`
	MapNum             *string      `json:"map_num"`
	Currency           string       `json:"currency"`
	Earnings           *float64     `json:"earnings,omitempty"`
	EventID            *string      `json:"event_id"`
	Handicap           *float64     `json:"handicap"`
	IsCombo            bool         `json:"is_combo,omitempty"`
	EuroOdds           string       `json:"euro_odds"`
	EventName          string       `json:"event_name"`
	MalayOdds          *string      `json:"malay_odds"`
	MemberCode         string       `json:"member_code,omitempty"`
	MemberOdds         string       `json:"member_odds"`
	TicketType         string       `json:"ticket_type"`
	DateCreated        time.Time    `json:"date_created"`
	GameTypeID         int16        `json:"game_type_id"`
	IsUnsettled        bool         `json:"is_unsettled"`
	BetSelection       string       `json:"bet_selection,omitempty"`
	BetTypeName        *string      `json:"bet_type_name"`
	MarketOption       *string      `json:"market_option"`
	ResultStatus       *string      `json:"result_status"`
	EventDatetime      time.Time    `json:"event_datetime"`
	GameTypeName       string       `json:"game_type_name"`
	RequestSource      *string      `json:"request_source"`
	CompetitionName    string       `json:"competition_name"`
	GameMarketName     string       `json:"game_market_name,omitempty"`
	MemberOddsStyle    string       `json:"member_odds_style"`
	ModifiedDateTime   *time.Time   `json:"modified_datetime,omitempty"`
	SettlementStatus   string       `json:"settlement_status"`
	SettlementDateTime *time.Time   `json:"settlement_datetime"`
	Tickets            []WTRCTicket `json:"tickets,omitempty"`
}

type WTResponse struct {
	ID              string    `json:"id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	RefNo           string    `json:"ref_no"`
	MemberID        int64     `json:"member_id"`
	PartnerID       int64     `json:"partner_id"`
	TicketID        string    `json:"ticket_id"`
	Ctime           time.Time `json:"ctime"`
	Mtime           time.Time `json:"mtime"`
	//Add TicketDetails
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func timeAppendPrefixZero(val int) string {
	temp := strconv.Itoa(val)
	if len(temp) == 1 {
		temp = "0" + temp
	}
	return temp
}

func getASCII(key int) string {
	a := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := fmt.Sprintf("%c", a[key])
	return b
}
