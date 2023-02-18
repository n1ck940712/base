package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm"
)

type EventService interface {
	CreateEvents(models.Event) (models.Event, error)
	GetFutureEvents(params map[string]interface{}) ([]models.Event, error)
	GetLastEventByTableID(tableID int64) *models.Event
	GetLastFutureEventByTableID(tableID int64) *models.Event
	GetEvent(tableID int64) *models.Event
	GetLatestHashSeq(map[string]interface{}) models.HashSequence
	CreateEventResult(payload models.EventResult) models.EventResult
	GetLastEventByEventID(eventID int64) *models.Event
	UpdateEventStatus(eventID int64, event models.Event) models.Event
	GetEventLolTower(eventID int64) string
	GetEventLolTowerBomb(eventID int64) string
	GetEventDetails(eventID int64) models.Event
	IsExist(eventID int64) int64
	GetCurrentEventByTableID(tableID int64, gameID int64) *models.Event
	SettleUnsettledPrevEvents(event models.Event, tableID int64)
}

type eventService struct {
	event models.Event
}

func NewEvent() EventService {
	return &eventService{}
}

func (service *eventService) CreateEvents(payload models.Event) (models.Event, error) {
	result := DB.Table("mini_game_event").Select(
		"local_data_version",
		"es_data_version",
		"sync_status",
		"ctime",
		"mtime",
		"name",
		"is_auto_play_executed",
		"status",
		"max_bet",
		"max_payout",
		"game_id",
		"mini_game_hash_sequence_id",
		"selection_header_id",
		"start_datetime",
		"ground_type",
		"mini_game_table_id",
		"prev_event_id",
	).Create(&payload)

	if result.Error != nil {
		logger.Error("Error Creating Event")
	}

	return payload, result.Error
}

func (service *eventService) GetFutureEvents(params map[string]interface{}) ([]models.Event, error) {
	var res []models.Event
	result := DB.Table("mini_game_event").
		Where(params).
		Where("start_datetime >= ?", time.Now()).
		Find(&res)

	return res, result.Error
}

func (service *eventService) GetLastEventByTableID(tableID int64) *models.Event {
	var res models.Event
	result := DB.Table("mini_game_event").
		Where("mini_game_table_id = ?", tableID).
		Last(&res)

	if result.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &res
}

func (service *eventService) GetEvent(eventID int64) *models.Event {
	var res models.Event
	result := DB.Table("mini_game_event").
		Where("id = ?", eventID).
		First(&res)

	if result.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &res
}

func (service *eventService) GetLatestHashSeq(param map[string]interface{}) models.HashSequence {
	where := map[string]interface{}{}
	if _, ok := param["mini_game_table_id"]; ok {
		where["mini_game_table_id"] = param["mini_game_table_id"]
	}

	var hq models.HashSequence

	result := DB.Raw(`
		SELECT hq.* FROM mini_game_event e
		JOIN mini_game_minigamehashsequence hq on e.mini_game_hash_sequence_id = hq.id 
		JOIN mini_game_minigamehash h on h.id = hq.mini_game_hash_id
		WHERE e.mini_game_table_id = ?
		AND h.status = 'active'
		ORDER BY e.id DESC LIMIT 1`, where["mini_game_table_id"]).Scan(&hq)

	if result.Error != nil {
		logger.Error("Error GetLatestHashSeq")
	}

	return hq
}

func (service *eventService) CreateEventResult(payload models.EventResult) models.EventResult {
	tm := time.Now()
	payload.Ctime = tm
	payload.Mtime = tm
	result := DB.Table("mini_game_eventresult").Create(&payload)

	if result.Error != nil {
		logger.Error("Error CreateEventResult")
	}

	return payload
}

func (service *eventService) GetLastFutureEventByTableID(tableID int64) *models.Event {
	var res models.Event
	result := DB.Table("mini_game_event").
		Where("mini_game_table_id = ?", tableID).
		Where("start_datetime >= ?", time.Now()).
		Last(&res)

	if result.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &res
}

func (service *eventService) GetLastEventByEventID(eventID int64) *models.Event {
	var res models.Event
	result := DB.Table("mini_game_event").
		Where("id = ?", eventID).
		First(&res)

	if result.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &res
}

