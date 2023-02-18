package request

import (
	"errors"
	"strings"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/validate"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func (r *Request) getFIFAShootupValidSelections() []any {
	return []any{
		constants_fifashootup.Selection1,
		constants_fifashootup.Selection2,
		constants_fifashootup.Selection3,
	}
}

func (r *Request) getFIFAShootupValidMarketTypes() []any {
	return []any{
		constants_fifashootup.MarketType,
		*types.Int(constants_fifashootup.MarketType).String().Ptr(),
	}
}

func (r *Request) FIFAShootupValidateBet() error {
	validSelections := r.getFIFAShootupValidSelections()

	if types.Array[string](strings.Split(constants_fifashootup.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		validSelections = append(validSelections, []any{constants_fifashootup.SelectionWin, constants_fifashootup.SelectionLose}...)
	}
	return validate.Compose(r.Data).
		Type("event_id", validate.Number, validate.String).
		Value("table_id", constants_fifashootup.TableID, *types.Int(constants_fifashootup.TableID).String().Ptr()).
		Value("tickets", validate.ValueCallback(func(value any) error {
			if tickets, ok := value.([]any); ok && len(tickets) != 1 {
				return errors.New("(tickets) must contain one item only")
			}
			return nil
		})).
		Value("tickets.selection", validSelections...).
		Type("tickets.amount", validate.Number).
		Value("tickets.market_type", r.getFIFAShootupValidMarketTypes()...).
		Check()
}

func (r *Request) FIFAShootupValidateSelection() error {
	validSelections := append(r.getFIFAShootupValidSelections(), constants_fifashootup.SelectionPayout)

	if types.Array[string](strings.Split(constants_fifashootup.SelectionWinLoseEnabledEnv, ",")).Constains(settings.GetEnvironment().String()) {
		validSelections = append(validSelections, []any{constants_fifashootup.SelectionWin, constants_fifashootup.SelectionLose}...)
	}
	return validate.Compose(r.Data).
		Value("selection", validSelections...).
		Check()
}
