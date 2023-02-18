package service

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm/clause"
)

type SelectionLineIDs struct {
	SelectionHeaderID int64 `json:"selection_header_id"`
	SelectionLineID   int64 `json:"selection_line_id"`
}

type SelectionService interface {
	GetSelection(gameID int64) models.SelectionHeader
	GetSelectionLineByGameIDAndName(gameID int64, Name string) *SelectionLineIDs
}

type selectionService struct {
	selection models.SelectionHeader
}

func NewSelection() SelectionService {
	return &selectionService{}
}

func (service *selectionService) GetSelection(gameID int64) models.SelectionHeader {

	result := DB.Table("mini_game_minigameselectionheader").
		Preload(clause.Associations).
		Where("game_id = ?", gameID).
		Where("status = ?", "active").
		First(&service.selection)

	if result.Error != nil {
		logger.Error("No Selection Header")
	}

	return service.selection
}

func (service *selectionService) GetSelectionLineByGameIDAndName(gameID int64, Name string) *SelectionLineIDs {
	var sl SelectionLineIDs
	result := DB.Raw(`
	SELECT
		sh.id as selection_header_id , 
		sl.id as selection_line_id
	FROM
		mini_game_minigameselectionline sl
		JOIN mini_game_minigameselectionheader sh ON sh.ID = sl.mini_game_selection_header_id 
	WHERE
		sh.game_id = ? 
		AND sl.name = ? 
		AND sh.status = 'active' `, gameID, Name).Scan(&sl)

	if result.Error != nil {
		logger.Error("No Selection Line")
	}

	return &sl
}
