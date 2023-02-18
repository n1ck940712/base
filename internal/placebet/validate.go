package placebet

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/tablemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
)

func (p *Placebet) ValidateRequest(ignoreHasticket bool, validateBalance bool, gameManager gamemanager.IGameManager) errors.FinalErrorMessage {
	if tableErr := p.validateTableSettings(); tableErr != nil {
		return tableErr
	}

	if balErr := p.validateMemberBalance(validateBalance); balErr != nil {
		return balErr
	}

	if eventErr := p.validateEvent(ignoreHasticket); eventErr != nil {
		return eventErr
	}

	if userErr := p.validateUserStatus(); userErr != nil {
		return userErr
	}

	ticket := p.reqTicket.Tickets[0]

	if selErr := p.validateBetSelection(gameManager, ticket.Selection); selErr != nil {
		return selErr
	}

	if martyErr := p.validateBetMarketType(ticket.MarketType); martyErr != nil {
		return martyErr
	}

	if betErr := p.validateBet(); betErr != nil {
		return betErr
	}

	return nil
}

func (p *Placebet) validateTableSettings() errors.FinalErrorMessage {
	gameTable := models.GameTable{
		ID: TableID,
	}

	service.Get(&gameTable)

	if !gameTable.IsEnabled || gameTable.DisabledByMgTrigger {
		return errors.FinalizeErrorMessage(errors.MINI_GAME_DISABLED, errors.IEID_GAME_DISABLED, false)
	}

	memberTable := models.MemberTable{
		TableID: TableID,
		UserID:  p.userMgDetails.ID,
	}

	service.Get(&memberTable)

	if !memberTable.IsEnabled {
		return errors.FinalizeErrorMessage(errors.MINI_GAME_DISABLED, errors.IEID_GAME_DISABLED, false)
	}

	return nil
}

func (p *Placebet) validateMemberBalance(validateBalance bool) errors.FinalErrorMessage {
	var response map[string]interface{}

	if validateBalance {
		err := api.NewAPI(settings.EBO_API + "/v4/wallet/balance").
			AddHeaders(map[string]string{
				"User-Agent":    settings.USER_AGENT,
				"Authorization": settings.SERVER_TOKEN,
				"Content-Type":  "application/json",
			}).AddQueries(map[string]string{"member_id": fmt.Sprint(p.userMgDetails.EsportsID)}).Get(&response)

		if err != nil {
			return errors.FinalizeErrorMessage(errors.ESPORTS_API_FAILED, errors.IEID_ESPORTS_API_FAILED, false)
		}

		if balance, ok := response["balance"].(float64); ok {
			if p.ticket.Amount > balance {
				return errors.FinalizeErrorMessage(errors.NOT_ENOUGH_BALANCE_EXCEPTION, errors.IEID_NOT_ENOUGH_BALANCE, false)
			}
		}

	}

	return nil
}

func (p *Placebet) validateBetSelection(gameManager gamemanager.IGameManager, selection interface{}) errors.FinalErrorMessage {
	if selErr := gameManager.ValidateSelected(selection, true); selErr != nil {
		return selErr
	}

	return nil
}

func (p *Placebet) validateBetMarketType(marketType int16) errors.FinalErrorMessage {
	if marketType != constants.LOL_TOWER_MARKET_TYPE {
		return errors.FinalizeErrorMessage(errors.VALIDATE_MARKET_TYPE_ERROR, errors.IEID_MARKET_TYPE_ERROR, false)
	}

	return nil
}

func (p *Placebet) validateEvent(ignoreHasticket bool) errors.FinalErrorMessage {
	event := service.NewEvent().IsExist(p.reqTicket.EventID)
	if event == 0 {
		return errors.FinalizeErrorMessage(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_NOT_FOUND, false)
	}

	// check has existing bet on event
	if !ignoreHasticket {
		hasTicket := iTicketService.MemberHasTicketOnEvent(p.reqTicket.EventID, p.userMgDetails.ID)
		if hasTicket {
			return errors.FinalizeErrorMessage(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, false)
		}
		activeTicket := iTicketService.GetMemberActiveTicket(p.userMgDetails.ID)
		if activeTicket != nil {
			return errors.FinalizeErrorMessage(errors.VALIDATE_EVENT_ERROR, errors.IEID_HAS_ACTIVE_MAIN_TICKET, false)
		}
	}

	// check event if still open for bet
	isOpenBet := p.IsEventOpenForBet()
	if !isOpenBet {
		return errors.FinalizeErrorMessage(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_NOT_ON_BETTING_PHASE, false)
	}

	return nil
}

