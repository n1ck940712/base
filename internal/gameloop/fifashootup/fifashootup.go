package gameloop_fifashootup

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	gamemanager_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	wsclient_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/measure"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"

	"gorm.io/gorm"
)

type fifaShootupGameloop struct {
	gl gameloop.Gameloop
	gm gamemanager.GameManager
	pb wsclient.MessagePublishBroker
	mg gameloop.MessageGenerator
}

func NewGameLoop() gameloop.Gamelooper {
	gm := gamemanager_fifashootup.NewGameManager() //load gamemanager
	gl := gameloop.NewGameLoop(gm.GetDatasource()) //reuse gamemanager datasource
	fifaShootupGameloop := fifaShootupGameloop{
		gm: gm,
		gl: gl,
		pb: wsclient_fifashootup.NewPublishBroker(),
		mg: gameloop.NewMessageGenerator(constants_fifashootup.Identifier),
	}

	fifaShootupGameloop.Initialize()
	return &fifaShootupGameloop
}

func (fsgl *fifaShootupGameloop) Start() {
	fsgl.gl.Start()
}

func (fsgl *fifaShootupGameloop) Stop() {
	fsgl.gl.Stop()
}

func (fsgl *fifaShootupGameloop) GetCurrentEvent() *models.Event {
	return fsgl.gl.GetCurrentEvent()
}

