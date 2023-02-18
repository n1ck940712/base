package process_bet

import (
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type SkipSelectionResponse struct { //temp response for support v1
	UserID        string  `json:"user_id"`
	TicketID      string  `json:"ticket_id"`
	ComboTicketID string  `json:"combo_ticket_id"`
	Level         int8    `json:"level"`
	Skip          int8    `json:"skip"`
	NextLevelOdds float64 `json:"next_level_odds"`
	Selection     string  `json:"selection"`
}

func (SkipSelectionResponse) Description() string {
	return "support for v1 skip selection response"
}

func (bp *betProcess) CreateTowerMemberLevel(tx *gorm.DB) error {
	tickets := bp.processDatasource.GetTickets()
	comboTickets := bp.processDatasource.GetComboTickets()
	level := int8(1)
	skip := int8(constants_loltower.MaxSkip)

	if prevMemberLevel := bp.GetPrevMemberLevel(); prevMemberLevel != nil {
		if (*comboTickets)[0].Selection == constants_loltower.SelectionSkip {
			level = prevMemberLevel.Level
			skip = prevMemberLevel.Skip - 1
		} else {
			level = prevMemberLevel.Level + 1
			skip = prevMemberLevel.Skip
		}
	}
	if err := tx.Create(&models.LolTowerMemberLevel{
		TicketID:      (*tickets)[0].ID,
		ComboTicketID: (*comboTickets)[0].ID,
		UserID:        bp.datasource.GetUser().ID,
		Level:         level,
		Skip:          skip,
		NextLevelOdds: *OddsFromLevel(level + 1).Ptr(),
	}).Error; err != nil {
		logger.Error("CreateTowerMemberLevel error: ", err.Error())
		return err
	}
	return nil
}

func (bp *betProcess) GetPrevMemberLevel() *models.LolTowerMemberLevel {
	if tickets := bp.processDatasource.GetTickets(); tickets != nil && bp.prevMemberLevel == nil {
		memberLevel := models.LolTowerMemberLevel{
			UserID:   bp.datasource.GetUser().ID,
			TicketID: (*tickets)[0].ID,
		}

		if err := db.Shared().Where(memberLevel).Order("ctime DESC").First(&memberLevel).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				logger.Error("GetPrevMemberLevel error: ", err.Error())
			}
			return nil
		}
		bp.prevMemberLevel = &memberLevel
	}

	return bp.prevMemberLevel
}

func OddsFromLevel(level int8) types.Odds {
	switch level {
	case 1:
		return constants_loltower.Level1Odds
	case 2:
		return constants_loltower.Level2Odds
	case 3:
		return constants_loltower.Level3Odds
	case 4:
		return constants_loltower.Level4Odds
	case 5:
		return constants_loltower.Level5Odds
	case 6:
		return constants_loltower.Level6Odds
	case 7:
		return constants_loltower.Level7Odds
	case 8:
		return constants_loltower.Level8Odds
	case 9:
		return constants_loltower.Level9Odds
	case 10:
		return constants_loltower.Level10Odds
	default:
		return 0
	}
}
