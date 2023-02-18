package proccess_ticket_state

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
)

const (
	TicketType         = "ticket"
	CurrentTicketsType = "current-tickets"
)

type TicketStateDatasource interface {
	GetIdentifier() string
	GetUser() *models.User
	GetEvent() *models.Event
	GetMemberTable() *models.MemberTable
	GetTickets() *[]models.Ticket
	GetEventResults() *[]models.EventResult
}

type TicketStateProcess interface {
	GetTicketState() response.ResponseData
}