func (fsgl *fifaShootupGameloop) Initialize() {
	fsgl.gl.SetPrepareCallback(func(elapsedTime gameloop.Milliseconds) {
		if err := fsgl.gm.CreateFutureHashes(); err != nil {
			logger.Error(fsgl.gl.GetIdentifier(), " CreateFutureHashes error: ", err.Error())
		}
		if err := fsgl.gm.CreateFutureEvents(); err != nil {
			logger.Error(fsgl.gl.GetIdentifier(), " CreateFutureEvents error: ", err.Error())
		}
	})
	fsgl.gl.CreatePhase(gameloop.NewGamePhase(
		"BETTING",
		0,
		constants_fifashootup.StartBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_fifashootup.StartBetMS)
			remainingTime := endMS - elapsedTime
			event := fsgl.gl.GetCurrentEvent()

			if remainingTime := endMS - fsgl.gl.ElapsedTime(); remainingTime > 0 {
				fsgl.pb.Publish(fsgl.mg.GenerateStateMessage("BETTING", utils.Ptr(fsgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
				utils.PerformAfter(50*time.Millisecond, func() {
					fsgl.PublishGameData(event)
				})
			}
			if event != nil {
				logger.Debug(fsgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				if err := fsgl.gl.UpdateEventStatus(db.Shared(), *event.ID, constants.EVENT_STATUS_ACTIVE); err != nil {
					logger.Error(fsgl.gl.GetIdentifier(), " UpdateEventStatus active(", constants.EVENT_STATUS_ACTIVE, ") error: ", err.Error())
				}
			} else {
				logger.Debug(fsgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
				if remainingTime := endMS - fsgl.gl.ElapsedTime(); remainingTime > 0 {
					fsgl.pb.Publish(fsgl.mg.GenerateStateMessage("BETTING", utils.Ptr(fsgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
				}
			}
			logger.Debug(fsgl.gl.GetIdentifier(), " BETTING total execution time: ", fsgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-fsgl.gl.ElapsedTimeOffset(elapsedTime))
		}))
	fsgl.gl.CreatePhase(gameloop.NewGamePhase(
		"STOP_BETTING",
		constants_fifashootup.StartBetMS,
		constants_fifashootup.StartBetMS+constants_fifashootup.StopBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_fifashootup.StartBetMS + constants_fifashootup.StopBetMS)
			remainingTime := endMS - elapsedTime
			event := fsgl.gl.GetCurrentEvent()

			if remainingTime := endMS - fsgl.gl.ElapsedTime(); remainingTime > 0 {
				fsgl.pb.Publish(fsgl.mg.GenerateStateMessage("STOP_BETTING", utils.Ptr(fsgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			if event != nil {
				logger.Debug(fsgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Debug(fsgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(fsgl.gl.GetIdentifier(), " STOP_BETTING total execution time: ", fsgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-fsgl.gl.ElapsedTimeOffset(elapsedTime))
		},
	))
	fsgl.gl.CreatePhase(gameloop.NewGamePhase(
		"SHOW_RESULT",
		constants_fifashootup.StartBetMS+constants_fifashootup.StopBetMS,
		constants_fifashootup.StartBetMS+constants_fifashootup.StopBetMS+constants_fifashootup.ShowResultMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_fifashootup.StartBetMS + constants_fifashootup.StopBetMS + constants_fifashootup.ShowResultMS)
			remainingTime := endMS - elapsedTime
			event := fsgl.gl.GetCurrentEvent()

			if remainingTime := endMS - fsgl.gl.ElapsedTime(); remainingTime > 0 {
				fsgl.pb.Publish(fsgl.mg.GenerateStateMessage("SHOW_RESULT", utils.Ptr(fsgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
				utils.PerformAfter(50*time.Millisecond, func() {
					fsgl.PublishResult(event)
				})
			}
			if event != nil {
				logger.Debug(fsgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				go func() {
					exec := measure.NewExecution()
					defer func() {
						logger.Debug(fsgl.gl.GetIdentifier(), " SettleEventsBefore/ProcessDailyMaxWinnings execution time: ", exec.Done())
					}()
					if err := fsgl.gl.SettleEventsBefore(db.Shared(), utils.TimeNow(), func(tx *gorm.DB, event *models.Event) error {
						return fsgl.SettleTicketsWithEvent(tx, event)
					}); err != nil {
						logger.Error(fsgl.gl.GetIdentifier(), " SettleEventsBefore error: ", err.Error())
					}
					if err := fsgl.gl.ProcessDailyMaxWinnings(db.Shared()); err != nil {
						logger.Error(fsgl.gl.GetIdentifier(), " ProcessDailyMaxWinnings error: ", err.Error())
					}
					logger.Debug(fsgl.gl.GetIdentifier(), " TotalWinLoss: ", fsgl.gl.GetTotalWinLoss())
				}()
			} else {
				logger.Debug(fsgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(fsgl.gl.GetIdentifier(), " SHOW_RESULT total execution time: ", fsgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-fsgl.gl.ElapsedTimeOffset(elapsedTime))
		},
	))
}

func (fsgl *fifaShootupGameloop) PublishGameData(event *models.Event) {
	if event == nil {
		logger.Error(fsgl.gl.GetIdentifier(), " PublishResults event is nil")
		return
	}
	if event.Results == nil || len((*event.Results)) == 0 {
		logger.Error(fsgl.gl.GetIdentifier(), " PublishResults event.Results is empty")
		return
	}
	for i := 0; i < len(*event.Results); i++ {
		if (*event.Results)[i].ResultType == constants_fifashootup.EventResultType1 {
			leftCard := (*response.FIFAShootupCard)(nil)
			rightCard := (*response.FIFAShootupCard)(nil)
			middleCard := (*response.FIFAShootupCard)(nil)
			selections := strings.Split((*event.Results)[i].Value, ",")
			leftSelections := strings.Split(selections[0], "-")
			rightSelections := strings.Split(selections[1], "-")
			leftRange, middleRange, rightRange := process_bet.GenerateLMRRanges(leftSelections[0], rightSelections[0])

			leftCard = &response.FIFAShootupCard{
				Card:  &response.FIFAShootupCardData{No: strings.ToUpper(leftSelections[0]), Type: leftSelections[1]},
				Odds:  *process_bet.GenerateOdds(leftRange).Ptr(),
				Range: leftRange,
			}
			rightCard = &response.FIFAShootupCard{
				Card:  &response.FIFAShootupCardData{No: strings.ToUpper(rightSelections[0]), Type: rightSelections[1]},
				Odds:  *process_bet.GenerateOdds(rightRange).Ptr(),
				Range: rightRange,
			}
			if len(middleRange) > 0 {
				middleCard = &response.FIFAShootupCard{
					Odds:  *process_bet.GenerateOdds(middleRange).Ptr(),
					Range: middleRange,
				}
			}
			gameData := response.FIFAShootupGameData{
				EventID:    *event.ID,
				LeftCard:   leftCard,
				RightCard:  rightCard,
				MiddleCard: middleCard,
			}

			fsgl.pb.Publish(fsgl.mg.GenerateMessage("game_data", &gameData))
			redis.SetFIFAShootupGameData(constants_fifashootup.Identifier, &gameData)
		}
	}
}

func (fsgl *fifaShootupGameloop) PublishResult(event *models.Event) {
	if event == nil {
		logger.Error(fsgl.gl.GetIdentifier(), " PublishResults event is nil")
		return
	}
	if event.Results == nil || len((*event.Results)) == 0 {
		logger.Error(fsgl.gl.GetIdentifier(), " PublishResults event.Results is empty")
		return
	}
	for i := 0; i < len(*event.Results); i++ {
		if (*event.Results)[i].ResultType == constants_fifashootup.EventResultType1 {
			selections := strings.Split((*event.Results)[i].Value, ",")
			leftSelections := strings.Split(selections[0], "-")
			rightSelections := strings.Split(selections[1], "-")
			resultSelections := strings.Split(selections[2], "-")

			resultData := response.FIFAShootupResultData{
				ResultCard: &response.FIFAShootupResultCardData{
					No:   strings.ToUpper(resultSelections[0]),
					Type: resultSelections[1],
				},
				Result: utils.Ptr(process_bet.GenerateResult(leftSelections[0], rightSelections[0], resultSelections[0])),
			}

			fsgl.pb.Publish(fsgl.mg.GenerateMessage("result", &resultData))
		}
	}
}

func (fsgl *fifaShootupGameloop) SettleTicketsWithEvent(tx *gorm.DB, event *models.Event) error {
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
	`, fsgl.gl.GetDatasource().GetTableID(), constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		constants.TICKET_RESULT_LOSS,
		*fsgl.GetCurrentEvent().ID,
		constants_fifashootup.MaxBetCount)

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
