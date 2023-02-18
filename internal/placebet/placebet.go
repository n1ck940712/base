package placebet

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/cache"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

var (
	letters                                             = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	iTicketService   service.TicketService              = *service.NewTicket()
	ILolTowerService service.LolTowerMemberLevelService = *service.NewLolTowerMemberLevelService()
	iEventService    service.EventService               = service.NewEvent()
	GameID           int
	TableID          int64
)

type Placebet struct {
	mu              sync.Mutex
	event           *models.Event
	userMgDetails   *models.User
	reqTicket       *gamemanager.ReqTicket
	ticketSelection *gamemanager.TicketSelection
	ticket          *models.Ticket
	comboTicket     *models.ComboTicket
	lock            *process.Lock
}

func NewPlacebet() *Placebet {
	return &Placebet{lock: process.NewLock()}
}

func (p *Placebet) Lock() bool {
	return p.lock.Lock()
}

func (p *Placebet) Unlock() {
	p.lock.Unlock()
}

// Placebet - main
func (p *Placebet) ProcessTicket(value map[string]interface{}) (*models.ResponseData, errors.FinalErrorMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cleanup()
	gameManager, ppError := p.prepareProcess("bet", value)

	if ppError != nil {
		return nil, ppError
	}

	if err := p.prepareTicket(); err != nil {
		return nil, err
	}

	switch GameID {
	case constants.LOL_TOWER_GAME_ID:
		if err := p.prepareComboTicket(1); err != nil {
			return nil, err
		}
	}
	if err := p.createTicket(gameManager); err != nil {
		return nil, err
	}

	jOutData := models.ResponseData{}
	switch GameID {
	case constants.LOL_TOWER_GAME_ID:
		gameManager.GenerateTowerMemberLevel(*p.comboTicket, p.userMgDetails, false, p.reqTicket.Tickets[0].Level, constants.LOL_TOWER_SKIP_COUNT)
		jOutData = models.ResponseData{Tickets: []models.TicketResponse{
			{
				ID:            p.ticket.ID,
				ComboTicketID: p.comboTicket.ID,
				Amount:        string(types.Float(p.ticket.Amount).String()),
				Ctime:         p.ticket.Ctime,
				EventID:       p.ticket.EventID,
				GameID:        constants.LOL_TOWER_GAME_ID,
				MarketType:    p.ticket.MarketType,
				Odds:          types.Float(p.ticket.Odds).FixedStr(2).Float().Float32(),
				ReferenceNo:   p.ticket.ReferenceNo,
				Selection:     p.ticket.Selection,
				SelectionData: p.ticket.SelectionData,
				Status:        p.ticket.Status,
				TableID:       int16(p.ticket.TableID),
			},
		},
			UserID: types.Bytes(fmt.Sprintf("%v%v", p.userMgDetails.EsportsID, p.userMgDetails.EsportsPartnerID)).SHA256(),
		}
	}

	go CheckDailyWinLoss()
	return &jOutData, nil
}

