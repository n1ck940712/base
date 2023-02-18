package process_fishprawncrab

import (
	"fmt"
	"sync"

	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/fishprawncrab"
	process_config "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-config"
	process_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/fishprawncrab"
	process_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-state"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/fishprawncrab"
)

type processFishPrawnCrab struct {
	mu                  sync.Mutex
	datasource          process.Datasource
	processDatasource   process.ProcessDatasource
	stateProcess        process_state.StateProcess
	configProcess       process_config.ConfigProcess
	oddsProcesss        process_odds.OddsProcess
	betProcess          process_bet.BetProcess
	ticketStateProccess process_ticket_state.TicketStateProcess

	eventID *int64
}

func NewFishPrawCrabProcess(datasource process.Datasource) process.ProcessRequest {
	processDatasource := process.NewProcessDatasource(datasource)
	processFishPrawnCrab := processFishPrawnCrab{
		datasource:          datasource,
		processDatasource:   processDatasource,
		stateProcess:        process_state.NewStateProcess(processDatasource),
		configProcess:       process_config.NewConfigProcess(processDatasource),
		oddsProcesss:        process_odds.NewOddsProcess(processDatasource),
		betProcess:          process_bet.NewBetProcess(processDatasource),
		ticketStateProccess: process_ticket_state.NewTicketStateProcess(processDatasource),
	}

	processDatasource.SetGetEventCallback(processFishPrawnCrab.getEvent)
	return &processFishPrawnCrab
}

func (pfpc *processFishPrawnCrab) ProcessRequest(jsonStr string) response.Response {
	pfpc.mu.Lock()
	defer pfpc.mu.Unlock()
	requestData := request.NewRequest()

	if err := requestData.ParseJSON(jsonStr); err != nil {
		logger.Info(pfpc.datasource.GetIdentifier(), " ProcessRequest ParseJSON error: ", err.Error())

		return pfpc.processDatasource.CreateResponse(process.ErrorType, response.ErrorBadRequest(""))
	}
	switch requestData.Type {
	case process.StateType:
		return pfpc.processDatasource.CreateResponse(requestData.Type, pfpc.stateProcess.GetState())
	case process.ConfigType:
		return pfpc.processDatasource.CreateResponse(requestData.Type, pfpc.configProcess.GetConfig())
	case process.OddsType:
		return pfpc.processDatasource.CreateResponse(requestData.Type, pfpc.oddsProcesss.GetOdds())
	case process.BetType:
		if err := requestData.FishPrawnCrabValidateBet(); err != nil {
			return pfpc.processDatasource.CreateResponse(process.ErrorType, response.ErrorWithMessage(err.Error(), requestData.Type))
		}
		return pfpc.processDatasource.CreateResponse(requestData.Type, pfpc.betProcess.Placebet(requestData.GetBetData()))
	case process.CurrentTicketsType, process.TicketType:
		return pfpc.processDatasource.CreateResponse(requestData.Type, pfpc.ticketStateProccess.GetTicketState())
	}
	return pfpc.processDatasource.CreateResponse(process.ErrorType, response.ErrorInValidType(requestData.Type))
}

func (pfpc *processFishPrawnCrab) SetEventID(eventID int64) {
	pfpc.eventID = &eventID
}

func (pfpc *processFishPrawnCrab) getEvent() *models.Event {
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
		`, pfpc.datasource.GetTableID(), constants_fishprawncrab.StartBetMS+constants_fishprawncrab.StopBetMS+constants_fishprawncrab.ShowResultMS)

	if err := db.Shared().Preload("Results").Raw(rawQuery).First(&event).Error; err != nil {
		logger.Error(pfpc.datasource.GetIdentifier(), " process GetEvent error: ", err.Error())
		return nil
	}
	return &event
}
