package process_bet

import "testing"

func TestGenerateTicketID(t *testing.T) {
	ticketID := GenerateTicketID(55554)

	println("ticket ID: ", ticketID)
	println("char count: ", len(ticketID))
}
