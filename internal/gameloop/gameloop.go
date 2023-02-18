package gameloop

import (
	"errors"
	"fmt"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/measure"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/mutex"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/jinzhu/now"
	"gorm.io/gorm"
)

const (
	LeewayMS = 30 //ms for accepting date related comparison on start
)

type PrepareCallback func(elapsedTime Milliseconds)
type SettleEventCallback func(tx *gorm.DB, event *models.Event) error

type Milliseconds int64

type Datasource interface {
	GetIdentifier() string
	GetTableID() int64
}

type Gamelooper interface {
	Start()
	Stop()
	GetCurrentEvent() *models.Event
}

type Gameloop interface {
	GetIdentifier() string                                                                //get identifier
	Start()                                                                               //start gameloop
	Stop()                                                                                //stop gameloop
	SetPrepareCallback(callback PrepareCallback)                                          //set prepare callback which will be called on start and on last item of the phase
	CreatePhase(phase GamePhase)                                                          //create phase
	GetUpcomingEvents() *[]models.Event                                                   //get upcoming events
	GetCurrentEvent() *models.Event                                                       //get current event can be nil if start datetime is greater than now
	GetStartDateTime() time.Time                                                          //get start date time
	ElapsedTime() Milliseconds                                                            //get elapsed time base on start datetime
	ElapsedTimeOffset(prevElapsedTime Milliseconds) Milliseconds                          //generate offset from elapsed time
	UpdateEventStatus(tx *gorm.DB, eventID int64, status int16) error                     //update event status of event
	SettleTicketsForEvent(tx *gorm.DB, eventID int64) error                               //settle tickets for event
	SettleEventsBefore(tx *gorm.DB, before time.Time, callback SettleEventCallback) error //settle events before specified time with tickets
	ProcessDailyMaxWinnings(tx *gorm.DB) error                                            //process daily max winnings
	GetTotalWinLoss() float64                                                             //get total win loss from process daily max winnings
	GetDatasource() Datasource                                                            //get datasource useful for reusing
}

type gameloop struct {
	datasource         Datasource
	isRunning          bool
	eventStartDateTime time.Time
	prepareCallback    *PrepareCallback
	phases             gamePhases
	upcomingEvents     mutex.Data[*[]models.Event]
	totalWinLoss       float64
}

func NewGameLoop(datasource Datasource) Gameloop {
	return &gameloop{
		datasource:         datasource,
		eventStartDateTime: utils.TimeNow(),
	}
}

func (gl *gameloop) GetIdentifier() string {
	return gl.datasource.GetIdentifier() + " gameloop"
}

func (gl *gameloop) Start() {
	gl.PreloadUpcomingEvents()
	gl.Prepare()
	gl.Run()
}

func (gl *gameloop) Stop() {
	gl.isRunning = false
}

func (gl *gameloop) SetPrepareCallback(callback PrepareCallback) {
	gl.prepareCallback = &callback
}

func (gl *gameloop) CreatePhase(phase GamePhase) {
	gl.phases = append(gl.phases, phase)
}

func (gl *gameloop) PreloadUpcomingEvents() {
	if gl.prepareCallback != nil {
		(*gl.prepareCallback)(0)
	}
	exec := measure.NewExecution()
	defer func() {
		logger.Debug(gl.GetIdentifier(), " PreloadUpcomingEvents execution time: ", exec.Done())
	}()
	events := []models.Event{}
	rawQuery := fmt.Sprintf(`
	(
	SELECT
		*
	FROM
		public.mini_game_event
	WHERE
		mini_game_table_id = %v
		AND start_datetime <= now()
		AND now() < (start_datetime + INTERVAL '%v milliseconds')
	ORDER BY
		ctime DESC
	LIMIT 1)
	UNION 
	(
	SELECT
		*
	FROM
		public.mini_game_event
	WHERE
		mini_game_table_id = %v
	AND start_datetime > now()
	ORDER BY
		ctime ASC
	)
	ORDER BY ctime ASC 
	`, gl.datasource.GetTableID(), gl.phases.TotalDuration(), gl.datasource.GetTableID())
	if err := db.Shared().Preload("Results", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("result_type ASC")
	}).Raw(rawQuery).Find(&events).Error; err != nil {
		logger.Error(gl.GetIdentifier(), " PreloadUpcomingEvents error: ", err.Error())
		return
	}
	gl.upcomingEvents.Lock()
	defer gl.upcomingEvents.Unlock()
	gl.upcomingEvents.Data = &events
}

func (gl *gameloop) GetUpcomingEvents() *[]models.Event {
	gl.upcomingEvents.Lock()
	defer gl.upcomingEvents.Unlock()
	return gl.upcomingEvents.Data
}

