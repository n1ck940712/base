package request

import "bitbucket.org/esportsph/minigame-backend-golang/internal/validate"

func (r *Request) ValidateBet() error {
	return validate.Compose(r.Data).
		Type("event_id", validate.Number, validate.String).
		Type("table_id", validate.Number, validate.String).
		Type("tickets.amount", validate.Number).
		Type("tickets.market_type", validate.Number, validate.String).
		Check()
}
