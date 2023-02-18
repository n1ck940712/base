package gamemanager

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/jinzhu/now"
	// "bitbucket.org/esportsph/minigame-backend-golang/internal/wsconsumer"
)

// this will contains any common functionality of GameManager
var (
	eventService       service.EventService        = service.NewEvent()
	hashService        service.HashService         = service.HashNew()
	selectionService   service.SelectionService    = service.NewSelection()
	gameService        service.GameService         = service.NewGame()
	memberTableService *service.MemberTableService = service.NewMemberTable()
)

type ReqTicket struct {
	EventID int64 `json:"event_id"`
	TableID int64 `json:"table_id"`
	Tickets []TicketDetails
}

type TicketDetails struct {
	Amount       float64 `json:"amount"`
	Selection    string  `json:"selection"`
	MarketType   int16   `json:"market_type"`
	ReferenceNo  string  `json:"reference_no"`
	HongkongOdds float64 `json:"hongkong_odds"`
	EuroOdds     float64 `json:"euro_odds"`
	Level        int16   `json:"level"`
}

type TicketSelection struct {
	ActiveTicket models.Ticket
	ComboTicket  models.ComboTicket
	Event        *models.Event
	Level        int16
	Skip         int8
	Selection    string
}

type CurrentEvent struct {
	ID            int64     `json:"id"`
	StartDatetime time.Time `json:"start_datetime"`
}

type Config struct {
	ID              int64        `json:"id"`
	MaxBetAmount    float64      `json:"max_bet_amount"`
	MaxPayoutAmount float64      `json:"max_payout_amount"`
	MinBetAmount    float64      `json:"min_bet_amount"`
	BetChips        *[]float64   `json:"bet_chips"`
	CurrentEvent    CurrentEvent `json:"current_event"`
	EffectsSound    float64      `json:"effects_sound"`
	GameSound       float64      `json:"game_sound"`
	IsAnonymous     bool         `json:"is_anonymous"`
	Enable          bool         `json:"enable"`
	ResultAnimation bool         `json:"result_animation"`
	ShowCharts      string       `json:"showCharts"`
	EnableAutoPlay  bool         `json:"enable_auto_play"`
	Tour            bool         `json:"tour"`
}

func createFutureEvents(gameID int64, tableID int64, maxNumEvents int) bool {
	params := map[string]interface{}{
		"status":             constants.EVENT_STATUS_ENABLED, // get enabled events
		"mini_game_table_id": tableID,
	}

	fEvents, _ := eventService.GetFutureEvents(params)
	lackingFutureEventCount := (maxNumEvents - len(fEvents))

	// get available events assure 3 events available [status enabled]
	for i := 0; i < lackingFutureEventCount; i++ {
		var prevEventID *int64
		tm := time.Now()
		startTm := tm

		lastEvent := eventService.GetLastEventByTableID(tableID)

		if lastEvent != nil {
			prevEventID = lastEvent.ID
			tempTm := lastEvent.StartDatetime.Add(time.Second * time.Duration(constants.LOL_TOWER_GAME_DURATION))
			diff := tm.Sub(tempTm).Seconds()

			if diff < float64(constants.LOL_TOWER_GAME_DURATION) {
				startTm = lastEvent.StartDatetime.Add(time.Second * time.Duration(constants.LOL_TOWER_GAME_DURATION))
			}
		}

		// get next sequence
		nhs := getNextHashSeq(tableID)
		if nhs.Value == "" {
			continue
		}

		result, selectionHeaderID := generateHashResult(nhs.Value, gameID)

		// get table info
		table := gameService.GetTable(tableID)

		payload := models.Event{
			LocalDataVersion:   1,
			ESDataVersion:      0,
			SyncStatus:         0,
			Ctime:              tm,
			Mtime:              tm,
			Name:               table.Name,
			IsAutoPlayExecuted: false,
			StartDatetime:      startTm,
			Status:             0,
			MaxBet:             table.MaxBetAmount,
			MaxPayout:          table.MaxPayoutAmount,
			GameID:             gameID,
			HashSequenceID:     nhs.ID, // get available Game hash sequence id
			SelectionHeaderID:  selectionHeaderID,
			GroundType:         constants.DEFAULT_GROUND_TYPE,
			TableID:            tableID,
			PrevEventID:        prevEventID,
		}

		createdEvent, err := eventService.CreateEvents(payload)
		if err != nil {
			logger.Warning("Event Not Created!")
			continue
		}

		// create event result
		var res models.EventResult
		res.ResultType = 28
		res.EventID = createdEvent.ID
		res.Value = result

		eventService.CreateEventResult(res)
	}

	return true
}