func (gl *gameloop) GetCurrentEvent() *models.Event {
	gl.upcomingEvents.Lock()
	defer gl.upcomingEvents.Unlock()
	if gl.upcomingEvents.Data != nil {
		timeNow := utils.TimeNow()

		for i := 0; i < len(*gl.upcomingEvents.Data); i++ {
			startTime := (*gl.upcomingEvents.Data)[i].StartDatetime
			totalDuration := time.Duration(gl.phases.TotalDuration()) * time.Millisecond
			endTime := startTime.Add(totalDuration)

			if startTime.Add(-LeewayMS*time.Millisecond).Before(timeNow) && endTime.After(timeNow) {
				return &(*gl.upcomingEvents.Data)[i]
			}
		}
		logger.Warning(gl.GetIdentifier(), " GetCurrentEvent error: ", utils.PrettyJSON(map[string]any{
			"prev_eventStartDateTime": gl.eventStartDateTime,
			"totalDuration":           gl.phases.TotalDuration(),
			"timeNow":                 timeNow,
			"upcomingEvents": types.Array[models.Event](*(gl.upcomingEvents.Data)).Map(func(value models.Event) any {
				return fmt.Sprint("ID: ", *value.ID, ", startDateTime: ", value.StartDatetime, ", diffNow: ", value.StartDatetime.Sub(timeNow).Milliseconds(), "ms")
			}),
		}))
	}
	return nil
}

func (gl *gameloop) GetStartDateTime() time.Time {
	return gl.eventStartDateTime
}

func (gl *gameloop) ElapsedTime() Milliseconds {
	return Milliseconds(time.Since(gl.eventStartDateTime).Milliseconds())
}

func (gl *gameloop) ElapsedTimeOffset(prevElapsedTime Milliseconds) Milliseconds {
	return gl.ElapsedTime() - prevElapsedTime
}

func (gl *gameloop) UpdateEventStatus(tx *gorm.DB, eventID int64, status int16) error {
	if err := tx.Exec(fmt.Sprintf(`
	UPDATE
		mini_game_event
	SET
		status = ?,
		mtime = NOW()%v
	WHERE
		id = ?
	`, utils.IfElse(status == constants.EVENT_STATUS_SETTLED, `,
	settlement_date = NOW(),
	local_data_version = local_data_version + 1`, "")), status, eventID).Error; err != nil {
		return errors.New("settle tickets for event: " + err.Error())
	}
	return nil
}

func (gl *gameloop) SettleTicketsForEvent(tx *gorm.DB, eventID int64) error {
	if err := tx.Exec(`
	UPDATE
		mini_game_ticket
	SET
		status = ?,
		mtime = NOW(),
		status_mtime = NOW(),
		local_data_version = local_data_version + 1
	WHERE
		event_id = ?
	`, constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT, eventID).Error; err != nil {
		return errors.New("settle tickets for event: " + err.Error())
	}
	return nil
}

func (gl *gameloop) SettleEventsBefore(tx *gorm.DB, before time.Time, callback SettleEventCallback) error {
	rawQuery := fmt.Sprintf(`
	SELECT
		*
	FROM
		mini_game_event
	WHERE
		mini_game_table_id = %v
	    AND status < %v
		AND now() > (start_datetime + INTERVAL '%v milliseconds')
	ORDER BY
		start_datetime DESC
	LIMIT 50
	`, gl.datasource.GetTableID(), constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT, time.Since(before).Milliseconds())

	return tx.Transaction(func(tx *gorm.DB) error {
		events := []models.Event{}

		if err := tx.Preload("Tickets.ComboTickets").Raw(rawQuery).Find(&events).Error; err != nil {
			return errors.New("SettleEventsBefore Get events: " + err.Error())
		}
		for i := 0; i < len(events); i++ {
			if err := callback(tx, &events[i]); err != nil {
				return errors.New("SettleEventsBefore callback: " + err.Error())
			}
			if err := gl.UpdateEventStatus(tx, *events[i].ID, constants.EVENT_STATUS_SETTLED); err != nil {
				return errors.New("SettleEventsBefore UpdateEventStatus: " + err.Error())
			}
		}
		return nil
	})
}

