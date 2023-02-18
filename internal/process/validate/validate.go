package validate

import (
	"fmt"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
)

func (user *User) ValidateBalance(betAmount float64) *response.ErrorData {
	if user.Data == nil {
		return response.ErrorWithMessage("user is nil", user.ProcessType)
	}
	balanceRes := map[string]any{}

	if err := api.
		NewAPI(settings.GetEBOAPI().String() + "/v4/wallet/balance").
		SetIdentifier("ValidateBalance").
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		AddQueries(map[string]string{"member_id": fmt.Sprint(user.Data.EsportsID)}).
		Get(&balanceRes); err != nil {
		return response.ErrorIE(errors.ESPORTS_API_FAILED, errors.IEID_ESPORTS_API_FAILED, user.ProcessType)
	}

	if balance, ok := balanceRes["balance"].(float64); ok {
		if betAmount > balance {
			return response.ErrorIE(errors.NOT_ENOUGH_BALANCE_EXCEPTION, errors.IEID_NOT_ENOUGH_BALANCE, user.ProcessType)
		}
	}
	return nil
}

func (user *User) ValidateStatus() *response.ErrorData {
	if user.Data == nil {
		return response.ErrorWithMessage("user is nil", user.ProcessType)
	}
	if user.Data.IsAccountFrozen {
		return response.ErrorGE(errors.ACCOUNT_IS_FROZEN, errors.IEID_ACCOUNT_IS_FROZEN, user.ProcessType)
	}
	if user.Data.SleepStatus > 0 {
		return response.ErrorGE(errors.USER_IS_IN_SLEEP_STATUS, errors.IEID_ACCOUNT_IS_FROZEN, user.ProcessType)
	}

	return nil
}

func (event *Event) ValidateTicket(ticket Ticket) *response.ErrorData {
	if event.Data == nil {
		return response.ErrorWithMessage("event is nil", event.ProcessType)
	}
	if ticket.Data == nil {
		return nil
	}
	if *event.Data.ID == ticket.Data.EventID {
		return response.ErrorIE(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, event.ProcessType)
	}
	return nil
}

func (event *Event) ValidateComboTicket(comboTicket ComboTicket) *response.ErrorData {
	if event.Data == nil {
		return response.ErrorWithMessage("event is nil", event.ProcessType)
	}
	if comboTicket.Data == nil {
		return nil
	}
	if *event.Data.ID == comboTicket.Data.EventID {
		return response.ErrorIE(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, event.ProcessType)
	}
	return nil
}

func (event *Event) ValidateBetEvent(eventID int64) *response.ErrorData {
	if event.Data == nil {
		return response.ErrorWithMessage("event is nil", event.ProcessType)
	}
	if *event.Data.ID != eventID {
		return response.ErrorIE(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_NOT_FOUND, event.ProcessType)
	}
	return nil
}

func (event *Event) ValidateIsOpen(startMS int64, endMS int64) *response.ErrorData {
	if event.Data == nil {
		return response.ErrorWithMessage("event is nil", event.ProcessType)
	}
	elapsedTime := time.Now().UnixMilli() - time.Time(event.Data.StartDatetime).UnixMilli()

	if startMS <= elapsedTime && elapsedTime < endMS && event.Data.Status == constants.EVENT_STATUS_ACTIVE {
		return nil
	}
	return response.ErrorIE(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_NOT_ON_BETTING_PHASE, event.ProcessType)
}

func (gameTable *GameTable) ValidateEnabled() *response.ErrorData {
	if gameTable.Data == nil {
		return response.ErrorWithMessage("game is nil", gameTable.ProcessType)
	}
	if !gameTable.Data.IsEnabled || gameTable.Data.DisabledByMgTrigger {
		return response.ErrorIE(errors.MINI_GAME_DISABLED, errors.IEID_GAME_DISABLED, gameTable.ProcessType)
	}
	return nil
}

func (memberTable *MemberTable) ValidateEnabled() *response.ErrorData {
	if memberTable.Data == nil {
		return response.ErrorWithMessage("member is nil", memberTable.ProcessType)
	}
	if !memberTable.Data.IsEnabled {
		return response.ErrorIE(errors.MINI_GAME_DISABLED, errors.IEID_GAME_DISABLED, memberTable.ProcessType)
	}
	return nil
}

func (memberTable *MemberTable) ValidateMaxBet(betAmount float64) *response.ErrorData {
	if memberTable.Data == nil {
		return response.ErrorWithMessage("member is nil", memberTable.ProcessType)
	}
	if betAmount > memberTable.Data.MaxBetAmount {
		return response.ErrorIE(errors.MINI_GAME_TABLE_ERROR, errors.IEID_MAX_BET_LIMIT, memberTable.ProcessType)
	}
	return nil
}

func (memberTable *MemberTable) ValidateMinBet(betAmount float64) *response.ErrorData {
	if memberTable.Data == nil {
		return response.ErrorWithMessage("member is nil", memberTable.ProcessType)
	}
	if betAmount < memberTable.Data.MinBetAmount {
		return response.ErrorIE(errors.MINI_GAME_TABLE_ERROR, errors.IEID_MIN_BET_LIMIT, memberTable.ProcessType)
	}
	return nil
}

func (memberTable *MemberTable) ValidateMaxPayout(betAmount float64, euroOdds float64) *response.ErrorData {
	if memberTable.Data == nil {
		return response.ErrorWithMessage("member is nil", memberTable.ProcessType)
	}

	if betAmount > memberTable.Data.MaxPayoutAmount/euroOdds {
		return response.ErrorIE(errors.MINI_GAME_TABLE_ERROR, errors.IEID_MEMBER_MAX_PAYOUT_LIMIT, memberTable.ProcessType)
	}
	return nil
}
