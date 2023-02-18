package response

type LOLTowerResultData struct {
	Bomb *string `json:"bomb"`
	Gts  float64 `json:"gts,omitempty"`
}

func (LOLTowerResultData) Description() string {
	return "this is result loltower data yo!"
}

type LOLCoupleResultData struct {
	Result1 *string `json:"result1"`
	Result2 *string `json:"result2"`
	Result3 *string `json:"result3"`
	Gts     float64 `json:"gts,omitempty"`
}

func (LOLCoupleResultData) Description() string {
	return "this is result lolcouple data yo!"
}

type FIFAShootupResultData struct {
	ResultCard *FIFAShootupResultCardData `json:"result_card"`
	Result     *string                    `json:"result"`
	Gts        float64                    `json:"gts,omitempty"`
}

type FIFAShootupResultCardData struct {
	No   string `json:"no"`
	Type string `json:"type"`
}

func (FIFAShootupResultData) Description() string {
	return "this is result fifashootup data yo!"
}

type FishPrawnCrabResult struct {
	Result *string `json:"result"`
	Gts    float64 `json:"gts,omitempty"`
}

func (FishPrawnCrabResult) Description() string {
	return "this is result fishprawncrab"
}
