package api

import (
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
)

func TestAPI(t *testing.T) {
	ticketResponse := make(map[string]map[string]int)
	err := NewWalletAPI().GetTicket("CCSBE11425865332MGFPLS8894", ticketResponse)
	logger.Info("ticketResponse", ticketResponse)
	if err != nil {
		t.Fatal("Error TestAPI", err.Error())
	}
}
