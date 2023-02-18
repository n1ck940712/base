package process_ticket_state

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type TicketStateProcess interface {
	process_ticket_state.TicketStateProcess
}

type ticketStateProcess struct {
	dataSource process_ticket_state.TicketStateDatasource
}

func NewTicketStateProcess(datasource process_ticket_state.TicketStateDatasource) TicketStateProcess {
	return &ticketStateProcess{
		datasource,
	}
}

func (ts *ticketStateProcess) GetTicketState() response.ResponseData {
	event := ts.dataSource.GetEvent()
	user := ts.dataSource.GetUser()

	if event == nil {
		return response.ErrorWithMessage("event is nil", process_ticket_state.TicketType)
	}
	if user == nil {
		return response.ErrorWithMessage("user is nil", process_ticket_state.TicketType)
	}
	ticketState := response.LOLTowerTicketState{}
	gameDurationMS := constants_loltower.StartBetMS + constants_loltower.StopBetMS + constants_loltower.ShowResultMS
	betToResultMS := constants_loltower.StartBetMS + constants_loltower.StopBetMS
	betToNextBetMS := gameDurationMS + betToResultMS
	maxLevel := constants_loltower.MaxLevel - constants_loltower.MaxSkip
	rawQuery := fmt.Sprintf(`
	SELECT
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d milliseconds')) AND ((combo_ticket.amount * tower_level.next_level_odds) > mgt.max_payout_amount)) THEN null ELSE combo_ticket.amount END AS bet_amount,
		CASE WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d milliseconds')) AND ((combo_ticket.amount * tower_level.next_level_odds) > mgt.max_payout_amount)) THEN %d ELSE tower_level.skip END AS skip,
		CASE
		    WHEN ((now() > (mg_event.start_datetime + INTERVAL '%d milliseconds')) AND ((combo_ticket.amount * tower_level.next_level_odds) > mgt.max_payout_amount)) THEN 0
			--skip retain future level
			WHEN combo_ticket.selection = 's' THEN tower_level.level + 1 
			--on show result and ticket is win, future level
			WHEN (now() > (mg_event.start_datetime + INTERVAL '%d milliseconds') AND combo_ticket.result = 0) THEN tower_level.level + 1 
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
			((tower_level.level - tower_level.skip) = %d OR combo_ticket.result = 1) AND mg_event.start_datetime >= (now() - INTERVAL '%d milliseconds')
			OR
			--evaluate until next event betting duration
			(tower_level.level - tower_level.skip) != %d AND combo_ticket.result = 0 AND mg_event.start_datetime >= (now() - INTERVAL '%d milliseconds')
		)
	ORDER BY tower_level.ctime DESC
	LIMIT 1`, betToResultMS, betToResultMS, constants.LOL_TOWER_SKIP_COUNT, betToResultMS, betToResultMS, *event.ID, constants_loltower.TableID, user.ID, maxLevel, betToResultMS, maxLevel, betToNextBetMS)
	if err := db.Shared().Raw(rawQuery).First(&ticketState).Error; err != nil {
		return &response.LOLTowerTicketState{
			BetAmount: nil,
			Level:     0,
			Skip:      constants_loltower.MaxSkip,
			Selection: nil,
		}
	}
	if ticketState.BetAmount != nil {
		euroOddsLevels := map[int]types.Odds{
			1:  constants_loltower.Level1Odds,
			2:  constants_loltower.Level2Odds,
			3:  constants_loltower.Level3Odds,
			4:  constants_loltower.Level4Odds,
			5:  constants_loltower.Level5Odds,
			6:  constants_loltower.Level6Odds,
			7:  constants_loltower.Level7Odds,
			8:  constants_loltower.Level8Odds,
			9:  constants_loltower.Level9Odds,
			10: constants_loltower.Level10Odds,
		}
		memberTable := ts.dataSource.GetMemberTable()

		if memberTable.IsEnabled {
			for level := 1; level <= constants_loltower.MaxLevel; level++ {
				maxPossibleLevel := maxLevel + ticketState.Skip

				if level > int(maxPossibleLevel) {
					break
				}
				if *ticketState.BetAmount > memberTable.MaxPayoutAmount/(*euroOddsLevels[level].Ptr()) {
					break
				}
				ticketState.MaxPayoutLevel = utils.Ptr(int8(level))
			}
		}
	}
	return &ticketState
}
