package controller

import (
	"fmt"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/tasks"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/hashutil"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type EventController interface {
	GenerateEvents(ctx *gin.Context)
	createTableEvents(gameID int64, tableID int64)
	UpdateEventState(event models.Event)
}

type eController struct {
	service  service.EventService
	gService service.GameService
	hService service.HashService
	sService service.SelectionService
}

func NewEController(service service.EventService, gService service.GameService, hService service.HashService, sService service.SelectionService) EventController {
	return &eController{
		service:  service,
		gService: gService,
		hService: hService,
		sService: sService,
	}
}

// 0 (Enabled),
// 1 (Active),
// 2 (Disabled),
// 3 (For Settlement),
// 4 (Settlement in Progress),
// 5 (Settled),
// 6 (Cancelled)

func (c *eController) GenerateEvents(ctx *gin.Context) {
	var gameID int64

	fmt.Sscan(ctx.Param("game_id"), &gameID)

	// get game
	games, _ := c.gService.GetGames(gameID)

	// get tables
	for _, table := range *games.GameTables {
		tableID := table.ID

		go c.createTableEvents(gameID, tableID)
	}
}

func (c *eController) createTableEvents(gameID int64, tableID int64) {
	params := map[string]interface{}{
		"status":             0, // get enabled events
		"mini_game_table_id": tableID,
	}

	futureEventCount := settings.FUTURE_EVENTS_COUNT
	gameDuration := settings.LOL_TOWERS_GAME_DURATION

	fEvents, _ := c.service.GetFutureEvents(params)
	lackingFutureEventCount := (futureEventCount - len(fEvents))

	// get available events assure 3 events available [status enabled]
	for i := 0; i < lackingFutureEventCount; i++ {
		var prevEventID *int64
		tm := time.Now()
		startTm := tm

		lastEvent := c.service.GetLastEventByTableID(tableID)

		if lastEvent != nil {
			prevEventID = lastEvent.ID
			tempTm := lastEvent.StartDatetime.Add(time.Second * time.Duration(gameDuration))
			diff := tm.Sub(tempTm).Seconds()

			if diff < float64(gameDuration) {
				startTm = lastEvent.StartDatetime.Add(time.Second * time.Duration(gameDuration))
			}
		}

		// get next sequence
		nhs := c.getNextHashSeq(tableID)
		// continue if no available hash to prevent panic
		if nhs.Value == "" {
			continue
		}
		result, selectionHeaderID := c.generateHashResult(nhs.Value, gameID)

		// get table info
		table := c.gService.GetTable(tableID)

		payload := models.Event{
			LocalDataVersion:   1,
			ESDataVersion:      0,
			SyncStatus:         0,
			Ctime:              tm,
			Mtime:              tm,
			Name:               table.Name,
			IsAutoPlayExecuted: false,
			StartDatetime:      startTm,
			Status:             0,                     // 0 - Enabled
			MaxBet:             table.MaxBetAmount,    // use table setting as default
			MaxPayout:          table.MaxPayoutAmount, // use table setting as default
			GameID:             gameID,
			HashSequenceID:     nhs.ID, // get available Game hash sequence id
			SelectionHeaderID:  selectionHeaderID,
			TableID:            tableID,
			PrevEventID:        prevEventID,
		}

		createdEvent, err := c.service.CreateEvents(payload)

		// proccess tasks event state
		go c.UpdateEventState(createdEvent)

		if err != nil {
			break
		}

		// create event result
		var res models.EventResult
		res.ResultType = 28
		res.EventID = createdEvent.ID
		res.Value = result

		c.service.CreateEventResult(res)
	}
}

func (c *eController) generateHashResult(hash string, gameID int64) (Result string, SelectionHeaderID int64) {
	result, err := hashutil.LOLTowerGenerateResult(hashutil.NewHash(hash))

	if err != nil {
		logger.Error("loltower LOLTowerGenerateResult error: ", err.Error())
	}
	// get result selection line
	sl := c.sService.GetSelection(gameID)

	return result, sl.ID
}

func (c *eController) getNextHashSeq(tableID int64) models.HashSequence {
	var nextSeq models.HashSequence

	temp := map[string]interface{}{
		"mini_game_table_id": tableID,
	}
	lastSeq := c.service.GetLatestHashSeq(temp)
	maxHQ := settings.DEF_MAX_SEQUENCE

	// no last sequence used
	if lastSeq.ID == 0 && lastSeq.Sequence == 0 {
		nextSeq = c.hService.GetNextQueuedHashSequence(tableID)
		//set to active
		c.hService.SetHashStatus(nextSeq.HashID, "active")
	} else if lastSeq.Sequence == maxHQ { // last sequence = to max seq use queued hash and update hash to done
		// update hash to done
		c.hService.SetHashStatus(lastSeq.HashID, "done")
		// get next hash header set to active and get next sequence
		nextSeq = c.hService.GetNextQueuedHashSequence(tableID)
		// set to active header
		c.hService.SetHashStatus(nextSeq.HashID, "active")
	} else { // take next sequence
		isActive := c.hService.IsActiveHash(lastSeq.HashID)
		if isActive {
			nextSeq = c.hService.GetNextHashSequence(tableID, lastSeq.Sequence, lastSeq.HashID)
		} else {
			nextSeq = c.hService.GetNextQueuedHashSequence(tableID)
			//set to active
			c.hService.SetHashStatus(nextSeq.HashID, "active")
		}
	}

	return nextSeq
}

func (c *eController) UpdateEventState(event models.Event) {
	// var eventID int64
	// fmt.Sscan(ctx.Param("event_id"), &eventID)

	// tasks testing
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: settings.REDIS_HOST,
		DB:   2,
	})
	defer client.Close()

	// event := c.service.GetLastEventByEventID(eventID)
	currentTime := time.Now()
	duration := event.StartDatetime.Sub(currentTime)

	// start betting phase
	task, _ := tasks.NewBettingTask(event)
	info, _ := client.Enqueue(task, asynq.ProcessIn(duration))
	logger.Info(fmt.Sprintf("enqueued task: id=%s queue=%s", info.ID, info.Queue))

	// start stop betting phase
	stopBettingTime := event.StartDatetime.Add(7 * time.Second)
	stDuration := stopBettingTime.Sub(currentTime)
	bTask, _ := tasks.NewStopBettingTask(event)
	bInfo, _ := client.Enqueue(bTask, asynq.ProcessIn(stDuration))
	logger.Info(fmt.Sprintf("enqueued task: id=%s queue=%s", bInfo.ID, bInfo.Queue))

	// show result phase
	showResultTime := event.StartDatetime.Add(10 * time.Second)
	sDuration := showResultTime.Sub(currentTime)
	sTask, _ := tasks.NewShowResultTask(event)
	sInfo, _ := client.Enqueue(sTask, asynq.ProcessIn(sDuration))
	logger.Info(fmt.Sprintf("enqueued task: id=%s queue=%s", sInfo.ID, sInfo.Queue))
}
