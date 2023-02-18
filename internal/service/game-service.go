package service

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm/clause"
)

type GameService interface {
	GetGames(gameID int64) (models.Game, error)
	GetTable(tableID int64) models.GameTable
	IsTableExist(tablID int64) bool
}

type gameService struct {
	game models.Game
}

func NewGame() GameService {
	return &gameService{}
}

func (service *gameService) GetGames(gameID int64) (models.Game, error) {
	result := DB.Table("mini_game_game").
		Preload("GameTables").
		Where("id = ?", gameID).
		Find(&service.game)

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return service.game, result.Error
}

func (service *gameService) GetTable(tableID int64) models.GameTable {
	var table models.GameTable

	result := DB.Table("mini_game_minigametable").
		Preload(clause.Associations).
		Where("id = ?", tableID).
		Find(&table)

	if result.Error != nil {
		logger.Error(result.Error)
	}

	return table
}

func (service *gameService) IsTableExist(tableID int64) bool {
	var table models.GameTable

	result := DB.Table("mini_game_minigametable").
		Preload(clause.Associations).
		Where("id = ?", tableID).
		Find(&table)

	if result.Error != nil {
		return false
	}

	return table.ID > 0
}