/*
Placebet - selection
response can be:
- models.ComboTicket
- models.LolTowerMemberLevel
*/
func (p *Placebet) ProcessSelection(value map[string]interface{}) (interface{}, errors.FinalErrorMessage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cleanup()
	gameManager, ppErr := p.prepareProcess("selection", value)
	if ppErr != nil {
		return nil, ppErr
	}

	logger.Info("request value: ", utils.PrettyJSON(value))
	data := value["data"].(map[string]interface{})

	if err := p.validateSelectionRequest(gameManager, data, false); err != nil {
		return nil, err
	}

	// New ticket: task handler for event done > check all ticket if loss then set to settle pending payout
	selection, ok := data["selection"].(string)

	if !ok { //selection is not string
		return nil, errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
	}
	switch selection {
	case "1", "2", "3", "4", "5":
		if (p.ticketSelection.Level - int16(p.ticketSelection.Skip)) > (constants.LOL_TOWER_MAX_LEVEL - constants.LOL_TOWER_SKIP_COUNT) {
			logger.Warning("!!WARNING -- Max level has been reach")
			curStatus := p.ticketSelection.ActiveTicket.Status
			if curStatus == constants.TICKET_STATUS_PAYMENT_CONFIRMED { // auto settle if max level is reached
				res, _ := iTicketService.UpdateTicketStatus(p.ticketSelection.ActiveTicket.ID, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT)
				return res, nil
			}

			return nil, errors.FinalizeErrorMessage(errors.VALIDATE_GAME_ERROR, errors.IEID_MAX_LOL_TOWER_LIMIT, false)
		}
		p.ticket = &p.ticketSelection.ActiveTicket
		result, winLossAmount, payoutAmount := p.generateResult(p.ticket.EventID, p.ticket.Selection, p.ticket.Odds, p.ticket.Amount)
		p.ticket.Result = &result
		p.ticket.WinLossAmount = &winLossAmount
		p.ticket.PayoutAmount = &payoutAmount
		if err := p.prepareComboTicket(int(p.ticketSelection.Level)); err != nil {
			return nil, err
		}
		//quick fix bug on selection p inserted - todo checking bug
		if p.ticket.Selection == "p" || p.comboTicket.Selection == "p" || p.ticket.Selection == "s" || p.comboTicket.Selection == "s" {
			logger.Info("selection should only 1-5 data: ", utils.PrettyJSON(value))
			logger.Info("selection should only 1-5 comboTicket: ", utils.PrettyJSON(p.comboTicket))
			logger.Info("selection should only 1-5 user: ", utils.PrettyJSON(p.userMgDetails))
			return nil, errors.FinalizeErrorMessage(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, false)
		}
		if err := p.createComboTicket(); err != nil {
			return nil, err
		}
		iTicketService.UpdateTicket(p.ticket, models.Ticket{
			EventID:       p.comboTicket.EventID, //update event_id to latest selection
			Odds:          p.ticket.Odds,
			EuroOdds:      p.ticket.EuroOdds,
			Result:        p.ticket.Result,
			WinLossAmount: p.ticket.WinLossAmount,
			PayoutAmount:  p.ticket.PayoutAmount,
			Selection:     p.ticket.Selection,
		})
		gameManager.GenerateTowerMemberLevel(*p.comboTicket, p.userMgDetails, false, p.ticketSelection.Level, p.ticketSelection.Skip)
		//remove attribute before passing to res interface
		jOutData := models.ResponseData{
			Tickets: []models.TicketResponse{
				{
					ID:            p.ticket.ID,
					ComboTicketID: p.comboTicket.ID,
					Amount:        string(types.Int(p.comboTicket.Amount).String()),
					Ctime:         p.comboTicket.Ctime,
					EventID:       p.comboTicket.EventID,
					GameID:        constants.LOL_TOWER_GAME_ID,
					MarketType:    p.comboTicket.MarketType,
					Odds:          types.Float(p.ticket.Odds).FixedStr(2).Float().Float32(),
					ReferenceNo:   p.comboTicket.ReferenceNo,
					Selection:     p.comboTicket.Selection,
					SelectionData: p.comboTicket.SelectionData,
					Status:        p.comboTicket.Status,
					TableID:       int16(p.comboTicket.TableID),
				},
			},
			UserID: types.Bytes(fmt.Sprintf("%v%v", p.userMgDetails.EsportsID, p.userMgDetails.EsportsPartnerID)).SHA256(),
		}
		go CheckDailyWinLoss()
		return jOutData, nil
	case "s":
		if p.ticketSelection.Skip == 0 {
			logger.Warning("!!WARNING -- player skip limit reach")
			return nil, errors.FinalizeErrorMessage(errors.SKIP_LIMIT_REACHED, errors.IEID_SKIP_LIMIT_REACHED, false)
		}
		//selection-level-skipped
		logger.Debug("Player skipped the round ----- ")
		//insert - skipped combo ticket
		p.ticket = &p.ticketSelection.ActiveTicket
		p.ticket.Selection = constants.LOL_SKIP_SELECTION
		p.ticket.Odds = p.ticketSelection.ComboTicket.Odds
		p.ticket.EuroOdds = p.ticketSelection.ComboTicket.EuroOdds
		p.ticket.Result = p.ticketSelection.ComboTicket.Result
		p.ticket.WinLossAmount = p.ticketSelection.ComboTicket.WinLossAmount
		p.ticket.PayoutAmount = p.ticketSelection.ComboTicket.PayoutAmount
		if err := p.prepareComboTicket(int(p.ticketSelection.Level)); err != nil {
			return nil, err
		}
		if err := p.createComboTicket(); err != nil {
			return nil, err
		}
		mgl := gameManager.GenerateTowerMemberLevel(*p.comboTicket, p.userMgDetails, true, p.ticketSelection.Level, p.ticketSelection.Skip)
		mgl.Level = mgl.Level + 1 // when skip return future level

		return mgl, nil
	case "p":
		p.ticket = &p.ticketSelection.ActiveTicket
		ticketIDs := []string{p.ticketSelection.ActiveTicket.ID}
		ticket, _ := iTicketService.UpdateTicketStatus(p.ticketSelection.ActiveTicket.ID, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT)
		iTicketService.UpdateComboTickets(ticketIDs, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT)
		jOutData := models.ResponseData{
			Tickets: []models.TicketResponse{
				{
					ID:            p.ticket.ID,
					Amount:        string(types.Int(ticket.Amount).String()),
					Ctime:         ticket.Ctime,
					EventID:       ticket.EventID,
					GameID:        constants.LOL_TOWER_GAME_ID,
					MarketType:    ticket.MarketType,
					Odds:          types.Float(ticket.Odds).FixedStr(2).Float().Float32(),
					ReferenceNo:   ticket.ReferenceNo,
					Selection:     constants.LOL_PAYOUT_SELECTION,
					SelectionData: ticket.SelectionData,
					Status:        ticket.Status,
					TableID:       int16(ticket.TableID),
				},
			},
			UserID: types.Bytes(fmt.Sprintf("%v%v", p.userMgDetails.EsportsID, p.userMgDetails.EsportsPartnerID)).SHA256(),
		}

		// proccess daily max winnings
		go gameManager.ProcessDailyMaxWinnings()

		return jOutData, nil
	case constants.CUSTOM_WIN_SELECTION, constants.CUSTOM_LOSS_SELECTION:
		return p.processCustomSelection(gameManager)
	default:
		return nil, errors.FinalizeErrorMessage(errors.VALIDATE_SELECTION_ERROR, errors.IEID_INVALID_SELECTION, false)
	}
}

