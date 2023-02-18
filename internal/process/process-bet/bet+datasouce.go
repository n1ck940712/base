package process_bet

import (
	"fmt"
	"strings"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process/validate"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type GetBetOpenRangeCallback = func() *BetOpenRange
type GetBetResultCallback = func(betTicket request.BetTicketData) *BetResult

type BetParamsDatasource interface {
	GetBetResult(betTicket request.BetTicketData) *BetResult
	GetBetOpenRange() *BetOpenRange
}

type BetProcessDatasource interface {
	SetProcessType(processType string)
	PrePrepareValidation() *response.ErrorData
	PrepareBet(betData *request.BetData) *response.ErrorData
	PrepareTickets() *response.ErrorData
	PrepareComboTickets() *response.ErrorData
	PreCreateTicketValidation() *response.ErrorData
	PreCreateComboTicketValidation() *response.ErrorData
	OnCreateTicketValidation() *response.ErrorData
	OnCreateComboTicketValidation() *response.ErrorData
	CreateTickets(callback func(tx *gorm.DB) error) *response.ErrorData
	CreateTicketsWithCombo(callback func(tx *gorm.DB) error) *response.ErrorData
	CreateComboTickets(callback func(tx *gorm.DB) error) *response.ErrorData
	Payout(selection string) *response.ErrorData
	GetBet() *request.BetData                      //get bet data
	GetTickets() *[]models.Ticket                  //get tickets
	GetComboTickets() *[]models.ComboTicket        //get combo tickets
	GetBetTickets() *[]response.BetTicketData      //tickets for response
	GetBetComboTickets() *[]response.BetTicketData //combo ticket for response
	GetEncryptedUserID() string                    //get encrypted user id
	LoadActiveTickets() *[]models.Ticket           //load active tickets to tickets
	CheckDailyWinLoss()
	CleanUp()

	//setter
	SetBetOpenRangeCallback(callback GetBetOpenRangeCallback)
	SetBetResultCallback(callback GetBetResultCallback)
}

type betProcessDatasource struct {
	processType      string
	datasource       BetDatasouce
	betData          *request.BetData
	tickets          *[]models.Ticket
	comboTickets     *[]models.ComboTicket
	eventTicket      *models.Ticket
	eventComboTicket *models.ComboTicket

	//callbacks
	getBetOpenRangeCallback *GetBetOpenRangeCallback
	getBetResultCallback    *GetBetResultCallback
}

func NewBetProcessDatasource(datasource BetDatasouce) BetProcessDatasource {
	return &betProcessDatasource{processType: BetType, datasource: datasource}
}

func (bpd *betProcessDatasource) SetProcessType(processType string) {
	bpd.processType = processType
}

func (bpd *betProcessDatasource) PrePrepareValidation() *response.ErrorData {
	validateUser := validate.User{ProcessType: bpd.processType, Data: bpd.datasource.GetUser()}
	validateGameTable := validate.GameTable{ProcessType: bpd.processType, Data: bpd.datasource.GetGameTable()}
	validateMemberTable := validate.MemberTable{ProcessType: bpd.processType, Data: bpd.datasource.GetMemberTable()}
	validateEvent := validate.Event{ProcessType: bpd.processType, Data: bpd.datasource.GetEvent()}

	if err := validateUser.ValidateStatus(); err != nil {
		return err
	}
	if err := validateGameTable.ValidateEnabled(); err != nil {
		return err
	}
	if err := validateMemberTable.ValidateEnabled(); err != nil {
		return err
	}
	if bpd.getBetOpenRangeCallback == nil {
		panic("SetBetOpenRangeCallback method must bet set")
	} else if betOpen := (*bpd.getBetOpenRangeCallback)(); betOpen != nil {
		if err := validateEvent.ValidateIsOpen(betOpen.MinMS, betOpen.MaxMS); err != nil {
			return err
		}
	}
	return nil
}

func (bpd *betProcessDatasource) PrepareBet(betData *request.BetData) *response.ErrorData {
	bpd.betData = betData
	return nil
}

func (bpd *betProcessDatasource) PrepareTickets() *response.ErrorData {
	if bpd.betData == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareBet not called")
	}
	user := bpd.datasource.GetUser()
	tickets := []models.Ticket{}

	for i := 0; i < len(bpd.betData.Tickets); i++ {
		betResult := bpd.GetBetResult(bpd.betData.Tickets[i])

		tickets = append(tickets, models.Ticket{
			ID:               GenerateTicketID(user.EsportsID),
			Amount:           bpd.betData.Tickets[i].Amount.ToFloat64(),
			EuroOdds:         betResult.EuroOdds,
			Odds:             betResult.HongkongOdds,
			Selection:        bpd.betData.Tickets[i].Selection.String(),
			SelectionData:    bpd.betData.Tickets[i].GetSelectionData(),
			EventID:          bpd.betData.EventID.Int64(),
			TableID:          bpd.betData.TableID.Int64(),
			MarketType:       bpd.betData.Tickets[i].MarketType.Int16(),
			ReferenceNo:      GenerateReferenceNo(bpd.betData.Tickets[i].ReferenceNo.String()),
			StatusMtime:      TimeNow(),
			Status:           DefaultCreatedTicketStatus(),
			UserID:           user.ID,
			IpAddress:        user.GetRequestIPAddress(),
			RequestSource:    user.GetRequestSource(),
			RequestUserAgent: user.GetRequestUserAgent(),
			ExchangeRate:     &user.ExchangeRate,
			Result:           &betResult.Result,
			WinLossAmount:    &betResult.WinLossAmount,
			PayoutAmount:     &betResult.PayoutAmount,
			OriginalOdds:     &betResult.OriginalOdds,
		})
	}
	bpd.tickets = &tickets
	return nil
}