func getNextHashSeq(tableID int64) models.HashSequence {
	var nextSeq models.HashSequence

	temp := map[string]interface{}{
		"mini_game_table_id": tableID,
	}
	lastSeq := eventService.GetLatestHashSeq(temp)
	maxHQ := settings.DEF_MAX_SEQUENCE

	// no last sequence used
	if lastSeq.ID == 0 && lastSeq.Sequence == 0 {
		nextSeq = hashService.GetNextQueuedHashSequence(tableID)
		//set to active
		hashService.SetHashStatus(nextSeq.HashID, "active")
	} else if lastSeq.Sequence == maxHQ { // last sequence = to max seq use queued hash and update hash to done
		// update hash to done
		hashService.SetHashStatus(lastSeq.HashID, "done")
		// get next hash header set to active and get next sequence
		nextSeq = hashService.GetNextQueuedHashSequence(tableID)
		// set header to active
		hashService.SetHashStatus(nextSeq.HashID, "active")
	} else { // take next sequence
		isActive := hashService.IsActiveHash(lastSeq.HashID)
		if isActive {
			nextSeq = hashService.GetNextHashSequence(tableID, lastSeq.Sequence, lastSeq.HashID)
		} else {
			nextSeq = hashService.GetNextQueuedHashSequence(tableID)
			//set to active
			hashService.SetHashStatus(nextSeq.HashID, "active")
		}
	}

	return nextSeq
}

func (conf *Config) setMemberConfig(gameID int64, userID int64) {
	memberConfig, _ := iMemberConfig.GetMemberConfig(gameID, userID)

	data := make(map[string]interface{})
	for _, v := range memberConfig {
		confType := constants.MEMBER_CONFIGS[v.Name]
		var interfaceValue interface{} = v.Value
		switch confType {
		case "string":
			data[v.Name] = interfaceValue.(string)
		case "json_string":
			data[v.Name] = strings.ReplaceAll(interfaceValue.(string), `"`, "'")
		case "bool":
			boolValue, _ := strconv.ParseBool(interfaceValue.(string))
			data[v.Name] = boolValue
		case "float64":
			data[v.Name] = types.String(interfaceValue.(string)).Float().Float64()
		}
	}

	jsonString, _ := json.Marshal(data)
	json.Unmarshal(jsonString, &conf)
}

func (l *LolTowerGameManager) UpsertConfig(userID int64, payload models.ConfigPatchRequestBody) (models.ConfigPatchRequestBody, error) {
	userDetails, _ := iUserService.GetMGDetails(userID)

	jsonString, _ := json.Marshal(payload)
	jsonMap := make(map[string]interface{})
	json.Unmarshal([]byte(jsonString), &jsonMap)

	var memberConfigs []models.MemberConfig

	for k, v := range jsonMap {
		strVal := fmt.Sprintf("%v", v)
		config := models.MemberConfig{
			Name:   k,
			Value:  strVal,
			UserID: userDetails.ID,
			GameId: l.gameID,
		}
		memberConfigs = append(memberConfigs, config)
	}

	res, err := iMemberConfig.BatchUpsertMemberConfig(memberConfigs)
	logger.Debug(res)
	return payload, err
}

func (l *LolTowerGameManager) ProcessDailyMaxWinnings() {
	miniGameTable := models.GameTable{
		ID:     l.tableID,
		GameID: l.gameID,
	}

	service.Get(&miniGameTable)

	checkMaxDailyWinnings := miniGameTable.DisableOnDailyMaxWinnings
	if checkMaxDailyWinnings == nil {
		return
	}
	maxDailyWinnings := *miniGameTable.DisableOnDailyMaxWinnings
	lastTriggered := miniGameTable.LastDailyMaxWinningsTriggeredDatetime

	var nowStart = now.BeginningOfDay()
	var nowEnd = now.EndOfDay()

	if lastTriggered != nil && lastTriggered.After(nowStart) {
		nowStart = *lastTriggered
	}

	total := iTicketService.GetTotalWinlossAmount(nowStart, nowEnd, l.tableID)
	logger.Info("TOTAL: ", total, (total >= maxDailyWinnings))
	if total >= maxDailyWinnings {
		var updateMaxWinningTriggeredDateTime = time.Now()

		var toUpdate = models.GameTable{
			ID:                                    l.tableID,
			GameID:                                l.gameID,
			LastDailyMaxWinningsTriggeredDatetime: &updateMaxWinningTriggeredDateTime,
			DisabledByMgTrigger:                   true,
		}

		service.Update(&toUpdate)
	}
}