func (p *Placebet) processCustomSelection(gameManager gamemanager.IGameManager) (interface{}, errors.FinalErrorMessage) {
	p.ticket = &p.ticketSelection.ActiveTicket
	sel := gameManager.GetResult(p.ticket.Selection, p.ticket.EventID).(string)
	result, winLossAmount, payoutAmount := p.generateResult(p.ticket.EventID, sel, p.ticket.Odds, p.ticket.Amount)
	p.ticket.Result = &result
	p.ticket.WinLossAmount = &winLossAmount
	p.ticket.PayoutAmount = &payoutAmount
	p.ticket.Selection = sel
	// p.ticketSelection.ActiveTicket.Selection = sel // use win or loss selection
	pcErr := p.prepareComboTicket(int(p.ticketSelection.Level))
	if pcErr != nil {
		return nil, pcErr
	}
	ccErr := p.createComboTicket()
	if ccErr != nil {
		return nil, ccErr
	}
	iTicketService.UpdateTicket(p.ticket, models.Ticket{
		EventID:       p.comboTicket.EventID,
		Odds:          p.ticket.Odds,
		EuroOdds:      p.ticket.EuroOdds,
		Result:        p.ticket.Result,
		WinLossAmount: p.ticket.WinLossAmount,
		PayoutAmount:  p.ticket.PayoutAmount,
	})
	gameManager.GenerateTowerMemberLevel(*p.comboTicket, p.userMgDetails, false, p.ticketSelection.Level, p.ticketSelection.Skip)
	//remove attribute before passing to res interface
	jOutData := models.ResponseData{Tickets: []models.TicketResponse{
		{
			ID:            p.ticket.ID,
			ComboTicketID: p.comboTicket.ID,
			Amount:        string(types.Int(p.ticket.Amount).String()),
			Ctime:         p.ticket.Ctime,
			EventID:       p.ticket.EventID,
			GameID:        constants.LOL_TOWER_GAME_ID,
			MarketType:    p.ticket.MarketType,
			Odds:          types.Float(p.ticket.Odds).FixedStr(2).Float().Float32(),
			ReferenceNo:   p.ticket.ReferenceNo,
			Selection:     p.comboTicket.Selection,
			SelectionData: p.ticket.SelectionData,
			Status:        p.ticket.Status,
			TableID:       int16(p.ticket.TableID),
		},
	},
		UserID: types.Bytes(fmt.Sprintf("%v%v", p.userMgDetails.EsportsID, p.userMgDetails.EsportsPartnerID)).SHA256(),
	}

	return jOutData, nil
}

