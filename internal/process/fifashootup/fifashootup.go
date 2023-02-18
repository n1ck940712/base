package process_fifashootup

import (
	"fmt"
	"sync"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/fifashootup"
	process_config "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-config"
	process_game_data "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-game-data/fifashootup"
	process_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/fifashootup"
	process_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-state"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/fifashootup"
)

type processFIFAShootup struct {
	mu                  sync.Mutex
	datasource          process.Datasource
	processDatasource   process.ProcessDatasource
	stateProcess        process_state.StateProcess
	configProcess       process_config.ConfigProcess
	oddsProcesss        process_odds.OddsProcess
	betProcess          process_bet.BetProcess
	ticketStateProccess process_ticket_state.TicketStateProcess
	gameDataProcess     process_game_data.GameDataProcess

	eventID *int64
}

func NewFIFAShootupProcess(datasource process.Datasource) process.ProcessRequest {
	processDatasource := process.NewProcessDatasource(datasource)
	processFIFAShootup := processFIFAShootup{
		datasource:          datasource,
		processDatasource:   processDatasource,
		stateProcess:        process_state.NewStateProcess(processDatasource),
		configProcess:       process_config.NewConfigProcess(processDatasource),
		oddsProcesss:        process_odds.NewOddsProcess(processDatasource),
		betProcess:          process_bet.NewBetProcess(processDatasource),
		ticketStateProccess: process_ticket_state.NewTicketStateProcess(processDatasource),
		gameDataProcess:     process_game_data.NewGameDataProcess(processDatasource),
	}

	processDatasource.SetGetEventCallback(processFIFAShootup.getEvent)
	return &processFIFAShootup
}

func (pfs *processFIFAShootup) ProcessRequest(jsonStr string) response.Response {
	pfs.mu.Lock()
	defer pfs.mu.Unlock()
	requestData := request.NewRequest()

	if err := requestData.ParseJSON(jsonStr); err != nil {
		logger.Info(pfs.datasource.GetIdentifier(), " ProcessRequest ParseJSON error: ", err.Error())

		return pfs.processDatasource.CreateResponse(process.ErrorType, response.ErrorBadRequest(""))
	}
	switch requestData.Type {
	case process.StateType:
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.stateProcess.GetState())
	case process.ConfigType:
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.configProcess.GetConfig())
	case process.BetType:
		if err := requestData.FIFAShootupValidateBet(); err != nil {
			return pfs.processDatasource.CreateResponse(process.ErrorType, response.ErrorWithMessage(err.Error(), requestData.Type))
		}
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.betProcess.Placebet(requestData.GetBetData()))
	case process.SelectionType:
		if err := requestData.FIFAShootupValidateSelection(); err != nil {
			return pfs.processDatasource.CreateResponse(process.ErrorType, response.ErrorWithMessage(err.Error(), requestData.Type))
		}
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.betProcess.PlaceSelection(requestData.GetSelectionData()))
	case process.OddsType:
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.oddsProcesss.GetOdds())
	case process.TicketType:
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.ticketStateProccess.GetTicketState())
	case process.GameDataType:
		return pfs.processDatasource.CreateResponse(requestData.Type, pfs.gameDataProcess.GetGameData())
	}
	return pfs.processDatasource.CreateResponse(process.ErrorType, response.ErrorInValidType(requestData.Type))
}

func (pfs *processFIFAShootup) SetEventID(eventID int64) {
	pfs.eventID = &eventID
}

func (pfs *processFIFAShootup) getEvent() *models.Event {
	event := models.Event{}
	rawQuery := fmt.Sprintf(`
		SELECT
			*
		FROM 
			mini_game_event
		WHERE 
			mini_game_table_id = %v AND
			start_datetime <= NOW() AND NOW() < (start_datetime + INTERVAL '%v milliseconds')
		ORDER BY ctime DESC
		LIMIT 1
		`, pfs.datasource.GetTableID(), constants_fifashootup.StartBetMS+constants_fifashootup.StopBetMS+constants_fifashootup.ShowResultMS)

	if err := db.Shared().Preload("Results").Raw(rawQuery).First(&event).Error; err != nil {
		logger.Error(pfs.datasource.GetIdentifier(), " process GetEvent error: ", err.Error())
		return nil
	}
	return &event
}
