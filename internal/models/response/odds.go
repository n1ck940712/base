package response

type Oddsdata struct {
	OddsBonus   *float64 `json:"bonus,omitempty"`
	OddsBunus7  *float64 `json:"bonus7,omitempty"`
	OddsDouble  *float64 `json:"double,omitempty"`
	OddsLevel1  *float64 `json:"level1,omitempty"`
	OddsLevel2  *float64 `json:"level2,omitempty"`
	OddsLevel3  *float64 `json:"level3,omitempty"`
	OddsLevel4  *float64 `json:"level4,omitempty"`
	OddsLevel5  *float64 `json:"level5,omitempty"`
	OddsLevel6  *float64 `json:"level6,omitempty"`
	OddsLevel7  *float64 `json:"level7,omitempty"`
	OddsLevel8  *float64 `json:"level8,omitempty"`
	OddsLevel9  *float64 `json:"level9,omitempty"`
	OddsLevel10 *float64 `json:"level10,omitempty"`
	OddsSingle  *float64 `json:"single,omitempty"`
	OddsTriple  *float64 `json:"triple,omitempty"`
	Odds1       *float64 `json:"1,omitempty"`
	Odds2       *float64 `json:"2,omitempty"`
	Odds3       *float64 `json:"3,omitempty"`
	Odds3Under  *float64 `json:"3_under,omitempty"`
	Odds4       *float64 `json:"4,omitempty"`
	Odds4Under  *float64 `json:"4_under,omitempty"`
	Odds5       *float64 `json:"5,omitempty"`
	Odds5Under  *float64 `json:"5_under,omitempty"`
	Odds6       *float64 `json:"6,omitempty"`
	Odds6Under  *float64 `json:"6_under,omitempty"`
	Odds7       *float64 `json:"7,omitempty"`
	Odds777     *float64 `json:"777,omitempty"`
	Odds8       *float64 `json:"8,omitempty"`
	Odds8Over   *float64 `json:"8_over,omitempty"`
	Odds9       *float64 `json:"9,omitempty"`
	Odds9Over   *float64 `json:"9_over,omitempty"`
	Odds10      *float64 `json:"10,omitempty"`
	Odds10Over  *float64 `json:"10_over,omitempty"`
	Odds11      *float64 `json:"11,omitempty"`
	Odds11Over  *float64 `json:"11_over,omitempty"`
	Odds12      *float64 `json:"12,omitempty"`
}

func (Oddsdata) Description() string {
	return "the lazy dog"
}