func (p *Placebet) IsEventOpenForBet() bool {
	var gameManager gamemanager.IGameManager = gamemanager.NewGameManager(p.reqTicket.TableID)

	return gameManager.IsEventOpenForBet(p.reqTicket.EventID)
}

func (p *Placebet) validateBet() errors.FinalErrorMessage {
	var tableManager tablemanager.ITableManager = tablemanager.NewTableManager(TableID)
	eventDetails := service.NewEvent().GetEventDetails(p.reqTicket.EventID)
	betAmount := float64(p.reqTicket.Tickets[0].Amount)

	if betAmount/p.userMgDetails.ExchangeRate > eventDetails.MaxBet {
		return errors.FinalizeErrorMessage(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_MAX_BET_LIMIT, false)
	}
	member := tableManager.GetMemberDetails(p.userMgDetails.ID)

	if betAmount > member.MaxBetAmount {
		return errors.FinalizeErrorMessage(errors.MINI_GAME_TABLE_ERROR, errors.IEID_MAX_BET_LIMIT, false)
	}
	if betAmount < member.MinBetAmount {
		return errors.FinalizeErrorMessage(errors.MINI_GAME_TABLE_ERROR, errors.IEID_MIN_BET_LIMIT, false)
	}
	euroOdds := settings.LOL_LEVELS[1]

	if utils.CalculateMaxPayout(betAmount, euroOdds) > member.MaxPayoutAmount {
		return errors.FinalizeErrorMessage(errors.MINI_GAME_TABLE_ERROR, errors.IEID_MEMBER_MAX_PAYOUT_LIMIT, false)
	}

	return nil
}

func (p *Placebet) validateUserStatus() errors.FinalErrorMessage {
	if p.userMgDetails.IsAccountFrozen {
		return errors.FinalizeErrorMessage(errors.ACCOUNT_IS_FROZEN, errors.IEID_ACCOUNT_IS_FROZEN, true)
	}
	if p.userMgDetails.SleepStatus > 0 {
		return errors.FinalizeErrorMessage(errors.USER_IS_IN_SLEEP_STATUS, errors.IEID_ACCOUNT_IS_FROZEN, true)
	}

	return nil
}

func (p *Placebet) validateSelectionRequest(gameManager gamemanager.IGameManager, data map[string]interface{}, ignoreHasticket bool) errors.FinalErrorMessage {
	logger.Info("validateSelectionRequest ----------------->")

	if tableErr := p.validateTableSettings(); tableErr != nil {
		return tableErr
	}

	if selErr := gameManager.ValidateSelected(data["selection"], false); selErr != nil {
		return selErr
	}

	if evnErr := p.validateCurrentEvent(ignoreHasticket); evnErr != nil {
		return evnErr
	}

	return nil
}

func (p *Placebet) validateCurrentEvent(ignoreHasticket bool) errors.FinalErrorMessage {

	// validate betting phase
	logger.Info("validateCurrentEvent ----------------->")
	var gameManager gamemanager.IGameManager = gamemanager.NewGameManager(TableID)
	isOpen := gameManager.IsEventOpenForBet(*p.event.ID)
	if !isOpen {
		return errors.FinalizeErrorMessage(errors.VALIDATE_EVENT_ERROR, errors.IEID_EVENT_NOT_ON_BETTING_PHASE, false)
	}

	// check has existing bet on event
	if !ignoreHasticket {
		hasTicket := iTicketService.MemberHasTicketOnEvent(*p.event.ID, p.userMgDetails.ID)
		if hasTicket {
			return errors.FinalizeErrorMessage(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, false)
		}
	}

	// check combo ticket exists
	activeTicket := iTicketService.GetMemberActiveTicket(p.userMgDetails.ID)
	if activeTicket == nil {
		return errors.FinalizeErrorMessage(errors.VALIDATE_BET_SELECTION_TYPE_ERROR, errors.IEID_NO_ACTIVE_TICKET, false)
	}

	//check if incoming bet selection payout is greaterthan user max payout
	if gameManager.IsGreaterThanMaxPayout(p.userMgDetails.ID) {
		return errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_MEMBER_MAX_PAYOUT_LIMIT, false)
	}

	return nil
}
