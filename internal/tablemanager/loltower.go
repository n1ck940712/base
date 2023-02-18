package tablemanager

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
)

var (
	iMemberService service.IMemberTable = service.NewMemberTable()
)

type LolTowerTableManager struct {
	tableID int64
}

func NewLolTowerTableManager(tableID int64) *LolTowerTableManager {
	return &LolTowerTableManager{
		tableID: tableID,
	}
}

func (t *LolTowerTableManager) GetMemberMaxBetAmount(userID int64) constants.CURRENCY_TYPE {
	res := t.getMemberTableDetails(userID)
	return res.MaxBetAmount
}

func (t *LolTowerTableManager) GetMemberMaxPayoutAmount(userID int64) constants.CURRENCY_TYPE {
	res := t.getMemberTableDetails(userID)
	return res.MaxPayoutAmount

}

func (t *LolTowerTableManager) GetMemberMinBetAmount(userID int64) constants.CURRENCY_TYPE {
	res := t.getMemberTableDetails(userID)
	return res.MinBetAmount
}

func (t *LolTowerTableManager) getMemberTableDetails(userID int64) *models.MemberTable {
	member := iMemberService.GetMemberTable(userID, t.tableID)
	return &member
}

func (t *LolTowerTableManager) GetMemberDetails(userID int64) *models.MemberTable {
	member := iMemberService.GetMemberTable(userID, t.tableID)
	return &member
}

func (t *LolTowerTableManager) GetTableDetails() *models.GameTable {
	tableDetails := service.NewGame().GetTable(t.tableID)
	return &tableDetails
}
