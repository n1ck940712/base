package process_loltower

import (
	"fmt"
	"sync"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/loltower"
	process_config "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-config"
	process_member_list "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-member-list"
	process_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/loltower"
	process_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-state"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/loltower"
)

type processLOLTower struct {
	mu                  sync.Mutex
	datasource          process.Datasource
	processDatasource   process.ProcessDatasource
	stateProcess        process_state.StateProcess
	configProcess       process_config.ConfigProcess
	oddsProcesss        process_odds.OddsProcess
	betProcess          process_bet.BetProcess
	ticketStateProccess process_ticket_state.TicketStateProcess
	memberListProcess   process_member_list.MemberListProcess

	eventID *int64
}

func NewLOLTowerProcess(datasource process.Datasource) process.ProcessRequest {
	processDatasource := process.NewProcessDatasource(datasource)
	processLOLTower := processLOLTower{
		datasource:          datasource,
		processDatasource:   processDatasource,
		stateProcess:        process_state.NewStateProcess(processDatasource),
		configProcess:       process_config.NewConfigProcess(processDatasource),
		oddsProcesss:        process_odds.NewOddsProcess(processDatasource),
		betProcess:          process_bet.NewBetProcess(processDatasource),
		ticketStateProccess: process_ticket_state.NewTicketStateProcess(processDatasource),
		memberListProcess:   process_member_list.NewMemberListProcess(processDatasource),
	}

	processDatasource.SetGetEventCallback(processLOLTower.getEvent)
	return &processLOLTower
}

func (plt *processLOLTower) ProcessRequest(jsonStr string) response.Response {
	plt.mu.Lock()
	defer plt.mu.Unlock()
	requestData := request.NewRequest()

	if err := requestData.ParseJSON(jsonStr); err != nil {
		logger.Info(plt.datasource.GetIdentifier(), " ProcessRequest ParseJSON error: ", err.Error())

		return plt.processDatasource.CreateResponse(process.ErrorType, response.ErrorBadRequest(""))
	}
	switch requestData.Type {
	case process.StateType:
		return plt.processDatasource.CreateResponse(requestData.Type, plt.stateProcess.GetState())
	case process.ConfigType:
		return plt.processDatasource.CreateResponse(requestData.Type, plt.configProcess.GetConfig())
	case process.BetType:
		if err := requestData.LOLTowerValidateBet(); err != nil {
			return plt.processDatasource.CreateResponse(process.ErrorType, response.ErrorWithMessage(err.Error(), requestData.Type))
		}
		return plt.processDatasource.CreateResponse(requestData.Type, plt.betProcess.Placebet(requestData.GetBetData()))
	case process.SelectionType:
		if err := requestData.LOLTowerValidateSelection(); err != nil {
			return plt.processDatasource.CreateResponse(process.ErrorType, response.ErrorWithMessage(err.Error(), requestData.Type))
		}
		return plt.processDatasource.CreateResponse(requestData.Type, plt.betProcess.PlaceSelection(requestData.GetSelectionData()))
	case process.OddsType:
		return plt.processDatasource.CreateResponse(requestData.Type, plt.oddsProcesss.GetOdds())
	case process.TicketType:
		return plt.processDatasource.CreateResponse(requestData.Type, plt.ticketStateProccess.GetTicketState())
	case process.MemberListType:
		return plt.processDatasource.CreateResponse(requestData.Type, plt.memberListProcess.GetMemberList())
	}
	return plt.processDatasource.CreateResponse(process.ErrorType, response.ErrorInValidType(requestData.Type))
}

func (plt *processLOLTower) SetEventID(eventID int64) {
	plt.eventID = &eventID
}

func (plt *processLOLTower) getEvent() *models.Event {
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
		`, plt.datasource.GetTableID(), constants_loltower.StartBetMS+constants_loltower.StopBetMS+constants_loltower.ShowResultMS)

	if err := db.Shared().Preload("Results").Raw(rawQuery).First(&event).Error; err != nil {
		logger.Error(plt.datasource.GetIdentifier(), " process GetEvent error: ", err.Error())
		return nil
	}
	return &event
}
