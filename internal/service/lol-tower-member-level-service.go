package service

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/cache"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm"
)

type LolTowerMemberLevelService struct{}
type TicketState struct {
	BetAmount      *float64 `json:"bet_amount"`
	Level          int      `json:"level"`
	Skip           int      `json:"skip"`
	Selection      *string  `json:"selection"`
	MaxPayoutLevel *int8    `json:"max_payout_level"`
}
type ILolTowerMemberLevelService interface {
	GetMemberCurrentLevel(ticketID string) int
	CreateLOLTowerMemberLevel(models.LolTowerMemberLevel) models.LolTowerMemberLevel
	GetMemberTicketState(tableID int64, userID int64) (*TicketState, error)
	GetMemberTicketStateV2(tableID int64, userID int64) (*TicketState, error)
	GetMemberTicketStateV3(tableID int64, userID int64) (*TicketState, error)
	GetLeaderboards(levels []string, offset int32) ([]models.LeaderBoard, error)
	GetChampionMembers() ([]models.User, error)
}

func NewLolTowerMemberLevelService() *LolTowerMemberLevelService {
	return &LolTowerMemberLevelService{}
}

func (l *LolTowerMemberLevelService) GetMemberCurrentLevel(ticketID string) (int, error) {
	var level int
	result := DB.Raw(`
		SELECT level from 
		mini_game_lol_tower_member_level l 
		WHERE ticket_id = ? ORDER BY level DESC`, ticketID).Find(&level)

	if result.Error != nil {
		return level, result.Error
	}

	return level, nil
}

func (l *LolTowerMemberLevelService) CreateLOLTowerMemberLevel(data models.LolTowerMemberLevel) (models.LolTowerMemberLevel, error) {
	tm := time.Now()
	data.Ctime = tm
	data.Mtime = tm
	result := DB.Table("mini_game_lol_tower_member_level").Create(&data)

	if result.Error != nil {
		logger.Error("LOL Member Tower create failed")
		return data, result.Error
	}

	return data, nil
}

func (l *LolTowerMemberLevelService) GetMemberTicketState(tableID int64, userID int64) (*TicketState, error) {
	var res TicketState
	betToResultDuration := constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION
	betToNextBetDuration := constants.LOL_TOWER_GAME_DURATION + betToResultDuration
	event, err := cache.GetEvent(tableID)

	if err != nil {
		return nil, err
	}

	rawQuery := fmt.Sprintf(`
	SELECT
		combo_ticket.amount as bet_amount,
		tower_level.skip,
		CASE 
			--skip retain future level
			WHEN combo_ticket.selection = 's' THEN tower_level.level + 1 
			--on show result and ticket is win, future level
			WHEN (now() > (mg_event.start_datetime + INTERVAL '%d seconds') AND combo_ticket.result = 0) THEN tower_level.level + 1 
			ELSE tower_level.level
		END AS level,
		CASE
			WHEN combo_ticket.event_id != %d THEN null
			ELSE combo_ticket.selection
		END AS selection
	FROM 
		mini_game_lol_tower_member_level AS tower_level
		LEFT JOIN mini_game_combo_ticket AS combo_ticket ON combo_ticket.id = tower_level.combo_ticket_id
		LEFT JOIN mini_game_event AS mg_event ON mg_event.id = combo_ticket.event_id
		LEFT JOIN mini_game_memberminigametable mgt ON (mgt.user_id = combo_ticket.user_id AND mgt.is_enabled = TRUE AND mgt.mini_game_table_id = combo_ticket.mini_game_table_id)
	WHERE 
		combo_ticket.mini_game_table_id = %d
		AND tower_level.user_id = %d
		AND (
			--on last level only evaluate until show result
			((tower_level.level - tower_level.skip) = 7 OR combo_ticket.result = 1) AND mg_event.start_datetime >= (now() - INTERVAL '%d seconds')
			OR
			--evaluate until next event betting duration
			(tower_level.level - tower_level.skip) != 7 AND combo_ticket.result = 0 AND mg_event.start_datetime >= (now() - INTERVAL '%d seconds')
		)
		AND ( tower_level.next_level_odds * combo_ticket.amount ) < mgt.max_payout_amount
	ORDER BY tower_level.ctime DESC
	LIMIT 1`, betToResultDuration, event.ID, tableID, userID, betToResultDuration, betToNextBetDuration)
	logger.Info("LOG SQL: ", rawQuery)
	result := DB.Raw(rawQuery).First(&res)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Error("No combo ticket found")
		return nil, result.Error
	}

	return &res, result.Error
}

