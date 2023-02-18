package gameloop_lolcouple

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	gamemanager_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	wsclient_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/measure"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type lolCoupleGameloop struct {
	gl gameloop.Gameloop
	gm gamemanager.GameManager
	pb wsclient.MessagePublishBroker
	mg gameloop.MessageGenerator
}

func NewGameLoop() gameloop.Gamelooper {
	gm := gamemanager_lolcouple.NewGameManager()   //load gamemanager
	gl := gameloop.NewGameLoop(gm.GetDatasource()) //reuse gamemanager datasource
	lolCoupleGameloop := lolCoupleGameloop{
		gm: gm,
		gl: gl,
		pb: wsclient_lolcouple.NewPublishBroker(),
		mg: gameloop.NewMessageGenerator(constants_lolcouple.Identifier),
	}

	lolCoupleGameloop.Initialize()
	return &lolCoupleGameloop
}

func (lcgl *lolCoupleGameloop) Start() {
	lcgl.gl.Start()
}

func (lcgl *lolCoupleGameloop) Stop() {
	lcgl.gl.Stop()
}

func (lcgl *lolCoupleGameloop) GetCurrentEvent() *models.Event {
	return lcgl.gl.GetCurrentEvent()
}

func (lcgl *lolCoupleGameloop) Initialize() {
	lcgl.gl.SetPrepareCallback(func(elapsedTime gameloop.Milliseconds) {
		if err := lcgl.gm.CreateFutureHashes(); err != nil {
			logger.Error(lcgl.gl.GetIdentifier(), " CreateFutureHashes error: ", err.Error())
		}
		if err := lcgl.gm.CreateFutureEvents(); err != nil {
			logger.Error(lcgl.gl.GetIdentifier(), " CreateFutureEvents error: ", err.Error())
		}
	})
	lcgl.gl.CreatePhase(gameloop.NewGamePhase(
		"BETTING",
		0,
		constants_lolcouple.StartBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_lolcouple.StartBetMS)
			remainingTime := endMS - elapsedTime
			event := lcgl.gl.GetCurrentEvent()

			if remainingTime := endMS - lcgl.gl.ElapsedTime(); remainingTime > 0 {
				lcgl.pb.Publish(lcgl.mg.GenerateStateMessage("BETTING", utils.Ptr(lcgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			if event != nil {
				logger.Debug(lcgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				if err := lcgl.gl.UpdateEventStatus(db.Shared(), *event.ID, constants.EVENT_STATUS_ACTIVE); err != nil {
					logger.Error(lcgl.gl.GetIdentifier(), " UpdateEventStatus active(", constants.EVENT_STATUS_ACTIVE, ") error: ", err.Error())
				}
			} else {
				logger.Debug(lcgl.gl.GetIdentifier(), " BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(lcgl.gl.GetIdentifier(), " BETTING total execution time: ", lcgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-lcgl.gl.ElapsedTimeOffset(elapsedTime))
		}))
	lcgl.gl.CreatePhase(gameloop.NewGamePhase(
		"STOP_BETTING",
		constants_lolcouple.StartBetMS,
		constants_lolcouple.StartBetMS+constants_lolcouple.StopBetMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_lolcouple.StartBetMS + constants_lolcouple.StopBetMS)
			remainingTime := endMS - elapsedTime
			event := lcgl.gl.GetCurrentEvent()

			if remainingTime := endMS - lcgl.gl.ElapsedTime(); remainingTime > 0 {
				lcgl.pb.Publish(lcgl.mg.GenerateStateMessage("STOP_BETTING", utils.Ptr(lcgl.gl.GetStartDateTime().Add(time.Duration(endMS)*time.Millisecond))))
			}
			if event != nil {
				logger.Debug(lcgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
			} else {
				logger.Debug(lcgl.gl.GetIdentifier(), " STOP_BETTING elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(lcgl.gl.GetIdentifier(), " STOP_BETTING total execution time: ", lcgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-lcgl.gl.ElapsedTimeOffset(elapsedTime))
		},
	))
	lcgl.gl.CreatePhase(gameloop.NewGamePhase(
		"SHOW_RESULT",
		constants_lolcouple.StartBetMS+constants_lolcouple.StopBetMS,
		constants_lolcouple.StartBetMS+constants_lolcouple.StopBetMS+constants_lolcouple.ShowResultMaxMS,
		func(elapsedTime gameloop.Milliseconds, phase gameloop.GamePhase) {
			endMS := gameloop.Milliseconds(constants_lolcouple.StartBetMS + constants_lolcouple.StopBetMS + constants_lolcouple.ShowResultMaxMS)
			remainingTime := endMS - elapsedTime
			overflowMS := gameloop.Milliseconds(0)
			event := lcgl.gl.GetCurrentEvent()

			if remainingTime := endMS - lcgl.gl.ElapsedTime(); remainingTime > 0 {
				lcgl.pb.Publish(lcgl.mg.GenerateStateMessage("SHOW_RESULT", nil)) //remove end since it will determine the result
				utils.PerformAfter(50*time.Millisecond, func() {
					lcgl.PublishResults(event, overflowMS)
				})
				go SendBonusNotificationIfNeeded(lcgl.gl.GetUpcomingEvents(), event, 3)
			}
			if event != nil {
				overflowDuration, err := GetOverflowDuration(event.Results)

				if err != nil {
					logger.Error(lcgl.gl.GetIdentifier(), " GetOverflowDuration error: ", err.Error())
				}
				overflowMS = gameloop.Milliseconds(overflowDuration.Milliseconds())
				endMS = endMS - overflowMS
				remainingTime := endMS - elapsedTime

				logger.Debug(lcgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", utils.PrettyJSON(event.ID))
				if err := lcgl.UpdateFutureEvents(event.Results, overflowDuration); err != nil {
					logger.Error(lcgl.gl.GetIdentifier(), " UpdateFutureEvents error: ", err.Error())
				}
				go func() {
					exec := measure.NewExecution()
					defer func() {
						logger.Debug(lcgl.gl.GetIdentifier(), " SettleEventsBefore/ProcessDailyMaxWinnings execution time: ", exec.Done())
					}()
					if err := lcgl.gl.SettleEventsBefore(db.Shared(), utils.TimeNow(), func(tx *gorm.DB, event *models.Event) error {
						return lcgl.gl.SettleTicketsForEvent(tx, *event.ID)
					}); err != nil {
						logger.Error(lcgl.gl.GetIdentifier(), " SettleEventsBefore error: ", err.Error())
					}
					if err := lcgl.gl.ProcessDailyMaxWinnings(db.Shared()); err != nil {
						logger.Error(lcgl.gl.GetIdentifier(), " ProcessDailyMaxWinnings error: ", err.Error())
					}
					logger.Debug(lcgl.gl.GetIdentifier(), " TotalWinLoss: ", lcgl.gl.GetTotalWinLoss())
				}()
			} else {
				logger.Debug(lcgl.gl.GetIdentifier(), " SHOW_RESULT elapsedTime: ", elapsedTime, "ms remainingTime: ", remainingTime, "ms event: ", nil)
			}
			logger.Debug(lcgl.gl.GetIdentifier(), " SHOW_RESULT total execution time: ", lcgl.gl.ElapsedTimeOffset(elapsedTime), "ms")
			phase.SetEndMSOffset(-(overflowMS + lcgl.gl.ElapsedTimeOffset(elapsedTime)))
		},
	))

}

func (lcgl *lolCoupleGameloop) UpdateFutureEvents(eventResults *[]models.EventResult, overflowDuration time.Duration) error {
	exec := measure.NewExecution()
	defer func() {
		logger.Debug(lcgl.gl.GetIdentifier(), " UpdateFutureEvents execution time: ", exec.Done())
	}()
	if futureEvents := lcgl.gm.GetFutureEvents(); futureEvents != nil {
		for i := 0; i < len(*futureEvents); i++ {
			uEvent := models.Event{
				ID:            (*futureEvents)[i].ID,
				StartDatetime: (*futureEvents)[i].StartDatetime.Add(-overflowDuration),
			}

			if err := db.Shared().Updates(&uEvent).Error; err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("no future events")
}

func (lcgl *lolCoupleGameloop) PublishResults(event *models.Event, overflowMS gameloop.Milliseconds) {
	if event == nil {
		logger.Error(lcgl.gl.GetIdentifier(), " PublishResults event is nil")
		return
	}
	if event.Results == nil || len((*event.Results)) == 0 {
		logger.Error(lcgl.gl.GetIdentifier(), " PublishResults event.Results is empty")
		return
	}
	results := []string{}

	if err := json.Unmarshal([]byte((*event.Results)[0].Value), &results); err != nil {
		logger.Error(lcgl.gl.GetIdentifier(), "PublishResults event.Result unmarshal error: ", err.Error())
		return
	}
	remainingOverflowMS := constants_lolcouple.ShowResultMaxMS - overflowMS
	baseMS := gameloop.Milliseconds(constants_lolcouple.StartBetMS + constants_lolcouple.StopBetMS)
	totalResultMS := 0
	resultData := response.LOLCoupleResultData{}
	logger.Debug(lcgl.gl.GetIdentifier(), " remainingOverflowMS: ", remainingOverflowMS, "ms")

	for i := 0; i < len(results); i++ {
		result := strings.Split(results[i], ":")
		maleValue := types.String(result[0]).Int().Int8()
		femaleValue := types.String(result[1]).Int().Int8()
		coupleValue := maleValue + femaleValue
		isBonus := coupleValue == constants_lolcouple.SelectionBonusCoupleValue
		startMS := baseMS + gameloop.Milliseconds(totalResultMS)
		totalResultMS += ResultIndexMS(i, isBonus)
		resultsMS := gameloop.Milliseconds(totalResultMS)
		endMS := baseMS + resultsMS
		shouldTerminate := (remainingOverflowMS - resultsMS) <= 0

		switch i {
		case 0:
			resultData.Result1 = &results[i]
		case 1:
			resultData.Result2 = &results[i]
		case 2:
			resultData.Result3 = &results[i]
		}
		elapsedTime := lcgl.gl.ElapsedTime()

		logger.Debug(lcgl.gl.GetIdentifier(), " startMS: ", startMS, "ms elapsedTime: ", elapsedTime, "ms endMS: ", endMS, "ms")
		if (startMS-gameloop.LeewayMS) <= elapsedTime && elapsedTime < endMS {
			logger.Debug(lcgl.gl.GetIdentifier(), " publish result: ", utils.PrettyJSON(result))
			resultData.Gts = utils.GenerateUnixTS()
			lcgl.pb.Publish(lcgl.mg.GenerateMessage("result", &resultData))
			if shouldTerminate {
				break
			}
		} else {
			logger.Warning(lcgl.gl.GetIdentifier(), " publish failed elapsed time diff is ", startMS-elapsedTime)
		}
		if sleepDuration := endMS - lcgl.gl.ElapsedTime(); sleepDuration > 0 {
			ctx, _ := utils.TerminateContext()

			utils.Sleep(time.Duration(sleepDuration)*time.Millisecond, ctx)
		}
		if shouldTerminate {
			break
		}
	}
}

func GetOverflowDuration(eventResults *[]models.EventResult) (time.Duration, error) {
	totalOverflow := constants_lolcouple.ShowResultMaxMS

	if eventResults != nil {
		results := []string{}

		if err := json.Unmarshal([]byte((*eventResults)[0].Value), &results); err != nil {
			return time.Duration(totalOverflow) * time.Millisecond, err
		}

		for i := 0; i < len(results); i++ {
			result := strings.Split(results[i], ":")
			maleValue := types.String(result[0]).Int().Int8()
			femaleValue := types.String(result[1]).Int().Int8()
			coupleValue := maleValue + femaleValue
			isBonus := coupleValue == constants_lolcouple.SelectionBonusCoupleValue

			totalOverflow -= ResultIndexMS(i, isBonus)
			if !isBonus { //return early when not 7
				break
			}
		}
	}
	return time.Duration(totalOverflow) * time.Millisecond, nil
}

func ResultIndexMS(index int, isBonus bool) int {
	switch index {
	case 0:
		if isBonus {
			return constants_lolcouple.ShowResult1BonusMS
		}
		return constants_lolcouple.ShowResult1MS
	case 1:
		return constants_lolcouple.ShowResult2MS
	case 2:
		if isBonus {
			return constants_lolcouple.ShowResult3BonusMS
		}
		return constants_lolcouple.ShowResult3MS
	default:
		panic("lolcouple ResultIndexMS index out of range")
	}
}

// next is the next events to check
func SendBonusNotificationIfNeeded(upcomingEvents *[]models.Event, currentEvent *models.Event, next int) {
	if !(types.Array[string]{"dev", "live"}).Constains(settings.GetEnvironment().String()) ||
		currentEvent == nil ||
		upcomingEvents == nil ||
		len(*upcomingEvents) == 0 {
		return
	}
	next = utils.IfElse(settings.GetEnvironment().String() == "live", 0, next)

	for i := 0; i < len(*upcomingEvents); i++ {
		if *(*upcomingEvents)[i].ID == *currentEvent.ID && len(*upcomingEvents) > (i+next) {
			if overflow, _ := GetOverflowDuration((*upcomingEvents)[i+next].Results); overflow == 0 && os.Getenv("MG_LOLCOUPLE_ALLOW_BONUS_NOTIFY") == "true" {
				switch settings.GetEnvironment().String() {
				case "dev":
					slack.SendPayload(slack.NewLootboxNotification("lolcouple", "BONUS ROUND ON AFTER "+*types.Int(next).String().Ptr()+" EVENTS!!!"), slack.LOLCoupleMonitor777Dev)
				case "live":
					slack.SendPayload(slack.NewLootboxNotification("lolcouple", "BONUS ROUND!!!"), slack.LOLCoupleMonitor777Live)
				}
			}
			break
		}
	}
}
