package gameloop_loltower

import (
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
	"time"

	betsimulator_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/betsimulator/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/cache"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gameloop"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	wsclient_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type Phase string

type MessageData struct {
	Bomb  string      `json:"bomb,omitempty"`
	Name  Phase       `json:"name,omitempty"`
	End   int64       `json:"end,omitempty"`
	Id    string      `json:"id,omitempty"`
	Level string      `json:"level,omitempty"`
	Users interface{} `json:"users,omitempty"`
}

type Message struct {
	Type Phase       `json:"type"`
	Data interface{} `json:"data"`
}

const (
	RESULT         Phase = "result"
	STATE          Phase = "state"
	BETTING        Phase = "BETTING"
	STOP_BETTING   Phase = "STOP_BETTING"
	SHOW_RESULT    Phase = "SHOW_RESULT"
	TOWER_CHAMPION Phase = "champion"
	MEMBER_LIST    Phase = "member_list"
)

var (
	broker                                           = wsclient_loltower.NewPublishBroker()
	icache         cache.ICache                      = cache.NewCache()
	eventService   service.EventService              = service.NewEvent()
	queueChampions types.Array[gamemanager.Champion] = make([]gamemanager.Champion, 0)
	betSimulator                                     = betsimulator_loltower.NewBetSimulator()
)

type oldlolTowerGameloop struct {
}

func (ds *oldlolTowerGameloop) GetIdentifier() string {
	return *types.String("loltower").Ptr()
}

func (ds *oldlolTowerGameloop) GetTableID() int64 {
	return constants_loltower.TableID
}

func (ds *oldlolTowerGameloop) GetDB() *gorm.DB {
	return db.Shared()
}

type selfGameloop struct {
	gl gameloop.Gameloop
}

func NewGameLoopOld() gameloop.Gamelooper {
	gl := gameloop.NewGameLoop(&oldlolTowerGameloop{})
	selfGameloop := selfGameloop{gl: gl}

	selfGameloop.Initialize()
	return &selfGameloop
}

func (sgl *selfGameloop) Start() {
	StartOLDGameloop() //old version - TODO: update to new if needed
}

func (sgl *selfGameloop) Stop() {
	sgl.gl.Stop()
}

func (sgl *selfGameloop) GetCurrentEvent() *models.Event {
	return sgl.gl.GetCurrentEvent()
}

func (sgl *selfGameloop) Initialize() {
}

func StartOLDGameloop() {
	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_loltower.Identifier)+"gameloop", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	for _, table := range constants.LOL_TOWER_TABLES {
		var lolTowerTable gamemanager.IGameManager = gamemanager.NewGameManager(table)
		go gameLoop(lolTowerTable, table)
	}
	// Block until a signal is received.
	s := <-c
	logger.Info("Got signal:", s)
}

