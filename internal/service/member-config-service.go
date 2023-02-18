package service

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm/clause"
)

type memberConfigService struct{}

type MemberConfigService interface {
	GetMemberConfig(gameID int64, userID int64) ([]models.MemberConfig, error)
	BatchUpsertMemberConfig(configs []models.MemberConfig) ([]models.MemberConfig, error)
}

func NewMemberConfig() MemberConfigService {
	return &memberConfigService{}
}

func (s *memberConfigService) GetMemberConfig(gameID int64, userID int64) ([]models.MemberConfig, error) {
	var configs []models.MemberConfig
	result := DB.Where(&models.MemberConfig{GameId: gameID, UserID: userID}).Find(&configs)

	return configs, result.Error
}

func (s *memberConfigService) BatchUpsertMemberConfig(configs []models.MemberConfig) ([]models.MemberConfig, error) {
	var res []models.MemberConfig

	DB.Table("mini_game_memberconfig").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "game_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "ctime"}),
	}).Create(&configs)

	return res, nil
}