func (bpd *betProcessDatasource) PrepareComboTickets() *response.ErrorData {
	if bpd.betData == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareBet not called")
	}
	if bpd.tickets == nil {
		return response.ErrorWithMessage(bpd.processType, "internal tickets was not supplied")
	}
	user := bpd.datasource.GetUser()
	comboTickets := []models.ComboTicket{}

	for i := 0; i < len(bpd.betData.Tickets); i++ {
		betResult := bpd.GetBetResult(bpd.betData.Tickets[i])

		comboTickets = append(comboTickets, models.ComboTicket{
			ID:                     GenerateTicketID(user.EsportsID),
			TicketID:               (*bpd.tickets)[0].ID,
			Amount:                 bpd.betData.Tickets[i].Amount.ToFloat64(),
			EuroOdds:               betResult.EuroOdds,
			Odds:                   betResult.HongkongOdds,
			Selection:              bpd.betData.Tickets[i].Selection.String(),
			SelectionData:          bpd.betData.Tickets[i].GetSelectionData(),
			EventID:                bpd.betData.EventID.Int64(),
			TableID:                bpd.betData.TableID.Int64(),
			MarketType:             bpd.betData.Tickets[i].MarketType.Int16(),
			StatusMtime:            TimeNow(),
			Status:                 DefaultCreatedTicketStatus(),
			ReferenceNo:            GenerateReferenceNo(bpd.betData.Tickets[i].ReferenceNo.String()),
			UserID:                 user.ID,
			IpAddress:              user.GetRequestIPAddress(),
			RequestSource:          user.GetRequestSource(),
			RequestUserAgent:       user.GetRequestUserAgent(),
			ExchangeRate:           &user.ExchangeRate,
			Result:                 &betResult.Result,
			WinLossAmount:          &betResult.WinLossAmount,
			PayoutAmount:           &betResult.PayoutAmount,
			OriginalOdds:           &betResult.OriginalOdds,
			PossibleWinningsAmount: &betResult.PossibleWinningsAmount,
		})
	}
	bpd.comboTickets = &comboTickets
	return nil
}

func (bpd *betProcessDatasource) GetBetResult(betTicket request.BetTicketData) *BetResult {
	if bpd.getBetResultCallback != nil {
		return (*bpd.getBetResultCallback)(betTicket)
	}
	betAmount := betTicket.Amount.ToFloat64()
	betResult := BetResult{
		Result:        1,
		WinLossAmount: -betAmount,
		PayoutAmount:  0,
		HongkongOdds:  betTicket.HongkongOdds,
		EuroOdds:      betTicket.EuroOdds,
		MalayOdds:     betTicket.MalayOdds,
		OriginalOdds:  betTicket.HongkongOdds,
	}

	if eventResults := bpd.datasource.GetEventResults(); eventResults != nil {
		for i := 0; i < len(*eventResults); i++ {
			if types.Array[string](strings.Split((*eventResults)[i].Value, ",")).Constains(betTicket.Selection.String()) {
				winLossAmount := betTicket.HongkongOdds * betAmount
				payoutAmount := betAmount + winLossAmount

				betResult.Result = 0
				betResult.WinLossAmount = winLossAmount
				betResult.PayoutAmount = payoutAmount
				break
			}
		}
	}
	return &betResult
}