func gameLoop(lolTowerTable gamemanager.IGameManager, table int64) {
	for {
		tableCacheKey := strconv.Itoa(int(table)) + "-current-event-status"
		var cacheEvent, _ = icache.GetOrig(tableCacheKey)

		if cacheEvent == "" {
			cEvent := lolTowerTable.GetCurrentEvent()
			if cEvent == nil {
				logger.Info("No Current Event.")
				continue
			}

			prepareCacheEvent(table, tableCacheKey, cEvent)
		} else {
			data := models.CacheEvent{}
			json.Unmarshal([]byte(cacheEvent), &data)

			// prevent nil pointer issue
			if data.EventID == nil {
				continue
			}

			if !data.IsBroadcastBetting {
				// send broadcast
				plusWaitTime := time.Duration(0) // no additional wait time for betting phase
				phaseDur := time.Duration(constants.LOL_TOWER_BETTING_DURATION)

				doSleep(data.Event, plusWaitTime, phaseDur)
				broadCastMessage(STATE, MessageData{
					Name: BETTING,
					End:  generateEndTime(data, phaseDur),
				})
				go betSimulator.StartBetting()
				// update event status to active
				data.Event.Status = constants.EVENT_STATUS_ACTIVE
				eventService.UpdateEventStatus(*data.EventID, *data.Event)

				// update  IsBroadcastBetting true
				data.IsBroadcastBetting = true
				p, _ := json.Marshal(data)
				// reset cache
				icache.Set(tableCacheKey, p, time.Second*constants.LOL_TOWER_GAME_DURATION)
			}

			if !data.IsBroadcastStopBetting {
				plusWaitTime := time.Duration(constants.LOL_TOWER_BETTING_DURATION) // plus BETTING duration
				phaseDur := time.Duration(constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION)

				doSleep(data.Event, plusWaitTime, phaseDur)
				broadCastMessage(STATE, MessageData{
					Name: STOP_BETTING,
					End:  generateEndTime(data, phaseDur),
				})
				go betSimulator.StartResulting()
				// update  IsBroadcastBetting true
				data.IsBroadcastStopBetting = true
				p, _ := json.Marshal(data)
				// reset cache
				icache.Set(tableCacheKey, p, time.Second*constants.LOL_TOWER_GAME_DURATION)
			}

			if !data.IsBroadcastShowResult {
				plusWaitTime := time.Duration(constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION) // plus BETTING and STOP_BETTING duration
				phaseDur := time.Duration(constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION + constants.LOL_TOWER_SHOW_RESULT_DURATION)

				doSleep(data.Event, plusWaitTime, phaseDur)
				broadCastMessage(STATE, MessageData{
					Name: SHOW_RESULT,
					End:  generateEndTime(data, phaseDur),
				})
				// update event status to for settlement
				data.Event.Status = constants.EVENT_STATUS_FOR_SETTLEMENT
				eventService.UpdateEventStatus(*data.EventID, *data.Event)
				go betSimulator.UpdateResults()
				go settleUnsettledPrevEvents(lolTowerTable, data)
				go settleEventTickets(lolTowerTable, data, table)
				// update  IsBroadcastBetting true
				data.IsBroadcastShowResult = true
				p, _ := json.Marshal(data)
				// reset cache
				icache.Set(tableCacheKey, p, time.Second*constants.LOL_TOWER_GAME_DURATION)
			}

			if !data.IsBroadcastShowBomb {
				plusWaitTime := time.Duration(constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION) // plus BETTING and STOP_BETTING duration
				phaseDur := time.Duration(constants.LOL_TOWER_BETTING_DURATION + constants.LOL_TOWER_STOP_BETTING_DURATION + constants.LOL_TOWER_SHOW_RESULT_DURATION)
				bomb := eventService.GetEventLolTowerBomb(*data.Event.ID)

				doSleep(data.Event, plusWaitTime, phaseDur)
				broadCastMessage(RESULT, MessageData{
					Bomb: bomb,
				})
				// update  IsBroadcastBetting true
				data.IsBroadcastShowBomb = true
				p, _ := json.Marshal(data)
				// reset cache
				icache.Set(tableCacheKey, p, time.Second*constants.LOL_TOWER_GAME_DURATION)
			}

			go retrieveLeaderboard(lolTowerTable, table)
			go retrieveChampionMembers(lolTowerTable, table)

			// if all done broadcast unset cache
			if data.IsBroadcastBetting && data.IsBroadcastStopBetting && data.IsBroadcastShowResult && data.IsBroadcastShowBomb {
				icache.Del(tableCacheKey)
			}
		}
	}
}

func settleUnsettledPrevEvents(lolTowerTable gamemanager.IGameManager, data models.CacheEvent) {
	go lolTowerTable.SettleUnsettledPrevEvents(*data.Event)
}

func settleEventTickets(lolTowerTable gamemanager.IGameManager, data models.CacheEvent, table int64) {
	go lolTowerTable.SettleTickets(*data.Event)
}

func prepareCacheEvent(table int64, tableCacheKey string, cEvent *models.Event) {
	toCacheEvent := models.CacheEvent{
		EventID:                cEvent.ID,
		Event:                  cEvent,
		IsBroadcastBetting:     false,
		IsBroadcastStopBetting: false,
		IsBroadcastShowResult:  false,
		IsBroadcastShowBomb:    false,
	}

	p, _ := json.Marshal(toCacheEvent)
	icache.Set(tableCacheKey, p, time.Second*constants.LOL_TOWER_GAME_DURATION)

	go cacheCurrentEvent(table, cEvent)
}

func cacheCurrentEvent(tableID int64, cEvent *models.Event) {
	timeNow := time.Now()
	waitTime := cEvent.StartDatetime.Sub(timeNow).Milliseconds()

	if waitTime > 0 {
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
	}

	cache.SaveEvent(tableID, cEvent)
}

