package process_ticket_state

import (
	"time"

	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
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
	lcTickets := response.LolCoupleTickets{}
	isResulting := time.Now().After(event.StartDatetime.Add((constants_lolcouple.StartBetMS + constants_lolcouple.StopBetMS) * time.Millisecond))

	for i := 0; i < len(*tickets); i++ {
		lcTicket := response.LolCoupleTicket{
			MarketType: (*tickets)[i].MarketType,
			Amount:     (*tickets)[i].Amount,
			Selection:  (*tickets)[i].Selection,
		}

		if isResulting {
			lcTicket.WinLossAmount = (*tickets)[i].WinLossAmount
		}
		lcTickets = append(lcTickets, lcTicket)
	}

	return &response.TicketState{Tickets: &lcTickets}
}
