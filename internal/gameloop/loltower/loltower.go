package gameloop_loltower

import (
	"errors"
	"fmt"
	"strings"
	"time"

	betsimulator_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/betsimulator/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	gamemanager_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	wsclient_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/measure"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/mutex"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type lolTowerGameloop struct {
	gl                    gameloop.Gameloop
	gm                    gamemanager.GameManager
	pb                    wsclient.MessagePublishBroker
	mg                    gameloop.MessageGenerator
	bs                    *betsimulator_loltower.BetSimulatorLOLTower
	queuedChampionsData   mutex.Data[types.Array[response.User]]
	isPublishingChampions bool
}

func NewGameLoop() gameloop.Gamelooper {
	gm := gamemanager_loltower.NewGameManager()    //load gamemanager
	gl := gameloop.NewGameLoop(gm.GetDatasource()) //reuse gamemanager datasource
	lolTowerGameloop := lolTowerGameloop{
		gm: gm,
		gl: gl,
		pb: wsclient_loltower.NewPublishBroker(),
		mg: gameloop.NewMessageGenerator(constants_loltower.Identifier),
		bs: betsimulator_loltower.NewBetSimulator(),
	}

	lolTowerGameloop.Initialize()
	return &lolTowerGameloop
}

func (ltgl *lolTowerGameloop) Start() {
	ltgl.gl.Start()
}

func (ltgl *lolTowerGameloop) Stop() {
	ltgl.gl.Stop()
}

func (ltgl *lolTowerGameloop) GetCurrentEvent() *models.Event {
	return ltgl.gl.GetCurrentEvent()
}

func (ltgl *lolTowerGameloop) Initialize() {
	ltgl.gl.SetPrepareCallback(func(elapsedTime gameloop.Milliseconds) {
		if err := ltgl.gm.CreateFutureHashes(); err != nil {
			logger.Error(ltgl.gl.GetIdentifier(), " CreateFutureHashes error: ", err.Error())
		}
		if err := ltgl.gm.CreateFutureEvents(); err != nil {
			logger.Error(ltgl.gl.GetIdentifier(), " CreateFutureEvents error: ", err.Error())
		}
	})
	ltgl.gl.CreatePhase(gameloop.NewGamePhase(
		"BETTING",
		0,
		constants_loltower.StartBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_loltower.StartBetMS)
			remainingTime := endMS - elapsedTime
			event := ltgl.gl.GetCurrentEvent()

			if remainingTime := endMS - ltgl.gl.ElapsedTime(); remainingTime > 0 {
				ltgl.pb.Publish(ltgl.mg.GenerateStateMessage("BETTING", utils.Ptr(ltgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			go ltgl.bs.StartBetting()
			if event != nil {
				logger.Debug(ltgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				if err := ltgl.gl.UpdateEventStatus(db.Shared(), *event.ID, constants.EVENT_STATUS_ACTIVE); err != nil {
					logger.Error(ltgl.gl.GetIdentifier(), " UpdateEventStatus active(", constants.EVENT_STATUS_ACTIVE, ") error: ", err.Error())
				}
			} else {
				logger.Debug(ltgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(ltgl.gl.GetIdentifier(), " BETTING total execution time: ", ltgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-ltgl.gl.ElapsedTimeOffset(elapsedTime))
		}))
	ltgl.gl.CreatePhase(gameloop.NewGamePhase(
		"STOP_BETTING",
		constants_loltower.StartBetMS,
		constants_loltower.StartBetMS+constants_loltower.StopBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_loltower.StartBetMS + constants_loltower.StopBetMS)
			remainingTime := endMS - elapsedTime
			event := ltgl.gl.GetCurrentEvent()

			if remainingTime := endMS - ltgl.gl.ElapsedTime(); remainingTime > 0 {
				ltgl.pb.Publish(ltgl.mg.GenerateStateMessage("STOP_BETTING", utils.Ptr(ltgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			go ltgl.bs.StartResulting()
			if event != nil {
				logger.Debug(ltgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Debug(ltgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(ltgl.gl.GetIdentifier(), " STOP_BETTING total execution time: ", ltgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-ltgl.gl.ElapsedTimeOffset(elapsedTime))
		}))
	ltgl.gl.CreatePhase(gameloop.NewGamePhase(
		"SHOW_RESULT",
		constants_loltower.StartBetMS+constants_loltower.StopBetMS,
		constants_loltower.StartBetMS+constants_loltower.StopBetMS+constants_loltower.ShowResultMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_loltower.StartBetMS + constants_loltower.StopBetMS + constants_loltower.ShowResultMS)
			remainingTime := endMS - elapsedTime
			event := ltgl.gl.GetCurrentEvent()

			if remainingTime := endMS - ltgl.gl.ElapsedTime(); remainingTime > 0 {
				ltgl.pb.Publish(ltgl.mg.GenerateStateMessage("SHOW_RESULT", utils.Ptr(ltgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
				utils.PerformAfter(50*time.Millisecond, func() {
					ltgl.PublishResults(event)
				})
			}
			if event != nil {
				logger.Debug(ltgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				go func() {
					exec := measure.NewExecution()
					defer func() {
						logger.Debug(ltgl.gl.GetIdentifier(), " SettleEventsBefore/ProcessDailyMaxWinnings execution time: ", exec.Done())
					}()
					if err := ltgl.gl.SettleEventsBefore(db.Shared(), utils.TimeNow(), func(tx *gorm.DB, event *models.Event) error {
						return ltgl.SettleTicketsWithEvent(tx, event)
					}); err != nil {
						logger.Error(ltgl.gl.GetIdentifier(), " SettleEventsBefore error: ", err.Error())
					}
					if err := ltgl.gl.ProcessDailyMaxWinnings(db.Shared()); err != nil {
						logger.Error(ltgl.gl.GetIdentifier(), " ProcessDailyMaxWinnings error: ", err.Error())
					}
					logger.Debug(ltgl.gl.GetIdentifier(), " TotalWinLoss: ", ltgl.gl.GetTotalWinLoss())
				}()
				go func() {
					exec := measure.NewExecution()
					defer func() {
						logger.Debug(ltgl.gl.GetIdentifier(), " betSimulator.UpdateResults/PublishLeaderboard/PublishChampions execution time: ", exec.Done())
					}()
					ltgl.bs.UpdateResults()
					ltgl.PublishLeaderboard()
					ltgl.PublishChampions()
				}()
			} else {
				logger.Debug(ltgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(ltgl.gl.GetIdentifier(), " SHOW_RESULT total execution time: ", ltgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-ltgl.gl.ElapsedTimeOffset(elapsedTime))
		}))
}

func (ltgl *lolTowerGameloop) SettleTicketsWithEvent(tx *gorm.DB, event *models.Event) error {
	//settleLossTickets
	//settleUnattendedTickets
	//settleMaxLevelTickets
	//settleMaxPayoutTickets
	//settleOldUnsettledTickets
	var tickets []models.Ticket
	rawQuery := fmt.Sprintf(`
	SELECT
		mgt.*
	FROM
		mini_game_ticket mgt
	LEFT JOIN mini_game_combo_ticket mgct ON
		mgct.ticket_id = mgt.id
	WHERE 
		mgt.mini_game_table_id = %v
		AND mgt.status = %v
	GROUP BY
		mgt.id
	HAVING
		(mgt."result" = %v
			OR mgt.event_id != %v
			OR count(mgct.*) >= %v OR count(mgct.*) = 0)
	ORDER BY
		mgt.ctime DESC
	`, ltgl.gl.GetDatasource().GetTableID(), constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		constants.TICKET_RESULT_LOSS,
		*ltgl.GetCurrentEvent().ID,
		constants_loltower.MaxBetCount)

	if err := tx.Raw(rawQuery).Find(&tickets).Error; err != nil {
		return err
	}
	ticketIDs := types.Array[models.Ticket](tickets).Map(func(value models.Ticket) any { return value.ID })

	if err := tx.Exec(`
	UPDATE
		mini_game_ticket
	SET
		status = ?,
		mtime = NOW(),
		status_mtime = NOW(),
		local_data_version = local_data_version + 1
	WHERE
		id IN (?)
	`, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT, ticketIDs).Error; err != nil {
		return errors.New("settle tickets: " + err.Error())
	}
	if err := tx.Exec(`
	UPDATE
   		mini_game_combo_ticket
	SET
		status = ?,
		mtime = NOW(),
		status_mtime = NOW(),
		local_data_version = local_data_version + 1
	WHERE
		ticket_id IN (?)
	`, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT, ticketIDs).Error; err != nil {
		return errors.New("settle combo tickets: " + err.Error())
	}
	return nil
}

func (ltgl *lolTowerGameloop) PublishResults(event *models.Event) {
	if event == nil {
		logger.Error(ltgl.gl.GetIdentifier(), " PublishResults event is nil")
		return
	}
	if event.Results == nil || len((*event.Results)) == 0 {
		logger.Error(ltgl.gl.GetIdentifier(), " PublishResults event.Results is empty")
		return
	}
	results := []string{}

	for _, v := range []string{"1", "2", "3", "4", "5"} {
		if !strings.Contains((*event.Results)[0].Value, v) {
			results = append(results, v)
		}
	}
	ltgl.pb.Publish(ltgl.mg.GenerateMessage("result", &response.LOLTowerResultData{
		Bomb: utils.Ptr(strings.Join(results, ",")),
		Gts:  utils.GenerateUnixTS(),
	}))
}

func (ltgl *lolTowerGameloop) PublishLeaderboard() {
	levels := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	gameDurationMS := constants_loltower.StartBetMS + constants_loltower.StopBetMS + constants_loltower.ShowResultMS
	offsetMS := gameDurationMS + constants_loltower.StartBetMS + constants_loltower.StopBetMS
	offset2MS := constants_loltower.StartBetMS + constants_loltower.StopBetMS
	leaderboards, err := GetLeaderboards(levels, offsetMS, offset2MS, ltgl.gl.GetDatasource().GetTableID())

	if err != nil {
		logger.Error(ltgl.gl.GetIdentifier(), " GetLeaderboards error: ", err.Error())
	}
	level := ""
	users := []response.User{}

	for _, leaderBoard := range leaderboards {
		level = leaderBoard.Level
		users = append(users, response.User{
			ID:   leaderBoard.GetEncryptedUserID(),
			Name: leaderBoard.GetEncryptedName(),
		})
	}
	leaderboardLevel := int8(types.String(level).Int())
	bsLeaderboardLevel, bsLeaderboard := ltgl.bs.GetLeaderBoard()

	if bsLeaderboardLevel > 0 {
		if bsLeaderboardLevel > leaderboardLevel {
			users = []response.User{} //override users
			for _, bsl := range bsLeaderboard {
				users = append(users, response.User{
					ID:   bsl.EncryptedID(),
					Name: bsl.EncryptedName(),
				})
			}
			level = string(types.Int(bsLeaderboardLevel).String())
		} else if bsLeaderboardLevel == leaderboardLevel {
			for _, bsl := range bsLeaderboard {
				users = append(users, response.User{
					ID:   bsl.EncryptedID(),
					Name: bsl.EncryptedName(),
				})
			}
		}
	}
	leaderboard := response.Leaderboard{Level: level, Users: users}

	ltgl.pb.Publish(ltgl.mg.GenerateMessage("member_list", &leaderboard))
	redis.SetLeaderboard(constants_loltower.Identifier, &leaderboard)
}

func (ltgl *lolTowerGameloop) PublishChampions() {
	offsetMS := constants_loltower.StartBetMS + constants_loltower.StopBetMS + constants_loltower.ShowResultMS
	offset2MS := constants_loltower.StartBetMS + constants_loltower.StopBetMS
	leaderboards, err := GetLeaderboards([]string{"10"}, offsetMS, offset2MS, ltgl.gl.GetDatasource().GetTableID())

	if err != nil {
		logger.Error(ltgl.gl.GetIdentifier(), " GetLeaderboards error: ", err.Error())
	}
	users := []response.User{}

	for _, leaderBoard := range leaderboards {
		users = append(users, response.User{
			ID:   leaderBoard.GetEncryptedUserID(),
			Name: leaderBoard.GetEncryptedName(),
		})
	}
	bsChampions := ltgl.bs.GetChampions()

	for _, bsc := range bsChampions {
		users = append(users, response.User{
			ID:   bsc.EncryptedID(),
			Name: bsc.EncryptedName(),
		})
	}
	if len(users) > 0 {
		ltgl.queuedChampionsData.Data = append(ltgl.queuedChampionsData.Data, users...)
		if !ltgl.isPublishingChampions {
			go ltgl.PublishChampion()
		}
	}
}

func (ltgl *lolTowerGameloop) PublishChampion() {
	ltgl.queuedChampionsData.Lock()
	defer ltgl.queuedChampionsData.Unlock()
	if len(ltgl.queuedChampionsData.Data) == 0 {
		ltgl.isPublishingChampions = false
		return
	}
	ltgl.isPublishingChampions = true
	ltgl.pb.Publish(ltgl.mg.GenerateMessage("champion", &ltgl.queuedChampionsData.Data[0]))
	ltgl.queuedChampionsData.Data.PopIndex(0)
	if len(ltgl.queuedChampionsData.Data) > 0 {
		ctx, _ := utils.TerminateContext()

		utils.Sleep(constants_loltower.ChampionIntervalMS*time.Millisecond, ctx)
	}
	go ltgl.PublishChampion()
}

func GetLeaderboards(levels []string, offsetMS int, offset2MS int, tableID int64) ([]models.LeaderBoard, error) {
	leaderBoards := []models.LeaderBoard{}
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
			(mg_event.start_datetime > (now() - INTERVAL '%v milliseconds') AND mg_event.start_datetime + INTERVAL '%v milliseconds' < now())
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
	`, strings.Join(levels, ","), offsetMS, offset2MS, tableID, tableID)

	if err := db.Shared().Raw(rawQuery).Find(&leaderBoards).Error; err != nil {
		return leaderBoards, err
	}
	return leaderBoards, nil
}