func (bpd *betProcessDatasource) PreCreateTicketValidation() *response.ErrorData {
	if bpd.betData == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareBet not called")
	}
	betEventID := bpd.betData.EventID.Int64()
	validateUser := validate.User{ProcessType: bpd.processType, Data: bpd.datasource.GetUser()}
	validateEvent := validate.Event{ProcessType: bpd.processType, Data: bpd.datasource.GetEvent()}
	validateMemberTable := validate.MemberTable{ProcessType: bpd.processType, Data: bpd.datasource.GetMemberTable()}

	if err := validateEvent.ValidateBetEvent(betEventID); err != nil {
		return err
	}
	if err := validateEvent.ValidateTicket(validate.Ticket{ProcessType: bpd.processType, Data: bpd.GetEventTicket()}); err != nil {
		return err
	}
	if err := validateUser.ValidateBalance(bpd.betData.Tickets.TotalAmount()); err != nil {
		return err
	}
	for i := 0; i < len(bpd.betData.Tickets); i++ {
		ticket := bpd.betData.Tickets[i]
		betAmount := ticket.Amount.ToFloat64()

		if err := validateMemberTable.ValidateMaxBet(betAmount); err != nil {
			return err
		}
		if err := validateMemberTable.ValidateMinBet(betAmount); err != nil {
			return err
		}
		if err := validateMemberTable.ValidateMaxPayout(betAmount, ticket.MaxPayoutEuroOdds); err != nil {
			return err
		}
	}
	return nil
}

func (bpd *betProcessDatasource) PreCreateComboTicketValidation() *response.ErrorData {
	if bpd.betData == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareBet not called")
	}
	betEventID := bpd.betData.EventID.Int64()
	validateEvent := validate.Event{ProcessType: bpd.processType, Data: bpd.datasource.GetEvent()}
	validateMemberTable := validate.MemberTable{ProcessType: bpd.processType, Data: bpd.datasource.GetMemberTable()}

	if err := validateEvent.ValidateBetEvent(betEventID); err != nil {
		return err
	}
	if err := validateEvent.ValidateComboTicket(validate.ComboTicket{ProcessType: "bet", Data: bpd.GetEventComboTicket()}); err != nil {
		return err
	}
	for i := 0; i < len(bpd.betData.Tickets); i++ {
		ticket := bpd.betData.Tickets[i]
		betAmount := ticket.Amount.ToFloat64()

		if err := validateMemberTable.ValidateMaxBet(betAmount); err != nil {
			return err
		}
		if err := validateMemberTable.ValidateMinBet(betAmount); err != nil {
			return err
		}
		if err := validateMemberTable.ValidateMaxPayout(betAmount, ticket.MaxPayoutEuroOdds); err != nil {
			return err
		}
	}
	return nil
}

func (bpd *betProcessDatasource) OnCreateTicketValidation() *response.ErrorData {
	validateEvent := validate.Event{ProcessType: bpd.processType, Data: bpd.datasource.GetEvent()}

	if bpd.getBetOpenRangeCallback == nil {
		panic("SetBetOpenRangeCallback method must bet set")
	} else if betOpen := (*bpd.getBetOpenRangeCallback)(); betOpen != nil {
		if err := validateEvent.ValidateIsOpen(betOpen.MinMS, betOpen.MaxMS); err != nil {
			return err
		}
	}
	return nil
}

func (bpd *betProcessDatasource) OnCreateComboTicketValidation() *response.ErrorData {
	validateEvent := validate.Event{ProcessType: bpd.processType, Data: bpd.datasource.GetEvent()}

	if bpd.getBetOpenRangeCallback == nil {
		panic("SetBetOpenRangeCallback method must bet set")
	} else if betOpen := (*bpd.getBetOpenRangeCallback)(); betOpen != nil {
		if err := validateEvent.ValidateIsOpen(betOpen.MinMS, betOpen.MaxMS); err != nil {
			return err
		}
	}
	return nil
}

