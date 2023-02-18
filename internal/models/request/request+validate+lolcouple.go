package request

import (
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/validate"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func (r *Request) getLOLCoupleValidSelections() []any {
	return []any{
		constants_lolcouple.Selection1,
		constants_lolcouple.Selection2,
		constants_lolcouple.Selection3,
		constants_lolcouple.Selection4,
		constants_lolcouple.Selection5,
		*types.String(constants_lolcouple.Selection5).Int().Ptr(),
		constants_lolcouple.Selection6,
		constants_lolcouple.Selection7,
		constants_lolcouple.Selection8,
		constants_lolcouple.Selection9,
		constants_lolcouple.Selection10,
		*types.String(constants_lolcouple.Selection10).Int().Ptr(),
	}
}

func (r *Request) getLOLCoupleValidMarketTypes() []any {
	return []any{
		constants_lolcouple.MarketType,
		*types.Int(constants_lolcouple.MarketType).String().Ptr(),
	}
}

func (r *Request) LOLCoupleValidateBet() error {
	validSelections := r.getLOLCoupleValidSelections()

	//TODO: added selection win lose if needed

	return validate.Compose(r.Data).
		Type("event_id", validate.Number, validate.String).
		Value("table_id", constants_lolcouple.TableID, *types.Int(constants_lolcouple.TableID).String().Ptr()).
		Value("tickets.selection", validSelections...).
		Type("tickets.amount", validate.Number).
		Value("tickets.market_type", r.getLOLCoupleValidMarketTypes()...).
		Check()
}