func generateEndTime(data models.CacheEvent, phaseDur time.Duration) int64 {
	dur := phaseDur * time.Second

	return data.Event.StartDatetime.Add(dur).UnixNano() / int64(time.Millisecond)
}

func broadCastMessage(messageType Phase, messageData interface{}) {
	message := &Message{}
	message.Type = messageType
	message.Data = messageData
	payload, _ := json.Marshal(message)
	if msgData, ok := messageData.(MessageData); ok && messageType == STATE {
		redis.SetPublishState(constants_loltower.Identifier, &response.State{
			Name: string(msgData.Name),
			End:  float64(msgData.End) / 1_000,
			Gts:  utils.GenerateUnixTS(),
		})
	}
	broker.Publish(string(payload))
}

func doSleep(event *models.Event, plusWaitTime time.Duration, phaseDur time.Duration) {
	timeNow := time.Now()
	eTime := event.StartDatetime.Add(plusWaitTime * time.Second)
	waitTime := eTime.Sub(timeNow).Milliseconds()

	if waitTime > 0 {
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
	}
}

func retrieveLeaderboard(lolTowerTable gamemanager.IGameManager, tableID int64) {
	level, champions := lolTowerTable.GetLeaderboards()
	bsLevel, bsLeaderboard := betSimulator.GetLeaderBoard()

	if bsLevel > 0 {
		if bsLevel > int8(types.String(level).Int()) {
			champions = []gamemanager.Champion{} //override champions

			for _, bsl := range bsLeaderboard {

				champions = append(champions, gamemanager.Champion{
					Id:   bsl.EncryptedID(),
					Name: bsl.EncryptedName(),
				})
			}

			level = string(types.Int(bsLevel).String())
		} else if bsLevel == types.String(level).Int().Int8() {
			for _, bsl := range bsLeaderboard {

				champions = append(champions, gamemanager.Champion{
					Id:   bsl.EncryptedID(),
					Name: bsl.EncryptedName(),
				})
			}
		}
	}
	messageData := MessageData{Level: level, Users: champions}
	event, err := cache.GetEvent(tableID)

	if err != nil {
		logger.Error("fail to load event on retrieveLeaderboard error: ", err.Error())

	} else {
		eventID := event.ID

		//start - support new implementation for memberlist
		users := []response.User{}

		for i := 0; i < len(champions); i++ {
			users = append(users, response.User{ID: champions[i].Id, Name: champions[i].Name})
		}
		if err := redis.SetLeaderboard(constants_loltower.Identifier, &response.Leaderboard{Level: messageData.Level, Users: users}); err != nil {
			logger.Error("fail to set redis leaderboard error: ", err.Error())
		}
		//end - support new implementation for memberlist
		if err := cache.SaveLeaderboard(*eventID, messageData); err != nil {
			logger.Error("fail to save leaderboard error: ", err.Error())
		}
	}
	//FE requested to send empty leaderboards --> MGGL-102
	broadCastMessage(MEMBER_LIST, messageData)
}

func retrieveChampionMembers(lolTowerTable gamemanager.IGameManager, table int64) {
	champions, err := lolTowerTable.GetChampionMembers()
	bsChampios := betSimulator.GetChampions()

	for _, bsc := range bsChampios {
		champions = append(champions, gamemanager.Champion{
			Id:   bsc.EncryptedID(),
			Name: bsc.EncryptedName(),
		})
	}

	if err != nil {
		logger.Error("retrieveChampionMembers Error-------> %v", err.Error())
	}

	if len(champions) > 0 {
		tableCacheKey := strconv.Itoa(int(table)) + "-is-champion-messages-running"
		queueChampions = append(queueChampions, champions...)

		var isRunningMessage, _ = icache.GetOrig(tableCacheKey)

		if isRunningMessage == "" || isRunningMessage == "false" {
			go RunChampionsMessages(tableCacheKey)
		}
	}
}

func RunChampionsMessages(tableCacheKey string) {
	for len(queueChampions) > 0 {
		icache.Set(tableCacheKey, "true", time.Second*constants.LOL_TOWER_GAME_DURATION)

		messageData := queueChampions[0]
		broadCastMessage(TOWER_CHAMPION, messageData)

		// pop first element
		queueChampions.PopIndex(0)

		time.Sleep(4 * time.Second)
	}

	icache.Set(tableCacheKey, "false", time.Second*constants.LOL_TOWER_GAME_DURATION)
}