func (gl *gameloop) ProcessDailyMaxWinnings(tx *gorm.DB) error {
	gameTable := models.GameTable{ID: gl.datasource.GetTableID()}

	if err := tx.First(&gameTable).Error; err != nil {
		return err
	}
	if gameTable.DisableOnDailyMaxWinnings == nil { //ignore no daily max winning set
		return nil
	}
	totalWinLoss := 0.0 //negative value means house wins
	startOfDay := now.BeginningOfDay()

	if gameTable.LastDailyMaxWinningsTriggeredDatetime != nil && gameTable.LastDailyMaxWinningsTriggeredDatetime.After(startOfDay) {
		startOfDay = *gameTable.LastDailyMaxWinningsTriggeredDatetime
	}
	if err := tx.Raw(`
	SELECT 
		CASE WHEN SUM( ticket.win_loss_amount ) IS NULL THEN 0 ELSE SUM ( ticket.win_loss_amount/ticket.exchange_rate ) END AS total_win_loss 
	FROM
		mini_game_ticket ticket
	WHERE
		ticket.ctime BETWEEN ?
		AND ?
		AND ticket.status >= ?
		AND mini_game_table_id = ?
	`, startOfDay, now.EndOfDay(), constants.TICKET_STATUS_SETTLEMENT_IN_PROGRESS, gl.datasource.GetTableID()).
		Pluck("total_win_loss", &totalWinLoss).Error; err != nil {
		return err
	}
	gl.totalWinLoss = totalWinLoss
	if totalWinLoss >= *gameTable.DisableOnDailyMaxWinnings {
		uGameTable := models.GameTable{
			ID:                                    gameTable.ID,
			LastDailyMaxWinningsTriggeredDatetime: utils.Ptr(utils.TimeNow()),
			DisabledByMgTrigger:                   true,
		}

		if err := tx.Updates(uGameTable).Error; err != nil {
			return err
		}
	}
	return nil
}

func (gl *gameloop) GetTotalWinLoss() float64 {
	return gl.totalWinLoss
}

func (gl *gameloop) GetDatasource() Datasource {
	return gl.datasource
}

// private
func (gl *gameloop) Run() {
	if gl.isRunning {
		return
	}
	termCtx, termCancel := utils.TerminateContext()
	noEventCount := 0

	gl.isRunning = true
	defer termCancel()
	go func() {
		<-termCtx.Done()
		logger.Info(gl.GetIdentifier(), " terminate signal received, exiting...")
		gl.phases.CancelSleep()
		logger.Info(gl.GetIdentifier(), " phases sleep cancelled")
		gl.isRunning = false
	}()
	for gl.isRunning {
		if gl.phases.IsTriggered() && gl.ElapsedTime() >= gl.phases.TotalDuration() {
			if err := gl.Prepare(); err != nil {
				logger.Error(gl.GetIdentifier(), " Prepare error: ", err.Error())
				time.Sleep(50 * time.Millisecond) //sleep for 50 milliseconds
				gl.PreloadUpcomingEvents()        //preload upcoming events
				noEventCount++
				if noEventCount >= 3 {
					//when no event for 3 count use timeNow as event start date time
					gl.phases.ResetTrigger()
					gl.phases.ResetOffset()
					gl.eventStartDateTime = utils.TimeNow()
					noEventCount = 0
				} else {
					gl.phases.Trigger()
					continue
				}
			} else {
				gl.phases.ResetTrigger()
				noEventCount = 0
			}
		}
		gl.phases.Run(gl.ElapsedTime(), gl.PreloadUpcomingEvents)
	}
	gl.isRunning = false
	logger.Info(gl.GetIdentifier(), " Run exited")
}

func (gl *gameloop) Prepare() error {
	if event := gl.GetCurrentEvent(); event != nil {
		gl.eventStartDateTime = event.StartDatetime
		return nil
	}

	return errors.New("current event is nil")
}

// static functions
func UnixNanoToTime(unixNano int64) time.Time {
	return time.Unix(0, unixNano)
}

// message generator
type MessageGenerator interface {
	GenerateStateMessage(name string, end *time.Time) string
	GenerateMessage(mType string, data response.ResponseData) string
}
type _messageGenerator struct {
	identifier string
}

func NewMessageGenerator(identfier string) MessageGenerator {
	return &_messageGenerator{identifier: identfier}
}

func (mg *_messageGenerator) GenerateStateMessage(name string, end *time.Time) string {
	endTS := 0.0

	if end != nil {
		endTS = utils.TimeToUnixTS(*end)
	}
	state := response.State{
		Name: name,
		End:  endTS,
		Gts:  utils.GenerateUnixTS(),
	}

	redis.SetPublishState(mg.identifier, &state)
	return mg.GenerateMessage("state", &state)
}

func (mg *_messageGenerator) GenerateMessage(mType string, data response.ResponseData) string {
	return (&response.Response{
		Type: mType,
		Data: data,
	}).JSON()
}
