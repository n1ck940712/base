package process_lolcouple

import (
	"fmt"
	"sync"

	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/lolcouple"
	process_config "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-config"
	process_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/lolcouple"
	process_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-state"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/lolcouple"
)

type processLOLCouple struct {
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

func NewLOLCoupleProcess(datasource process.Datasource) process.ProcessRequest {
	processDatasource := process.NewProcessDatasource(datasource)
	processLOLCouple := processLOLCouple{
		datasource:          datasource,
		processDatasource:   processDatasource,
		stateProcess:        process_state.NewStateProcess(processDatasource),
		configProcess:       process_config.NewConfigProcess(processDatasource),
		oddsProcesss:        process_odds.NewOddsProcess(processDatasource),
		betProcess:          process_bet.NewBetProcess(processDatasource),
		ticketStateProccess: process_ticket_state.NewTicketStateProcess(processDatasource),
	}

	processDatasource.SetGetEventCallback(processLOLCouple.getEvent)
	return &processLOLCouple
}

func (plc *processLOLCouple) ProcessRequest(jsonStr string) response.Response {
	plc.mu.Lock()
	defer plc.mu.Unlock()
	requestData := request.NewRequest()

	if err := requestData.ParseJSON(jsonStr); err != nil {
		logger.Info(plc.datasource.GetIdentifier(), " ProcessRequest ParseJSON error: ", err.Error())

		return plc.processDatasource.CreateResponse(process.ErrorType, response.ErrorBadRequest(""))
	}
	switch requestData.Type {
	case process.StateType:
		return plc.processDatasource.CreateResponse(requestData.Type, plc.stateProcess.GetState())
	case process.ConfigType:
		return plc.processDatasource.CreateResponse(requestData.Type, plc.configProcess.GetConfig())
	case process.BetType:
		if err := requestData.LOLCoupleValidateBet(); err != nil {
			return plc.processDatasource.CreateResponse(process.ErrorType, response.ErrorWithMessage(err.Error(), requestData.Type))
		}
		return plc.processDatasource.CreateResponse(requestData.Type, plc.betProcess.Placebet(requestData.GetBetData()))
	case process.OddsType:
		return plc.processDatasource.CreateResponse(requestData.Type, plc.oddsProcesss.GetOdds())
	case process.CurrentTicketsType, process.TicketType:
		return plc.processDatasource.CreateResponse(requestData.Type, plc.ticketStateProccess.GetTicketState())
	}
	return plc.processDatasource.CreateResponse(process.ErrorType, response.ErrorInValidType(requestData.Type))
}

func (plc *processLOLCouple) SetEventID(eventID int64) {
	plc.eventID = &eventID
}

func (plc *processLOLCouple) getEvent() *models.Event {
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
		`, plc.datasource.GetTableID(), constants_lolcouple.StartBetMS+constants_lolcouple.StopBetMS+constants_lolcouple.ShowResultMaxMS)

	if err := db.Shared().Preload("Results").Raw(rawQuery).First(&event).Error; err != nil {
		logger.Error(plc.datasource.GetIdentifier(), " process GetEvent error: ", err.Error())
		return nil
	}
	return &event
}