func (l *LolTowerMemberLevelService) GetMemberTicketStateV2(tableID int64, userID int64) (*TicketState, error) {
	var res TicketState
	betToResultDuration := constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION
	betToNextBetDuration := constants.LOL_TOWER_GAME_DURATION + betToResultDuration
	event, err := cache.GetEvent(tableID)

	if err != nil {
		return nil, err
	}

	rawQuery := fmt.Sprintf(`
	SELECT
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND (((combo_ticket.amount * tower_level.next_level_odds) - combo_ticket.amount) > mgt.max_payout_amount)) THEN null ELSE combo_ticket.amount END AS bet_amount,
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND (((combo_ticket.amount * tower_level.next_level_odds) - combo_ticket.amount) > mgt.max_payout_amount)) THEN 3 ELSE tower_level.skip END AS skip,
		CASE
		    WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND (((combo_ticket.amount * tower_level.next_level_odds) - combo_ticket.amount) > mgt.max_payout_amount)) THEN 0
			--skip retain future level
			WHEN combo_ticket.selection = 's' THEN tower_level.level + 1 
			--on show result and ticket is win, future level
			WHEN (now() > (mg_event.start_datetime + INTERVAL '%d seconds') AND combo_ticket.result = 0) THEN tower_level.level + 1 
			ELSE tower_level.level
		END AS level,
		CASE
			WHEN combo_ticket.event_id != %d THEN null
			ELSE combo_ticket.selection
		END AS selection,
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND (((combo_ticket.amount * tower_level.next_level_odds) - combo_ticket.amount) > mgt.max_payout_amount)) THEN true ELSE false END AS is_next_level_payout
	FROM 
		mini_game_lol_tower_member_level AS tower_level
		LEFT JOIN mini_game_combo_ticket AS combo_ticket ON combo_ticket.id = tower_level.combo_ticket_id
		LEFT JOIN mini_game_event AS mg_event ON mg_event.id = combo_ticket.event_id
		LEFT JOIN mini_game_memberminigametable mgt ON (mgt.user_id = combo_ticket.user_id AND mgt.is_enabled = TRUE AND mgt.mini_game_table_id = combo_ticket.mini_game_table_id)
	WHERE 
		combo_ticket.mini_game_table_id = %d
		AND tower_level.user_id = %d
		AND (
			--on last level only evaluate until show result
			((tower_level.level - tower_level.skip) = 7 OR combo_ticket.result = 1) AND mg_event.start_datetime >= (now() - INTERVAL '%d seconds')
			OR
			--evaluate until next event betting duration
			(tower_level.level - tower_level.skip) != 7 AND combo_ticket.result = 0 AND mg_event.start_datetime >= (now() - INTERVAL '%d seconds')
		)
	ORDER BY tower_level.ctime DESC
	LIMIT 1`, betToResultDuration, betToResultDuration, betToResultDuration, betToResultDuration, event.ID, betToResultDuration, tableID, userID, betToResultDuration, betToNextBetDuration)
	result := DB.Raw(rawQuery).First(&res)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Error("No combo ticket found")
		return nil, result.Error
	}

	return &res, result.Error
}

func (l *LolTowerMemberLevelService) GetMemberTicketStateV3(tableID int64, userID int64) (*TicketState, error) {
	event, err := cache.GetEvent(tableID)

	if err != nil {
		return nil, err
	}

	var res TicketState
	betToResultDuration := constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION
	betToNextBetDuration := constants.LOL_TOWER_GAME_DURATION + betToResultDuration
	maxLevel := constants.LOL_TOWER_MAX_LEVEL - constants.LOL_TOWER_SKIP_COUNT
	rawQuery := fmt.Sprintf(`
	SELECT
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND ((combo_ticket.amount * tower_level.next_level_odds) > mgt.max_payout_amount)) THEN null ELSE combo_ticket.amount END AS bet_amount,
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND ((combo_ticket.amount * tower_level.next_level_odds) > mgt.max_payout_amount)) THEN %d ELSE tower_level.skip END AS skip,
		CASE
		    WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d seconds')) AND ((combo_ticket.amount * tower_level.next_level_odds) > mgt.max_payout_amount)) THEN 0
			--skip retain future level
			WHEN combo_ticket.selection = 's' THEN tower_level.level + 1 
			--on show result and ticket is win, future level
			WHEN (now() > (mg_event.start_datetime + INTERVAL '%d seconds') AND combo_ticket.result = 0) THEN tower_level.level + 1 
			ELSE tower_level.level
		END AS level,
		CASE
			WHEN combo_ticket.event_id != %d THEN null
			ELSE combo_ticket.selection
		END AS selection
	FROM 
		mini_game_lol_tower_member_level AS tower_level
		LEFT JOIN mini_game_combo_ticket AS combo_ticket ON combo_ticket.id = tower_level.combo_ticket_id
		LEFT JOIN mini_game_event AS mg_event ON mg_event.id = combo_ticket.event_id
		LEFT JOIN mini_game_memberminigametable mgt ON (mgt.user_id = combo_ticket.user_id AND mgt.is_enabled = TRUE AND mgt.mini_game_table_id = combo_ticket.mini_game_table_id)
	WHERE 
		combo_ticket.mini_game_table_id = %d
		AND tower_level.user_id = %d
		AND (
			--on last level only evaluate until show result
			((tower_level.level - tower_level.skip) = %d OR combo_ticket.result = 1) AND mg_event.start_datetime >= (now() - INTERVAL '%d seconds')
			OR
			--evaluate until next event betting duration
			(tower_level.level - tower_level.skip) != %d AND combo_ticket.result = 0 AND mg_event.start_datetime >= (now() - INTERVAL '%d seconds')
		)
	ORDER BY tower_level.ctime DESC
	LIMIT 1`, betToResultDuration, betToResultDuration, constants.LOL_TOWER_SKIP_COUNT, betToResultDuration, betToResultDuration, *event.ID, tableID, userID, maxLevel, betToResultDuration, maxLevel, betToNextBetDuration)
	result := DB.Raw(rawQuery).First(&res)

	if result.Error != nil {
		return nil, result.Error
	}

	return &res, result.Error
}