func (bpd *betProcessDatasource) CreateTickets(callback func(tx *gorm.DB) error) *response.ErrorData {
	if bpd.tickets == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareTickets not called")
	}
	if err := db.Shared().Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
		SELECT 
			id
		FROM 
			mini_game_memberminigametable 
		WHERE 
			mini_game_table_id=? 
			AND user_id=? 
		FOR UPDATE
		`, bpd.datasource.GetGameTable().ID, bpd.datasource.GetUser().ID).Error; err != nil {
			return err
		}
		var lockSelect []map[string]interface{}
		if tx.Raw(`
		SELECT
			id
		FROM
			mini_game_ticket
		WHERE
			event_id = ?
			AND user_id = ?
		`, (*bpd.tickets)[0].EventID, (*bpd.tickets)[0].UserID).Scan(&lockSelect).RowsAffected > 0 {
			return response.ErrorIE(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, "bet")
		}
		for i := 0; i < len(*bpd.tickets); i++ {
			if err := tx.Create(&(*bpd.tickets)[i]).Error; err != nil {
				return err
			}
		}
		if err := bpd.OnCreateTicketValidation(); err != nil {
			return err
		}
		return callback(tx)
	}); err != nil {
		if err := err.(*response.ErrorData); err != nil {
			return err
		}
		logger.Info(bpd.datasource.GetIdentifier(), " CreateTickets tickets: ", utils.PrettyJSON(bpd.tickets))
		logger.Error(bpd.datasource.GetIdentifier(), " CreateTickets error: ", err.Error())
		return response.ErrorIE(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, "bet")
	}

	return nil
}

func (bpd *betProcessDatasource) CreateTicketsWithCombo(callback func(tx *gorm.DB) error) *response.ErrorData {
	if bpd.tickets == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareTickets not called")
	}
	if bpd.comboTickets == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareComboTickets not called")
	}
	if err := db.Shared().Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
		SELECT 
			id
		FROM 
			mini_game_memberminigametable 
		WHERE 
			mini_game_table_id=? 
			AND user_id=? 
		FOR UPDATE
		`, bpd.datasource.GetGameTable().ID, bpd.datasource.GetUser().ID).Error; err != nil {
			return err
		}
		var lockSelect []map[string]interface{}
		if tx.Raw(`
			SELECT 
                id
			FROM
				mini_game_ticket
			WHERE
				event_id = ?
				AND user_id = ?
			`, (*bpd.tickets)[0].EventID, (*bpd.tickets)[0].UserID).Scan(&lockSelect).RowsAffected > 0 {
			return response.ErrorIE(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, "bet")
		}
		for i := 0; i < len(*bpd.tickets); i++ {
			if err := tx.Create(&(*bpd.tickets)[i]).Error; err != nil {
				return err
			}
		}
		for i := 0; i < len(*bpd.comboTickets); i++ {
			if err := tx.Create(&(*bpd.comboTickets)[i]).Error; err != nil {
				return err
			}
		}
		if err := bpd.OnCreateTicketValidation(); err != nil {
			return err
		}
		return callback(tx)
	}); err != nil {
		if err := err.(*response.ErrorData); err != nil {
			return err
		}
		logger.Info(bpd.datasource.GetIdentifier(), " CreateTicketsWithCombo tickets: ", utils.PrettyJSON(bpd.tickets))
		logger.Info(bpd.datasource.GetIdentifier(), " CreateTicketsWithCombo combo tickets: ", utils.PrettyJSON(bpd.comboTickets))
		logger.Error(bpd.datasource.GetIdentifier(), " CreateTicketWithCombo error: ", err.Error())
		return response.ErrorIE(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, "bet")
	}
	return nil
}

