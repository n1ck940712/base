package gameloop_fishprawncrab

import (
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	gamemanager_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	wsclient_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/measure"
	"gorm.io/gorm"
)

type fishPrawnCrabGameloop struct {
	gl gameloop.Gameloop
	gm gamemanager.GameManager
	pb wsclient.MessagePublishBroker
	mg gameloop.MessageGenerator
}

func NewGameLoop() gameloop.Gamelooper {
	gm := gamemanager_fishprawncrab.NewGameManager() //load gamemanager
	gl := gameloop.NewGameLoop(gm.GetDatasource())   //reuse gamemanager datasource
	fishPrawnCrabGameloop := fishPrawnCrabGameloop{
		gm: gm,
		gl: gl,
		pb: wsclient_fishprawncrab.NewPublishBroker(),
		mg: gameloop.NewMessageGenerator(constants_fishprawncrab.Identifier),
	}

	fishPrawnCrabGameloop.Initialize()
	return &fishPrawnCrabGameloop
}

func (fpcgl *fishPrawnCrabGameloop) Start() {
	fpcgl.gl.Start()
}

func (fpcgl *fishPrawnCrabGameloop) Stop() {
	fpcgl.gl.Stop()
}

func (fpcgl *fishPrawnCrabGameloop) GetCurrentEvent() *models.Event {
	return fpcgl.gl.GetCurrentEvent()
}

func (fpcgl *fishPrawnCrabGameloop) Initialize() {
	fpcgl.gl.SetPrepareCallback(func(elapsedTime gameloop.Milliseconds) {
		if err := fpcgl.gm.CreateFutureHashes(); err != nil {
			logger.Error(fpcgl.gl.GetIdentifier(), " CreateFutureHashes error: ", err.Error())
		}
		if err := fpcgl.gm.CreateFutureEvents(); err != nil {
			logger.Error(fpcgl.gl.GetIdentifier(), " CreateFutureEvents error: ", err.Error())
		}
	})
	fpcgl.gl.CreatePhase(gameloop.NewGamePhase(
		"BETTING",
		0,
		constants_fishprawncrab.StartBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_fishprawncrab.StartBetMS)
			remainingTime := endMS - elapsedTime
			event := fpcgl.gl.GetCurrentEvent()

			if remainingTime := endMS - fpcgl.gl.ElapsedTime(); remainingTime > 0 {
				fpcgl.pb.Publish(fpcgl.mg.GenerateStateMessage("BETTING", utils.Ptr(fpcgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			if event != nil {
				logger.Debug(fpcgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				if err := fpcgl.gl.UpdateEventStatus(db.Shared(), *event.ID, constants.EVENT_STATUS_ACTIVE); err != nil {
					logger.Error(fpcgl.gl.GetIdentifier(), " UpdateEventStatus active(", constants.EVENT_STATUS_ACTIVE, ") error: ", err.Error())
				}
			} else {
				logger.Debug(fpcgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(fpcgl.gl.GetIdentifier(), " BETTING total execution time: ", fpcgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-fpcgl.gl.ElapsedTimeOffset(elapsedTime))
		}))
	fpcgl.gl.CreatePhase(gameloop.NewGamePhase(
		"STOP_BETTING",
		constants_fishprawncrab.StartBetMS,
		constants_fishprawncrab.StartBetMS+constants_fishprawncrab.StopBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_fishprawncrab.StartBetMS + constants_fishprawncrab.StopBetMS)
			remainingTime := endMS - elapsedTime
			event := fpcgl.gl.GetCurrentEvent()

			if remainingTime := endMS - fpcgl.gl.ElapsedTime(); remainingTime > 0 {
				fpcgl.pb.Publish(fpcgl.mg.GenerateStateMessage("STOP_BETTING", utils.Ptr(fpcgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			if event != nil {
				logger.Debug(fpcgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Debug(fpcgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(fpcgl.gl.GetIdentifier(), " STOP_BETTING total execution time: ", fpcgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-fpcgl.gl.ElapsedTimeOffset(elapsedTime))
		},
	))
	fpcgl.gl.CreatePhase(gameloop.NewGamePhase(
		"SHOW_RESULT",
		constants_fishprawncrab.StartBetMS+constants_fishprawncrab.StopBetMS,
		constants_fishprawncrab.StartBetMS+constants_fishprawncrab.StopBetMS+constants_fishprawncrab.ShowResultMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_fishprawncrab.StartBetMS + constants_fishprawncrab.StopBetMS + constants_fishprawncrab.ShowResultMS)
			remainingTime := endMS - elapsedTime
			overflowMS := gameloop.Milliseconds(0)
			event := fpcgl.gl.GetCurrentEvent()

			if remainingTime := endMS - fpcgl.gl.ElapsedTime(); remainingTime > 0 {
				fpcgl.pb.Publish(fpcgl.mg.GenerateStateMessage("SHOW_RESULT", utils.Ptr(fpcgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
				utils.PerformAfter(50*time.Millisecond, func() {
					fpcgl.PublishResults(event, overflowMS)
				})
			}
			if event != nil {
				logger.Debug(fpcgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				go func() {
					exec := measure.NewExecution()
					defer func() {
						logger.Debug(fpcgl.gl.GetIdentifier(), " SettleEventsBefore/ProcessDailyMaxWinnings execution time: ", exec.Done())
					}()
					if err := fpcgl.gl.SettleEventsBefore(db.Shared(), utils.TimeNow(), func(tx *gorm.DB, event *models.Event) error {
						return fpcgl.gl.SettleTicketsForEvent(tx, *event.ID)
					}); err != nil {
						logger.Error(fpcgl.gl.GetIdentifier(), " SettleEventsBefore error: ", err.Error())
					}
					if err := fpcgl.gl.ProcessDailyMaxWinnings(db.Shared()); err != nil {
						logger.Error(fpcgl.gl.GetIdentifier(), " ProcessDailyMaxWinnings error: ", err.Error())
					}
					logger.Debug(fpcgl.gl.GetIdentifier(), " TotalWinLoss: ", fpcgl.gl.GetTotalWinLoss())
				}()
			} else {
				logger.Debug(fpcgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(fpcgl.gl.GetIdentifier(), " SHOW_RESULT total execution time: ", fpcgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-(overflowMS + fpcgl.gl.ElapsedTimeOffset(elapsedTime)))
		},
	))
}

func (fpcgl *fishPrawnCrabGameloop) PublishResults(event *models.Event, overflowMS gameloop.Milliseconds) {
	if event == nil {
		logger.Error(fpcgl.gl.GetIdentifier(), " PublishResults event is nil")
		return
	}
	if event.Results == nil || len((*event.Results)) == 0 {
		logger.Error(fpcgl.gl.GetIdentifier(), " PublishResults event.Results is empty")
		return
	}
	resultData := response.FishPrawnCrabResult{
		Result: &(*event.Results)[0].Value,
	}

	fpcgl.pb.Publish(fpcgl.mg.GenerateMessage("result", &resultData))
}
