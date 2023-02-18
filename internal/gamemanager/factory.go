package gamemanager

func NewGameManager(tableID int64) IGameManager {
	switch tableID {
	case 11:
		return NewLolTowerGameManager(tableID)
	default:
		return nil
	}
}
