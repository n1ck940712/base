package validate

import "bitbucket.org/esportsph/minigame-backend-golang/internal/models"

type User struct {
	ProcessType string
	Data        *models.User
}

type Event struct {
	ProcessType string
	Data        *models.Event
}

type EventResult struct {
	ProcessType string
	Data        *models.EventResult
}

type GameTable struct {
	ProcessType string
	Data        *models.GameTable
}

type MemberTable struct {
	ProcessType string
	Data        *models.MemberTable
}

type MemberConfig struct {
	ProcessType string
	Data        *models.MemberConfig
}

type Ticket struct {
	ProcessType string
	Data        *models.Ticket
}

type ComboTicket struct {
	ProcessType string
	Data        *models.ComboTicket
}