/* initialize data */
func (p *Placebet) prepareProcess(mtype string, value map[string]interface{}) (gamemanager.IGameManager, errors.FinalErrorMessage) {
	if gameID, ok := value["gameID"].(int); ok {
		GameID = gameID
	}
	if tableID, ok := value["tableID"].(int64); ok {
		TableID = tableID
	}
	if user, ok := value["user"].(*models.User); ok {
		p.userMgDetails = user
	}
	event, eErr := cache.GetEvent(TableID)

	if eErr != nil {
		return nil, errors.FinalizeErrorMessage(errors.EVENT_STATUS_ERROR, errors.IEID_EVENT_NOT_FOUND, false)
	}
	p.event = event

	if mtype == "bet" {
		p.overrideBetValueIfNeeded(&value)
	}

	param := map[string]interface{}{
		"user_id":  p.userMgDetails.ID,
		"table_id": TableID,
		"data":     value["data"],
	}

	gameManager := gamemanager.NewGameManager(TableID)
	ticket, selection, err := gameManager.HandleTicket(mtype, param, p.userMgDetails)
	if err != nil {
		return nil, err
	}

	if ticket != nil {
		p.reqTicket = ticket
	}
	if selection != nil {
		p.ticketSelection = selection
	}

	return gameManager, nil
}

func (p *Placebet) prepareTicket() errors.FinalErrorMessage {
	event, err := cache.GetEvent(TableID)
	if err != nil {
		logger.Error("Event is nil")
		return errors.FinalizeErrorMessage(errors.EVENT_STATUS_ERROR, errors.IEID_EVENT_NOT_FOUND, false)
	}
	ticketDetails := p.reqTicket.Tickets[0]
	selection := fmt.Sprintf("%v", ticketDetails.Selection)
	result, winLossAmount, payoutAmount := p.generateResult(*event.ID, selection, ticketDetails.HongkongOdds, ticketDetails.Amount)
	exchnangeRate := float64(p.userMgDetails.ExchangeRate)
	p.ticket = &models.Ticket{
		ID:               p.generateTicketID(),
		MarketType:       ticketDetails.MarketType,
		Amount:           float64(ticketDetails.Amount),
		EuroOdds:         ticketDetails.EuroOdds,
		Selection:        selection,
		EventID:          *event.ID,
		TableID:          p.reqTicket.TableID,
		UserID:           p.userMgDetails.ID,
		Result:           &result,
		WinLossAmount:    &winLossAmount,
		PayoutAmount:     &payoutAmount,
		ReferenceNo:      ticketDetails.ReferenceNo,
		Odds:             ticketDetails.HongkongOdds, // hongkong odds default
		SelectionData:    "{}",
		StatusMtime:      time.Now(),
		IpAddress:        p.userMgDetails.GetRequestIPAddress(),
		Status:           constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		RequestSource:    p.userMgDetails.GetRequestSource(),
		RequestUserAgent: p.userMgDetails.GetRequestUserAgent(),
		ExchangeRate:     &exchnangeRate,
	}
	return nil
}

