package process_ticket_state

import (
	"time"

	constants_fpc "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state"
)

type TicketStateProcess interface {
	process_ticket_state.TicketStateProcess
}

type ticketStateProcess struct {
	dataSource process_ticket_state.TicketStateDatasource
}

func NewTicketStateProcess(datasource process_ticket_state.TicketStateDatasource) TicketStateProcess {
	return &ticketStateProcess{
		datasource,
	}
}

func (ts *ticketStateProcess) GetTicketState() response.ResponseData {
	event := ts.dataSource.GetEvent()
	tickets := ts.dataSource.GetTickets()

	if event == nil {
		return response.ErrorWithMessage("event is nil", process_ticket_state.TicketType)
	}
	if tickets == nil {
		return response.ErrorWithMessage("tickets is nil", process_ticket_state.TicketType)
	}
	if len(*tickets) == 0 {
		return &response.TicketState{}
	}
	fpcTickets := response.FishPrawnCrabTickets{}
	isResulting := time.Now().After(event.StartDatetime.Add((constants_fpc.StartBetMS + constants_fpc.StopBetMS) * time.Millisecond))

	for i := 0; i < len(*tickets); i++ {
		fpcTicket := response.FishPrawnCrabTicket{
			MarketType: (*tickets)[i].MarketType,
			Amount:     (*tickets)[i].Amount,
			Selection:  (*tickets)[i].Selection,
		}

		if isResulting {
			fpcTicket.WinLossAmount = (*tickets)[i].WinLossAmount
		}
		fpcTickets = append(fpcTickets, fpcTicket)
	}

	return &response.TicketState{Tickets: &fpcTickets}
}
