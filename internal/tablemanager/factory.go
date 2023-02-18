package tablemanager

func NewTableManager(tableID int64) ITableManager {
	switch tableID {
	case 11:
		return NewLolTowerTableManager(tableID)
	default:
		return nil
	}
}
