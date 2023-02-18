package bind

type ValidateToken struct {
	ID       int64    `json:"id"`   //esports_id when type is member
	Type     string   `json:"type"` //user or server or member
	Metadata metadata `json:"metadata"`
}

type metadata struct {
	IsSuperUser     bool     `json:"is_superuser"`
	Group           []string `json:"group"`
	MemberCode      string   `json:"member_code"`
	PartnerID       int32    `json:"partner_id"`
	IsAccountFrozen bool     `json:"is_account_frozen"`
	CurrencyCode    string   `json:"currency_code"`
	ExchangeRate    float32  `json:"exchange_rate"`
	CurrencyRation  float32  `json:"currency_ration"`
}
