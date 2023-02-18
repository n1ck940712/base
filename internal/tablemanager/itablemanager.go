package tablemanager

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
)

type ITableManager interface {
	// GetTableManager() tablemanager.ITableManager
	GetMemberMaxBetAmount(userID int64) constants.CURRENCY_TYPE
	GetMemberMaxPayoutAmount(userID int64) constants.CURRENCY_TYPE
	GetMemberMinBetAmount(userID int64) constants.CURRENCY_TYPE
	GetMemberDetails(userID int64) *models.MemberTable
	GetTableDetails() *models.GameTable
}
