package process_ticket_state

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/fifashootup"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
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
	tickets := []models.Ticket{}
	event := ts.dataSource.GetEvent()
	user := ts.dataSource.GetUser()
	comboTickets := (*[]models.ComboTicket)(nil)

	if event == nil {
		return response.ErrorWithMessage("event is nil", "ticket")
	}
	if user == nil {
		return response.ErrorWithMessage("user is nil", "ticket")
	}

	rawQuery := fmt.Sprintf(`
	SELECT
		*
	FROM
		mini_game_ticket 
	WHERE
		user_id = %v 
		AND mini_game_table_id = %v
	AND (event_id = %v OR status = %v)
	`, user.ID, ts.dataSource.GetMemberTable().TableID, *event.ID, constants.TICKET_STATUS_PAYMENT_CONFIRMED)

	if err := db.Shared().Preload("ComboTickets", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("ctime ASC")
	}).Preload("ComboTickets.Results").Raw(rawQuery).Find(&tickets).Error; err != nil {
		logger.Error(ts.dataSource.GetIdentifier(), " process GetTickets error: ", err.Error())
	}
	isResulting := time.Now().After(event.StartDatetime.Add((constants_fifashootup.StartBetMS + constants_fifashootup.StopBetMS) * time.Millisecond))

	if len(tickets) == 0 || (*tickets[0].Result == constants.TICKET_RESULT_LOSS && isResulting) {
		return &response.TicketState{}
	} else {
		fsTickets := response.FifaShootupTickets{}
		comboTickets = (tickets)[0].ComboTickets
		comboLen := len(*comboTickets)

		fsTicket := response.FifaShootupTicket{
			Amount:       (*comboTickets)[0].Amount,
			PayoutAmount: nil,
			Selection:    nil,
		}

		for i := 0; i < comboLen; i++ {
			ticket := (*comboTickets)[i]

			if ticket.Results == nil {
				return response.ErrorWithMessage("Ticket results is nil", "ticket")
			}

			ball := (*string)(nil)
			if (*event.ID != ticket.EventID) || (isResulting && *event.ID == ticket.EventID) {
				possibleWinningsAmount := types.Float(*ticket.PossibleWinningsAmount).Fixed(2)
				fsTicket.PayoutAmount = possibleWinningsAmount.Ptr()
				ball = utils.Ptr(ticket.Selection + "," + *possibleWinningsAmount.String().Ptr())

				var selections []string
				if (*ticket.Results)[0].ResultType == constants_fifashootup.EventResultType1 {
					selections = strings.Split((*ticket.Results)[0].Value, ",")
				} else {
					selections = strings.Split((*ticket.Results)[1].Value, ",")
				}

				leftSelections := strings.Split(selections[0], "-")
				rightSelections := strings.Split(selections[1], "-")
				leftRange, middleRange, rightRange := process_bet.GenerateLMRRanges(leftSelections[0], rightSelections[0])

				switch ticket.Selection {
				case constants_fifashootup.Selection1:
					*ball += "," + *types.Int(len(leftRange)).String().Ptr()
				case constants_fifashootup.Selection2:
					*ball += "," + *types.Int(len(middleRange)).String().Ptr()
				case constants_fifashootup.Selection3:
					*ball += "," + *types.Int(len(rightRange)).String().Ptr()
				case constants_fifashootup.SelectionPayout:
					ball = nil
				}
			}

			if *event.ID == ticket.EventID {
				fsTicket.Selection = &ticket.Selection
			}

			// add balls detail
			switch i {
			case 0:
				fsTicket.Ball1 = ball
			case 1:
				fsTicket.Ball2 = ball
			case 2:
				fsTicket.Ball3 = ball
			case 3:
				fsTicket.Ball4 = ball
			case 4:
				fsTicket.Ball5 = ball
			}
		}

		fsTickets = append(fsTickets, fsTicket)

		return &response.TicketState{Tickets: fsTickets}
	}
}

func (ts *ticketStateProcess) GetEventResultValue() (value string) {
	eventResults := ts.dataSource.GetEventResults()

	for i := 0; i < len(*eventResults); i++ {
		if (*eventResults)[i].ResultType == constants_fifashootup.EventResultType1 {
			value = (*eventResults)[i].Value
		}
	}
	return value
}
