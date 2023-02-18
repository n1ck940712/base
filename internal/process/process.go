package process

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet"
	process_config "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-config"
	process_game_data "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-game-data"
	process_member_list "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-member-list"
	process_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds"
	process_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-state"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state"
	"gorm.io/gorm"
)

type ProcessType = string
type GetUserCallback = func() *models.User
type GetEventCallback = func() *models.Event
type GetEventResultsCallback = func() *[]models.EventResult
type GetGameTableCallback = func() *models.GameTable
type GetMemberTableCallback = func() *models.MemberTable
type GetMemberConfigsCallback = func() *[]models.MemberConfig
type GetBetChipsCallback = func() *[]float64
type GetTicketStateCallback = func() *[]models.Ticket

const (
	StateType          ProcessType = process_state.StateType
	ConfigType         ProcessType = process_config.ConfigType
	BetType            ProcessType = process_bet.BetType
	SelectionType      ProcessType = process_bet.SelectionType
	TicketType         ProcessType = process_ticket_state.TicketType
	CurrentTicketsType ProcessType = process_ticket_state.CurrentTicketsType
	OddsType           ProcessType = process_odds.OddsType
	MemberListType     ProcessType = process_member_list.MemberListType
	GameDataType       ProcessType = process_game_data.GameDataType
	ErrorType          ProcessType = "error"
)

// DATASOURCE interface
type Datasource interface {
	GetIdentifier() string
	GetUser() *models.User
	GetGameID() int64
	GetTableID() int64
}

type datasource struct {
	identifier string
	user       *models.User
	gameID     int64
	tableID    int64
}

func NewDatasource(identfier string, user *models.User, gameID int64, tableID int64) Datasource {
	return &datasource{
		identifier: identfier,
		user:       user,
		gameID:     gameID,
		tableID:    tableID,
	}
}

func (ds *datasource) GetIdentifier() string {
	return ds.identifier
}

func (ds *datasource) GetUser() *models.User {
	return ds.user
}

func (ds *datasource) GetGameID() int64 {
	return ds.gameID
}

func (ds *datasource) GetTableID() int64 {
	return ds.tableID
}

// PROCESS GETTER interface
type ProcessRequest interface {
	ProcessRequest(request string) response.Response

	SetEventID(eventID int64)
}

// added process datasource here and implement if needed
type ProcessDatasource interface {
	process_config.ConfigDatasource
	process_bet.BetDatasouce
	process_odds.OddsDatasource
	process_ticket_state.TicketStateDatasource
	//internal
	CreateResponse(processType ProcessType, data response.ResponseData) response.Response
	//overriding callbacks
	SetGetUserCallback(callback GetUserCallback)
	SetGetEventCallback(callback GetEventCallback)
	SetGetEventResultsCallback(callback GetEventResultsCallback)
	SetGetGameTableCallback(callback GetGameTableCallback)
	SetGetMemberTableCallback(callback GetMemberTableCallback)
	SetGetMemberConfigsCallback(callback GetMemberConfigsCallback)
	SetGetBetChipsCallback(callback GetBetChipsCallback)
	SetGetTicketStateCallback(callback GetTicketStateCallback)
}

// main datasource - contain default source for minigame process
type processDatasource struct {
	datasource Datasource
	//internal
	user          *models.User
	event         *models.Event
	eventResults  *[]models.EventResult
	gameTable     *models.GameTable
	memberTable   *models.MemberTable
	memberConfigs *[]models.MemberConfig
	betChipset    *[]float64
	tickets       *[]models.Ticket
	//overriding getter
	getUserCallback          *GetUserCallback
	getEventCallback         *GetEventCallback
	getEventResultCallback   *GetEventResultsCallback
	getGameTableCallback     *GetGameTableCallback
	getMemberTableCallback   *GetMemberTableCallback
	getMemberConfigsCallback *GetMemberConfigsCallback
	getBetChipsCallback      *GetBetChipsCallback
	getTicketStateCallback   *GetTicketStateCallback
}

