package gamemanager_lolcouple

import "testing"

func TestLOLCoupleCreateFutureEvents(t *testing.T) {
	gmLOLCouple := NewGameManager()

	if err := gmLOLCouple.CreateFutureEvents(); err != nil {
		t.Fatal(err.Error())
	}
}