func (bpd *betProcessDatasource) CreateComboTickets(callback func(tx *gorm.DB) error) *response.ErrorData {
	if bpd.comboTickets == nil {
		return response.ErrorWithMessage(bpd.processType, "internal PrepareComboTicket not called")
	}
	if err := db.Shared().Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
		SELECT 
			id
		FROM 
			mini_game_memberminigametable 
		WHERE 
			mini_game_table_id=? 
			AND user_id=? 
		FOR UPDATE
		`, bpd.datasource.GetGameTable().ID, bpd.datasource.GetUser().ID).Error; err != nil {
			return err
		}
		var lockSelect []map[string]interface{}
		if tx.Raw(`
			SELECT
				id
			FROM
				mini_game_combo_ticket
			WHERE
				event_id = ?
				AND user_id = ?
			`, (*bpd.comboTickets)[0].EventID, (*bpd.comboTickets)[0].UserID).Scan(&lockSelect).RowsAffected > 0 {
			return response.ErrorIE(errors.DUPLICATE_BET, errors.IEID_DUPLICATE_BET, "bet")
		}
		for i := 0; i < len(*bpd.comboTickets); i++ {
			if err := tx.Create(&(*bpd.comboTickets)[i]).Error; err != nil {
				return err
			}
		}
		if err := bpd.OnCreateComboTicketValidation(); err != nil {
			return err
		}
		return callback(tx)
	}); err != nil {
		if err := err.(*response.ErrorData); err != nil {
			return err
		}
		logger.Info(bpd.datasource.GetIdentifier(), " CreateComboTickets combo tickets: ", utils.PrettyJSON(bpd.comboTickets))
		logger.Error(bpd.datasource.GetIdentifier(), " CreateComboTickets error: ", err.Error())
		return response.ErrorIE(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, "bet")
	}
	return nil
}

func (bpd *betProcessDatasource) Payout(selection string) *response.ErrorData {
	if bpd.tickets == nil {
		return response.ErrorWithMessage(bpd.processType, "internal No active tickets")
	}
	ticketIDs := types.Array[models.Ticket](*bpd.tickets).Map(func(value models.Ticket) any { return value.ID })

	if err := db.Shared().Transaction(func(tx *gorm.DB) error {
		if err := tx.Debug().Where("id IN ?", ticketIDs).Updates(models.Ticket{
			// EventID:   *bpd.datasource.GetEvent().ID,
			Selection: selection,
			Status:    constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT,
		}).Error; err != nil {
			return err
		}
		if err := tx.Debug().Where("ticket_id IN ?", ticketIDs).Updates(models.ComboTicket{
			Status: constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		logger.Info(bpd.datasource.GetIdentifier(), " Payout tickets: ", utils.PrettyJSON(bpd.tickets))
		logger.Error(bpd.datasource.GetIdentifier(), " Payout error: ", err.Error())
		return response.ErrorIE(errors.TICKET_CREATION_ERROR, errors.IEID_TICKET_CREATION_ERROR, "bet")
	}
	return nil
}

func (bpd *betProcessDatasource) GetBet() *request.BetData {
	return bpd.betData
}

func (bpd *betProcessDatasource) GetTickets() *[]models.Ticket {
	return bpd.tickets
}

func (bpd *betProcessDatasource) GetComboTickets() *[]models.ComboTicket {
	return bpd.comboTickets
}

func (bpd *betProcessDatasource) GetBetTickets() *[]response.BetTicketData {
	if bpd.tickets == nil {
		return nil
	}
	tickets := []response.BetTicketData{}
	for _, ticket := range *bpd.tickets {
		tickets = append(tickets, response.BetTicketData{
			ID:            ticket.ID,
			Ctime:         ticket.Ctime,
			EventID:       ticket.EventID,
			GameID:        bpd.datasource.GetGameTable().GameID,
			Selection:     ticket.Selection,
			Amount:        string(types.Float(ticket.Amount).FixedStr(2)),
			MarketType:    ticket.MarketType,
			ReferenceNo:   ticket.ReferenceNo,
			Status:        ticket.Status,
			TableID:       ticket.TableID,
			SelectionData: types.String(ticket.SelectionData).JSON(),
			Odds:          *types.Float(ticket.Odds).FixedStr(2).Ptr(),
		})
	}
	return &tickets
}

func (bpd *betProcessDatasource) GetBetComboTickets() *[]response.BetTicketData {
	if bpd.comboTickets == nil {
		return nil
	}
	betComboTickets := []response.BetTicketData{}
	for _, comboTicket := range *bpd.comboTickets {
		betComboTickets = append(betComboTickets, response.BetTicketData{
			ID:            comboTicket.TicketID,
			ComboTicketID: comboTicket.ID,
			Ctime:         comboTicket.Ctime,
			EventID:       comboTicket.EventID,
			GameID:        bpd.datasource.GetGameTable().GameID,
			Selection:     comboTicket.Selection,
			Amount:        string(types.Float(comboTicket.Amount).FixedStr(2)),
			MarketType:    comboTicket.MarketType,
			ReferenceNo:   comboTicket.ReferenceNo,
			Status:        comboTicket.Status,
			TableID:       comboTicket.TableID,
			SelectionData: types.String(comboTicket.SelectionData).JSON(),
			Odds:          *types.Float(comboTicket.Odds).FixedStr(2).Ptr(),
		})
	}
	return &betComboTickets
}

func (bpd *betProcessDatasource) GetEncryptedUserID() string {
	user := bpd.datasource.GetUser()

	return types.Bytes(fmt.Sprintf("%v%v", user.EsportsID, user.EsportsPartnerID)).SHA256()
}

func (bpd *betProcessDatasource) GetEventTicket() *models.Ticket {
	if bpd.eventTicket == nil {
		if err := db.Shared().Where(models.Ticket{
			UserID:  bpd.datasource.GetUser().ID,
			EventID: *bpd.datasource.GetEvent().ID,
			TableID: bpd.datasource.GetGameTable().ID,
		}).Order("ctime DESC").First(&bpd.eventTicket).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				logger.Error(bpd.datasource.GetIdentifier(), " GetEventTicket error: ", err.Error())
			}
			return nil
		}
	}

	return bpd.eventTicket
}

func (bpd *betProcessDatasource) GetEventComboTicket() *models.ComboTicket {
	if bpd.eventComboTicket == nil {
		if err := db.Shared().Where(models.ComboTicket{
			UserID:  bpd.datasource.GetUser().ID,
			EventID: *bpd.datasource.GetEvent().ID,
			TableID: bpd.datasource.GetGameTable().ID,
		}).Order("ctime DESC").First(&bpd.eventComboTicket).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				logger.Error(bpd.datasource.GetIdentifier(), " GetEventComboTicket error: ", err.Error())
			}
		}
	}
	return bpd.eventComboTicket
}

func (bpd *betProcessDatasource) LoadActiveTickets() *[]models.Ticket {
	if bpd.tickets == nil {
		if err := db.Shared().Preload("ComboTickets").Where(models.Ticket{
			UserID:  bpd.datasource.GetUser().ID,
			TableID: bpd.datasource.GetGameTable().ID,
			Status:  constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		}).Order("ctime DESC").Find(&bpd.tickets).Error; err != nil {
			logger.Error(bpd.datasource.GetIdentifier(), " LoadActiveTickets tickets error: ", err.Error())
		}
		if bpd.tickets != nil && len(*bpd.tickets) == 0 {
			bpd.tickets = nil
		}
	}
	return bpd.tickets
}

func (bpd *betProcessDatasource) CheckDailyWinLoss() {
	tableID := types.Int(bpd.datasource.GetGameTable().ID).String()

	if err := api.NewAPI(settings.GetEBOAPI().String() + "/mini-game/admin/v1/check-daily-winloss/" + string(tableID) + "/").
		SetIdentifier(bpd.datasource.GetIdentifier() + " CheckDailyWinLoss").
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": "Token " + settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		Post(nil); err != nil {
		if err.GetResponse() != nil && err.GetResponse().StatusCode >= 500 {
			go slack.SendPayload(slack.NewLootboxNotification(
				slack.IdentifierToTitle(bpd.datasource.GetIdentifier())+"api CheckDailyWinLoss",
				fmt.Sprint("*Status Code:* \n> *", err.GetResponse().StatusCode, "*\n*Response Body:* \n> *", err.GetResponseBody(), "*"),
			), slack.LootboxHealthCheck)
		}
	}
}

func (bpd *betProcessDatasource) CleanUp() {
	bpd.betData = nil
	bpd.tickets = nil
	bpd.comboTickets = nil
	bpd.eventTicket = nil
	bpd.eventComboTicket = nil
}

// setter
func (bpd *betProcessDatasource) SetBetOpenRangeCallback(callback GetBetOpenRangeCallback) {
	bpd.getBetOpenRangeCallback = &callback
}

func (bpd *betProcessDatasource) SetBetResultCallback(callback GetBetResultCallback) {
	bpd.getBetResultCallback = &callback
}