func NewProcessDatasource(datasource Datasource) ProcessDatasource {
	return &processDatasource{datasource: datasource}
}

func (pd *processDatasource) GetIdentifier() string {
	return pd.datasource.GetIdentifier()
}

func (pd *processDatasource) GetUser() *models.User {
	if pd.getUserCallback != nil {
		return (*pd.getUserCallback)()
	}
	return pd.datasource.GetUser()
}

func (pd *processDatasource) GetEvent() *models.Event {
	if pd.getEventCallback != nil {
		return (*pd.getEventCallback)()
	}
	if pd.event == nil {
		event := models.Event{}
		rawQuery := fmt.Sprintf(`
		SELECT
			*
		FROM 
			mini_game_event
		WHERE 
			mini_game_table_id = %v AND
			start_datetime <= NOW()
		ORDER BY ctime DESC
		LIMIT 1
		`, pd.datasource.GetTableID())

		if err := db.Shared().Preload("Results", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("result_type ASC")
		}).Raw(rawQuery).First(&event).Error; err != nil {
			logger.Error(pd.GetIdentifier(), " process GetEvent error: ", err.Error())
			return nil
		}
		pd.event = &event
	}
	return pd.event
}

func (pd *processDatasource) GetEventResults() *[]models.EventResult {
	if pd.getEventResultCallback != nil {
		return (*pd.getEventResultCallback)()
	}
	if event := pd.GetEvent(); event != nil && pd.eventResults == nil {
		pd.eventResults = event.Results
	}
	return pd.eventResults
}

func (pd *processDatasource) GetGameTable() *models.GameTable {
	if pd.getGameTableCallback != nil {
		return (*pd.getGameTableCallback)()
	}
	if pd.gameTable == nil {
		gameTable := models.GameTable{
			ID: pd.datasource.GetTableID(),
		}

		if err := db.Shared().Where(gameTable).Order("ctime DESC").First(&gameTable).Error; err != nil {
			logger.Error(pd.GetIdentifier(), " process GetGameTable error: ", err.Error())
			return nil
		}
		pd.gameTable = &gameTable
	}
	return pd.gameTable
}

func (pd *processDatasource) GetMemberTable() *models.MemberTable {
	if pd.getMemberTableCallback != nil {
		return (*pd.getMemberTableCallback)()
	}
	if pd.memberTable == nil {
		memberTable := models.MemberTable{
			UserID:  pd.GetUser().ID,
			TableID: pd.datasource.GetTableID(),
		}

		if err := db.Shared().Where(memberTable).Order("ctime DESC").First(&memberTable).Error; err != nil {
			logger.Error(pd.GetIdentifier(), " process GetMemberTable error: ", err.Error())
			return nil
		}
		pd.memberTable = &memberTable
	}
	return pd.memberTable
}

func (pd *processDatasource) GetMemberConfigs() *[]models.MemberConfig {
	if pd.getMemberConfigsCallback != nil {
		return (*pd.getMemberConfigsCallback)()
	}
	if pd.memberConfigs == nil {
		memberConfigs := []models.MemberConfig{}

		if err := db.Shared().Where("game_id = ? AND user_id = ?", pd.datasource.GetGameID(), pd.GetUser().ID).
			Find(&memberConfigs).Error; err != nil {
			logger.Error(pd.GetIdentifier(), " process GetMemberConfigs error: ", err.Error())
			return nil
		}
		pd.memberConfigs = &memberConfigs
	}
	return pd.memberConfigs
}

