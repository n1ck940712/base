package service

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm"
)

type MemberTableService struct {
	MemberTable *models.MemberTable
}

type IMemberTable interface {
	GetMemberTable(id int64, tableID int64) models.MemberTable
}

func NewMemberTable() *MemberTableService {
	return &MemberTableService{}
}

func (m *MemberTableService) GetMemberTable(esID int64, tableID int64) models.MemberTable {
	var member models.MemberTable
	result := DB.Table(`mini_game_memberminigametable`).
		Where("user_id = ?", esID).
		Where("mini_game_table_id = ?", tableID).
		Find(&member)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Debug("User Not found!")
	}

	return member
}