func (p *Placebet) createTicket(gameManager gamemanager.IGameManager) errors.FinalErrorMessage {
	validationErr := p.ValidateRequest(false, true, gameManager) //1st validation

	if validationErr != nil {
		return validationErr
	}

	initWTResponse, initWTErr := p.initWalletTransaction() //deduct wallet

	if initWTErr != nil {
		return initWTErr
	}

	_, commitTWErr := p.commitTransaction(initWTResponse) //commit wallet transaction

	if commitTWErr != nil {
		return commitTWErr
	}

	err := iTicketService.CreateTicket(p.ticket, p.comboTicket, TableID, func() error {
		if err := p.ValidateRequest(true, false, gameManager); err != nil { //2nd validation
			validationErr = err
		}
		return nil
	}) //create ticket

	if err != nil {
		logger.Error("Create mini-game-ticket error: ", err.Error())
		_, rollbackWTErr := p.rollbackTransaction(initWTResponse) //rollback wallet transaction

		if rollbackWTErr != nil {
			return rollbackWTErr
		}

		if validationErr != nil {
			return validationErr
		}

		return errors.FinalizeErrorMessage(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, false)
	}
	logger.Info("Create mini-game-ticket width id: ", p.ticket.ID)
	return nil
}

func (p *Placebet) prepareComboTicket(level int) errors.FinalErrorMessage {
	if p.ticket == nil {
		logger.Error("Main ticket is nil")
		return errors.FinalizeErrorMessage(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, false)
	}
	euroOdds := 0.0
	if level == 1 {
		euroOdds = settings.LOL_LEVELS[level]
	} else {
		euroOdds = types.Float(settings.LOL_LEVELS[level] / settings.LOL_LEVELS[level-1]).Round(2).Float64()
	}
	hkOdds := *types.Odds(euroOdds).EuroToHK(2).Ptr()
	exchnangeRate := float64(p.userMgDetails.ExchangeRate)
	winLossAmount := -p.ticket.Amount                    //default loss
	payoutAmount := 0.0                                  //default loss
	if p.ticket.Result != nil && *p.ticket.Result == 0 { //win
		winLossAmount = hkOdds * p.ticket.Amount
		payoutAmount = p.ticket.Amount + winLossAmount
	}
	p.comboTicket = &models.ComboTicket{
		ID:               p.generateTicketID(),
		TicketID:         p.ticket.ID,
		MarketType:       p.ticket.MarketType,
		Amount:           p.ticket.Amount,
		EuroOdds:         euroOdds,
		Selection:        p.ticket.Selection,
		EventID:          p.ticket.EventID,
		TableID:          p.ticket.TableID,
		UserID:           p.ticket.UserID,
		ReferenceNo:      p.ticket.ReferenceNo,
		Result:           p.ticket.Result,
		WinLossAmount:    &winLossAmount,
		PayoutAmount:     &payoutAmount,
		Odds:             hkOdds,
		SelectionData:    p.ticket.SelectionData,
		AutoPlayID:       nil,
		Payload:          nil,
		StatusMtime:      p.ticket.StatusMtime,
		IpAddress:        p.ticket.IpAddress,
		Status:           constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		RequestSource:    p.ticket.RequestSource,
		RequestUserAgent: p.ticket.RequestUserAgent,
		ExchangeRate:     &exchnangeRate,
	}
	return nil
}