func (pd *processDatasource) GetBetChips() *[]float64 {
	if pd.getBetChipsCallback != nil {
		return (*pd.getBetChipsCallback)()
	}
	if pd.betChipset == nil {
		user := pd.GetUser()
		betChipset := models.BetChipset{
			Currency:      user.CurrencyCode,
			CurrencyRatio: user.CurrencyRatio,
			TableID:       pd.datasource.GetTableID(),
		}

		if err := db.Shared().Where(betChipset).Order("ctime DESC").First(&betChipset).Error; err != nil {
			betChipset.CurrencyRatio = 0
			if err := db.Shared().Where(betChipset).Order("ctime DESC").First(&betChipset).Error; err != nil {
				betChipset.TableID = 0
				betChipset.Default = true
				if err := db.Shared().Where(betChipset).Order("ctime DESC").First(&betChipset).Error; err != nil {
					betChipset.Default = false
					if err := db.Shared().Where(betChipset).Order("ctime DESC").First(&betChipset).Error; err != nil {
						logger.Error(pd.GetIdentifier(), " process GetBetChips error: ", err.Error())
						return nil
					}
				}
			}
		}
		pd.betChipset = betChipset.GetBetChips()
	}

	return pd.betChipset
}

func (pd *processDatasource) GetTickets() *[]models.Ticket {
	if pd.getTicketStateCallback != nil {
		return (*pd.getTicketStateCallback)()
	}
	if event := pd.GetEvent(); event != nil && pd.tickets == nil {
		tickets := []models.Ticket{}
		rawQuery := fmt.Sprintf(`
		SELECT
			* 
		FROM
			mini_game_ticket 
		WHERE
			user_id = %v 
			AND mini_game_table_id = %v
		AND event_id = %v
		`, pd.GetUser().ID, pd.datasource.GetTableID(), *event.ID)

		if err := db.Shared().Preload("ComboTickets", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("ctime ASC")
		}).Raw(rawQuery).Find(&tickets).Error; err != nil {
			logger.Error(pd.GetIdentifier(), " process GetTickets error: ", err.Error())
			return nil
		}
		pd.tickets = &tickets
	}
	return pd.tickets
}

func (pd *processDatasource) CreateResponse(processType string, data response.ResponseData) response.Response {
	defer pd.cleanUp(processType)
	if err, ok := data.(*response.ErrorData); ok { //handle error override response type for error response
		return CreateResponse(ErrorType, err)
	}
	return CreateResponse(processType, data)
}

func (pd *processDatasource) cleanUp(processType string) {
	logger.Info("=================> (", processType, ") end <==================")
	//implement return object will be save internally and nil on cleanup
	//when first get on object it will query then if not nil it will just return the internal value
	pd.user = nil
	pd.event = nil
	pd.eventResults = nil
	pd.gameTable = nil
	pd.memberTable = nil
	pd.memberConfigs = nil
	pd.betChipset = nil
	pd.tickets = nil
}

// Setter
func (pd *processDatasource) SetGetUserCallback(callback GetUserCallback) {
	pd.getUserCallback = &callback
}

func (pd *processDatasource) SetGetEventCallback(callback GetEventCallback) {
	pd.getEventCallback = &callback
}

func (pd *processDatasource) SetGetEventResultsCallback(callback GetEventResultsCallback) {
	pd.getEventResultCallback = &callback
}

func (pd *processDatasource) SetGetGameTableCallback(callback GetGameTableCallback) {
	pd.getGameTableCallback = &callback
}

func (pd *processDatasource) SetGetMemberTableCallback(callback GetMemberTableCallback) {
	pd.getMemberTableCallback = &callback
}

func (pd *processDatasource) SetGetMemberConfigsCallback(callback GetMemberConfigsCallback) {
	pd.getMemberConfigsCallback = &callback
}

func (pd *processDatasource) SetGetBetChipsCallback(callback GetBetChipsCallback) {
	pd.getBetChipsCallback = &callback
}

func (pd *processDatasource) SetGetTicketStateCallback(callback GetTicketStateCallback) {
	pd.getTicketStateCallback = &callback
}

// static methods
func CreateResponse(responseType string, data response.ResponseData) response.Response {
	return response.Response{
		Type: responseType,
		Data: data,
	}
}
