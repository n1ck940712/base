package response

type FIFAShootupGameData struct {
	EventID    int64            `json:"event_id"`
	LeftCard   *FIFAShootupCard `json:"left_card"`
	RightCard  *FIFAShootupCard `json:"right_card"`
	MiddleCard *FIFAShootupCard `json:"middle_card"`
}

type FIFAShootupCard struct {
	Card  *FIFAShootupCardData `json:"card"`
	Odds  float64              `json:"odds"`
	Range []string             `json:"range"`
}

type FIFAShootupCardData struct {
	No   string `json:"no"`
	Type string `json:"type"`
}

func (FIFAShootupGameData) Description() string {
	return "response fifashootup game data only for json responses"
}
