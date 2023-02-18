package request

import (
	"strings"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/validate"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func (r *Request) getLOLTowerValidSelections() []any {
	return []any{
		*constants_loltower.Selection1.Ptr(),
		*constants_loltower.Selection2.Ptr(),
		*constants_loltower.Selection3.Ptr(),
		*constants_loltower.Selection4.Ptr(),
		*constants_loltower.Selection5.Ptr(),
		*constants_loltower.Selection1.String().Ptr(),
		*constants_loltower.Selection2.String().Ptr(),
		*constants_loltower.Selection3.String().Ptr(),
		*constants_loltower.Selection4.String().Ptr(),
		*constants_loltower.Selection5.String().Ptr(),
	}
}

func (r *Request) getLOLTowerValidMarketTypes() []any {
	return []any{
		constants_loltower.MarketType,
		*types.Int(constants_loltower.MarketType).String().Ptr(),
	}
}

func (r *Request) LOLTowerValidateBet() error {
	validSelections := r.getLOLTowerValidSelections()

	if types.Array[string](strings.Split(constants_loltower.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		validSelections = append(validSelections, []any{constants_loltower.SelectionWin, constants_loltower.SelectionLose}...)
	}
	return validate.Compose(r.Data).
		Type("event_id", validate.Number, validate.String).
		Value("table_id", constants_loltower.TableID, *types.Int(constants_loltower.TableID).String().Ptr()).
		Value("tickets.selection", validSelections...).
		// Value("tickets.amount", validate.ValueCallback(func(value any) error {
		// 	if fValue, ok := value.(float64); ok {
		// 		if !(fValue >= 10) {
		// 			return errors.New("must not be greater than or equal to 10")
		// 		}
		// 	}
		// 	return nil
		// })).
		Type("tickets.amount", validate.Number).
		Value("tickets.market_type", r.getLOLTowerValidMarketTypes()...).
		Check()
}

func (r *Request) LOLTowerValidateSelection() error {
	validSelections := append(r.getLOLTowerValidSelections(), []any{constants_loltower.SelectionPayout, constants_loltower.SelectionSkip}...)

	if types.Array[string](strings.Split(constants_loltower.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		validSelections = append(validSelections, []any{constants_loltower.SelectionWin, constants_loltower.SelectionLose}...)
	}
	return validate.Compose(r.Data).
		Value("selection", validSelections...).
		Check()
}