func (l *LolTowerMemberLevelService) GetLeaderboards(levels []string, offset int, offset2 int, tableID int64) ([]models.LeaderBoard, error) {
	leaderBoards := []models.LeaderBoard{}
	levelsJoin := strings.Join(levels, ",")
	rawQuery := fmt.Sprintf(`
	WITH mg_tower_max_level AS (
		SELECT 
			MAX(mg_tower_level.level) AS max_level,
			mg_event.id AS event_id
		FROM mini_game_lol_tower_member_level AS mg_tower_level
			LEFT JOIN mini_game_combo_ticket AS combo_ticket ON mg_tower_level.combo_ticket_id = combo_ticket.id
			LEFT JOIN mini_game_event AS mg_event ON mg_event.id = combo_ticket.event_id
		WHERE 
			mg_tower_level.level IN (%v) AND 
			combo_ticket.result = 0 AND --0 -> WIN
			(mg_event.start_datetime > (now() - INTERVAL '%d seconds') AND mg_event.start_datetime + INTERVAL '%d seconds' < now())
			AND combo_ticket.mini_game_table_id = %v
		GROUP BY mg_event.id
		ORDER BY mg_event.ctime DESC
		LIMIT 1
	)
				
	SELECT
		DISTINCT mg_user.id as user_id,
		mg_user.esports_id,
		mg_user.esports_partner_id,
		mg_user.member_code, 
		mg_tower_level.level
	FROM mini_game_lol_tower_member_level AS mg_tower_level
		LEFT JOIN mini_game_combo_ticket AS ticket ON ticket.id = mg_tower_level.combo_ticket_id                          
		LEFT JOIN mini_game_user AS mg_user ON  mg_user.id = mg_tower_level.user_id
		LEFT JOIN mini_game_event AS mg_event ON mg_event.id = ticket.event_id
		INNER JOIN mg_tower_max_level ON (mg_tower_max_level.max_level = mg_tower_level.level
		AND mg_event.id = mg_tower_max_level.event_id)
	WHERE 
		ticket.result = 0 AND --0 -> WIN
		mg_event.id = mg_tower_max_level.event_id
		and ticket.ticket_id = mg_tower_level.ticket_id
		AND ticket.mini_game_table_id = %v
	`, levelsJoin, offset, offset2, tableID, tableID)
	dbResult := DB.Raw(rawQuery).Find(&leaderBoards)

	return leaderBoards, dbResult.Error

	// 	SELECT
	// 	DISTINCT mg_user.id as user_id,
	// 	mg_user.member_code,
	// 	mg_tower_level.ctime,
	// 	mg_tower_level.level
	// FROM mini_game_lol_tower_member_level AS mg_tower_level
	// 	LEFT JOIN mini_game_ticket AS ticket ON (ticket.id = mg_tower_level.ticket_id)
	// 	LEFT JOIN mini_game_user AS mg_user ON  mg_user.id = mg_tower_level.user_id
	// 	LEFT JOIN mini_game_event AS mg_event ON mg_event.id = ticket.event_id
	// 	RIGHT JOIN mg_tower_max_level ON (mg_tower_max_level.max_level = mg_tower_level.level
	// 	AND mg_event.id = mg_tower_max_level.event_id)
	// WHERE
	// 	ticket.result = 0 AND --0 -> WIN
	// 	-- mg_event.id = mg_tower_max_level.event_id
	// 	mg_tower_level.event_id = mg_tower_max_level.event_id

}