func (p *Placebet) createComboTicket() errors.FinalErrorMessage {
	if p.comboTicket == nil {
		logger.Error("Combo ticket is nil")
		return errors.FinalizeErrorMessage(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, false)
	}

	logger.Debug("Creating combo ticket")
	err := iTicketService.CreateComboTicket(p.comboTicket, TableID)
	if err != nil {
		return errors.FinalizeErrorMessage(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, false)
	}
	return nil
}

// response: ticket { result, win_loss_amount, payout_amount }
func (p *Placebet) generateResult(eventID int64, selection string, odds float64, amount float64) (int16, float64, float64) {
	resultValue := iEventService.GetEventLolTower(eventID)

	if utils.Contains(strings.Split(resultValue, ","), selection) {
		winLossAmount := odds * amount
		payoutAmount := amount + winLossAmount

		return 0, winLossAmount, payoutAmount
	}

	return 1, -amount, 0
}

func (p *Placebet) generateTicketID() string {
	m_id := p.userMgDetails.EsportsID
	curYear := time.Now().Year()
	year_tens_value := math.Floor(float64(curYear%100) / float64(10))
	year_ones_value := curYear % 10
	yr1 := getASCII(int(year_tens_value))
	yr2 := getASCII(int(year_ones_value))
	mo1 := getASCII(int(time.Now().Month()) + 12)

	day_tens_value := math.Floor(float64(time.Now().Day()) / float64(10))
	day_ones_value := time.Now().Day() % 10
	day1 := getASCII(int(day_tens_value))
	day2 := getASCII(int(day_ones_value))
	hr := time.Now().Hour()
	m := timeAppendPrefixZero(time.Now().Minute())
	s := timeAppendPrefixZero(time.Now().Second())

	r := randSeq(4)

	base_ticket_id := fmt.Sprintf("%v%v%v%v%v%v%v%v%vMG%v", yr1, yr2, mo1, day1, day2, m_id, hr, m, s, r)

	secretKey := settings.APP_SECRET_KEY
	base_ticket_id = base_ticket_id + secretKey[0:4]
	return strings.ToUpper(base_ticket_id)
}

func (p *Placebet) overrideBetValueIfNeeded(value *map[string]interface{}) {
	if utils.InArray(settings.ENVIRONMENT, []string{"dev", "local"}) {
		TableID = (*value)["tableID"].(int64)
		jsonValue, _ := json.Marshal((*value)["data"])
		updatedJSONStr := strings.Replace(string(jsonValue), `"event_id":"auto"`, fmt.Sprintf(`"event_id":%d`, *p.event.ID), 1)
		var updatedData map[string]any
		if err := json.Unmarshal([]byte(updatedJSONStr), &updatedData); err != nil {
			return
		}
		(*value)["data"] = updatedData
	}
}

func (p *Placebet) cleanup() {
	p.ticket = nil
	p.comboTicket = nil
}

func convertInt64ToString(i *int64) *string {
	if i == nil {
		return nil
	}
	s := string(types.Int(*i).String())
	return &s
}

func CheckDailyWinLoss() {
	if err := api.NewAPI(settings.GetEBOAPI().String() + "/mini-game/admin/v1/check-daily-winloss/11/").
		SetIdentifier(constants_loltower.Identifier + " CheckDailyWinLoss").
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": "Token " + settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		Post(nil); err != nil {
		if err.GetResponse() != nil && err.GetResponse().StatusCode >= 500 {
			go slack.SendPayload(slack.NewLootboxNotification(
				slack.IdentifierToTitle(constants_loltower.Identifier)+"api CheckDailyWinLoss",
				fmt.Sprint("*Status Code:* \n> *", err.GetResponse().StatusCode, "*\n*Response Body:* \n> *", err.GetResponseBody(), "*"),
			), slack.LootboxHealthCheck)
		}
	}
}
