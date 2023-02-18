package gamemanager_loltower

import "testing"

func TestLOLTowerCreateFutureEvents(t *testing.T) {
	gmLOLTower := NewGameManager()

	if err := gmLOLTower.CreateFutureEvents(); err != nil {
		t.Fatal(err.Error())
	}
}