func (service *eventService) UpdateEventStatus(eventID int64, event models.Event) models.Event {
	event.Mtime = time.Now()
	var result *gorm.DB
	if event.Status == constants.EVENT_STATUS_SETTLEMENT_IN_PROGRESS || event.Status == constants.EVENT_STATUS_SETTLED {
		event.SettlementDate = &event.Mtime
		if event.Status == constants.EVENT_STATUS_SETTLED {
			event.LocalDataVersion += 1
		}
	}

	result = DB.Table("mini_game_event").Select("status", "mtime", "settlement_date", "local_data_version").Updates(&event).Where("id = ?", eventID)

	if result.Error != nil {
		logger.Error("Error UpdateEvent")
	}

	return event
}

func (service *eventService) checkBombContains(axe []int, str int) bool {
	for _, v := range axe {
		if v == str {
			return true
		}
	}

	return false
}

func (service *eventService) GetEventLolTower(eventID int64) string {
	eventResult := models.EventResult{}

	if DB.Where(&models.EventResult{EventID: &eventID}).First(&eventResult).RowsAffected == 0 {
		return ""
	}

	fmt.Println("eventResult.Value:", eventResult.Value)

	return eventResult.Value
}

func (service *eventService) GetEventLolTowerBomb(eventID int64) string {
	var res models.EventResult
	var r string
	result := DB.Table("mini_game_eventresult").
		Where("event_id = ?", eventID).
		First(&res)

	if result.Error == gorm.ErrRecordNotFound {
		return ""
	}

	cards := []int{1, 2, 3, 4, 5}
	bomb := []int{}
	axeStrings := strings.Split(res.Value, ",")
	axe := make([]int, len(axeStrings))
	for i, s := range axeStrings {
		axe[i], _ = strconv.Atoi(s)
	}

	for _, e := range cards {
		if !service.checkBombContains(axe, e) {
			bomb = append(bomb, e)
		}
	}

	b, _ := json.Marshal(bomb)
	r = strings.Trim(string(b), "[]")

	return r
}

func (service *eventService) GetEventDetails(eventId int64) models.Event {
	DB.Table("mini_game_event").Where("id = ?", eventId).Find(&service.event)
	return service.event
}

func (service *eventService) IsExist(eventId int64) int64 {
	var num int64
	DB.Table("mini_game_event").Where("id = ?", eventId).Count(&num)
	return num
}

func (service *eventService) GetCurrentEventByTableID(tableID int64, gameID int64) *models.Event {
	var res models.Event
	tm := time.Now()
	ftm := tm.Add(-time.Second * time.Duration(constants.LOL_TOWER_GAME_DURATION))

	// result := DB.Table("mini_game_event").
	// 	Where("mini_game_table_id = ?", tableID).
	// 	Where("status", constants.EVENT_STATUS_ENABLED).
	// 	Where("start_datetime >= ?", ftm).
	// 	First(&res)

	result := DB.Raw(`
		SELECT
			* 
		FROM
			"mini_game_event" 
		WHERE
			mini_game_table_id = ? 
			AND "status" = ?
			AND game_id = ?
			AND start_datetime >= ? 
		ORDER BY
			"mini_game_event"."start_datetime" 
			LIMIT 1`,
		tableID, constants.EVENT_STATUS_ENABLED, gameID, ftm).Scan(&res)

	if result.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &res
}

func (service *eventService) SettleUnsettledPrevEvents(event models.Event, tableID int64) {
	var result *gorm.DB
	sel := []string{"status", "mtime", "settlement_date"}
	sDate := time.Now()

	var eventCon = models.Event{
		Mtime:          time.Now(),
		Status:         constants.EVENT_STATUS_SETTLED,
		SettlementDate: &sDate,
	}
	result = DB.Table("mini_game_event").Select(sel).Where(
		"mini_game_table_id = ? AND start_datetime < ? AND status < ?",
		tableID, event.StartDatetime, constants.EVENT_STATUS_SETTLED).Updates(&eventCon)

	if result.Error != nil {
		logger.Error("Error UpdateEvent")
	}
}
