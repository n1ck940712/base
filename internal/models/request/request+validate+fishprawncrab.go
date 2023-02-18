package request

import (
	"strings"

	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/validate"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func (r *Request) getFishPrawnCrabValidSelections() []any {
	return []any{
		constants_fishprawncrab.Selection1,
		constants_fishprawncrab.Selection2,
		constants_fishprawncrab.Selection3,
		constants_fishprawncrab.Selection4,
		constants_fishprawncrab.Selection5,
		constants_fishprawncrab.Selection6,
	}
}

func (r *Request) getFishPrawnCrabValidMarketTypes() []any {
	return []any{
		constants_fishprawncrab.MarketTypeSingle,
		constants_fishprawncrab.MarketTypeDouble,
		constants_fishprawncrab.MarketTypeTriple,
		*types.Int(constants_fishprawncrab.MarketTypeSingle).String().Ptr(),
		*types.Int(constants_fishprawncrab.MarketTypeDouble).String().Ptr(),
		*types.Int(constants_fishprawncrab.MarketTypeTriple).String().Ptr(),
	}
}

func (r *Request) FishPrawnCrabValidateBet() error {
	validSelections := r.getFishPrawnCrabValidSelections()

	if types.Array[string](strings.Split(constants_fishprawncrab.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		validSelections = append(validSelections, []any{constants_fishprawncrab.SelectionWin, constants_fishprawncrab.SelectionLose}...)
	}
	return validate.Compose(r.Data).
		Type("event_id", validate.Number, validate.String).
		Value("table_id", constants_fishprawncrab.TableID, *types.Int(constants_fishprawncrab.TableID).String().Ptr()).
		Value("tickets.selection", validSelections...).
		Type("tickets.amount", validate.Number).
		Value("tickets.market_type", r.getFishPrawnCrabValidMarketTypes()...).
		Check()
}
