package wsconsumer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/cache"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	e "bitbucket.org/esportsph/minigame-backend-golang/internal/errors"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/placebet"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

const TABLE_UPDATE_INTERVAL_SEC = 60

var (
	upgrader                         = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	icache                           = cache.NewCache()
	iUserService service.UserService = *service.NewUser()
	consumers                        = make(map[int64]string)
)

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Cts  time.Time   `json:"cts"`
}

type Consumer struct {
	broker          wsclient.MessageReceiverBroker
	brokerChan      chan string
	stop            bool
	conn            *websocket.Conn
	token           string
	user            *models.User
	tableID         int64
	gameID          int
	expirationTS    int64
	placeBet        placebet.IPlacebet
	connID          *string
	lastTableUpdate time.Time
}

func (consumer *Consumer) newConnID() string {
	connID := consumer.token
	consumers[consumer.user.EsportsID] = connID
	consumer.connID = &connID

	return connID
}

func NewConsumer(broker wsclient.MessageReceiverBroker, gameID int, tableID int64) *Consumer {
	return &Consumer{
		broker,
		make(chan string),
		false,
		nil,
		"",
		nil,
		tableID,
		gameID,
		0,
		placebet.NewPlacebet(),
		nil,
		time.Now(),
	}
}

func (consumer *Consumer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("HandleRequest Error: ", err.Error())
		http.Error(w, "Unable to upgrade connection", 500)
		return
	}
	defer conn.Close()
	consumer.conn = conn
	if err := consumer.authenticate(r); err != nil {
		logger.Error("Unable to authenticate " + err.Error())
		consumer.SendErrorMessage("Invalid token")
		return
	}
	go consumer.updateTable()
	consumer.newConnID()
	consumer.sendLoginMessage()
	consumer.sendTicketState()
	consumer.sendMemberList()
	consumer.run()

	// delete consumers key when disconnected
	delete(consumers, consumer.user.EsportsID)
	logger.Debug("Closing websocket connection with token: \"" + consumer.token + "\"")
}

func (consumer *Consumer) ClientRead(clientChan chan<- string, conn *websocket.Conn) {
	defer close(clientChan)

	for !consumer.stop {
		mt, message, err := conn.ReadMessage()
		if err != nil || err == io.EOF {
			logger.Error("Client Read Error: ", err.Error())
			break
		}
		if mt != websocket.TextMessage {
			logger.Error("Message is not text type")
			break
		}
		clientChan <- string(message[:])
	}
	consumer.stop = true
}

func getDefaultErrMsg() WSMessage {
	//## TODO - actual error message from DB
	msgDetails := map[string]interface{}{ // Sample Error message
		"eid":  1,
		"type": "error",
		"ieid": 0,
		"msg":  "invalid type",
		"mid":  0,
	}

	errMsg := WSMessage{
		Type: "error",
		Data: msgDetails,
	}
	return errMsg
}

func (consumer *Consumer) OnReceive(msg WSMessage) {
	if *consumer.connID != consumers[consumer.user.EsportsID] {
		consumer.SendErrorMessage("Invalid token")
		consumer.conn.Close()
		return
	}

	if !consumer.IsValidMessage(msg) {
		errMsg := getDefaultErrMsg()
		consumer.SendMessage(errMsg)
	} else {
		msg := consumer.processReceivedMsg(msg)
		consumer.SendMessage(msg)
	}
}

func (consumer *Consumer) processReceivedMsg(msg WSMessage) WSMessage {
	go consumer.updateTable()

	var wsMsg WSMessage
	var res interface{}
	var err e.FinalErrorMessage
	temp := map[string]interface{}{
		"gameID":  consumer.gameID,
		"tableID": consumer.tableID,
		"user":    consumer.user,
		"data":    msg.Data,
	}

	switch msg.Type {
	case "state":
	case "selection":
		if consumer.placeBet.Lock() {
			res = nil
			err = e.FinalizeErrorMessage(e.TICKET_CREATION_ERROR, e.IEID_BET_PROCESS_ONGOING, false)
		} else {
			res, err = consumer.placeBet.ProcessSelection(temp)
		}
		consumer.placeBet.Unlock()
	case "bet":
		if consumer.placeBet.Lock() {
			res = nil
			err = e.FinalizeErrorMessage(e.TICKET_CREATION_ERROR, e.IEID_BET_PROCESS_ONGOING, false)
		} else {
			res, err = consumer.placeBet.ProcessTicket(temp)
		}
		consumer.placeBet.Unlock()
	case "odds":
		res, err = consumer.getGameOdds()
	case "config":
		res, err = consumer.getConfig()
	case "ticket", "current-tickets":
		res, err = consumer.getTicketState()
	case "test":
		res, err = consumer.getTestReplyMsg(temp)
	default:
		msg.Type = "state"
		res, err = consumer.getDefaultReplyMsg()
	}
	data, msgType := validateMsgResponse(msg.Type, res, err)
	wsMsg = WSMessage{Type: msgType, Data: data}

	return wsMsg
}

func (consumer *Consumer) getConfig() (gamemanager.Config, e.FinalErrorMessage) {
	var gameManager gamemanager.IGameManager = gamemanager.NewGameManager(consumer.tableID)
	res, err := gameManager.GetTableConfig(consumer.user.EsportsID)

	return res, err
}

func (consumer *Consumer) getGameOdds() (map[string]interface{}, e.FinalErrorMessage) {
	var gameManager gamemanager.IGameManager = gamemanager.NewGameManager(consumer.tableID)
	res := gameManager.GetOdds()

	return res, nil
}

func (consumer *Consumer) getTicketState() (*service.TicketState, e.FinalErrorMessage) {
	var gameManager gamemanager.IGameManager = gamemanager.NewGameManager(consumer.tableID)
	res, err := gameManager.GetMemberTicketState(consumer.user.EsportsID)

	// no combo bet found
	if res == nil || err == gorm.ErrRecordNotFound {
		res = &service.TicketState{
			BetAmount: nil,
			Level:     0,
			Skip:      int(constants.LOL_TOWER_SKIP_COUNT),
			Selection: nil,
		}
	}

	if res.BetAmount != nil {
		memberTable := models.MemberTable{UserID: consumer.user.ID, TableID: consumer.tableID}

		if err := service.Get(&memberTable); err != nil {
			logger.Error("getTicketState memberTable error: ", err.Error())
		}
		if memberTable.IsEnabled {
			maxPayoutLevel := int8(8)

			for level := 1; level <= constants.LOL_TOWER_MAX_LEVEL; level++ {
				maxPossibleLevel := constants.LOL_TOWER_MAX_LEVEL - constants.LOL_TOWER_SKIP_COUNT + int8(res.Skip)

				if level > int(maxPossibleLevel) {
					break
				}
				euroOdds := settings.LOL_LEVELS[level]

				if utils.CalculateMaxPayout(*res.BetAmount, euroOdds) > memberTable.MaxPayoutAmount { //if calculatedPayout is higher than maxPayoutAmount exit range
					break
				}
				maxPayoutLevel = int8(level)
			}

			res.MaxPayoutLevel = &maxPayoutLevel
		}
	}

	return res, nil
}

func validateMsgResponse(mType string, data interface{}, errors e.FinalErrorMessage) (interface{}, string) {

	if errors != nil {
		temp := models.ResponseData{
			Type:    mType,
			Ied:     errors.Ied(),
			Ieid:    errors.Ieid(),
			Message: errors.Message(),
			Mid:     "",
		}

		return temp, "error"
	}

	return data, mType

}

func (consumer *Consumer) getTestReplyMsg(temp map[string]interface{}) (interface{}, e.FinalErrorMessage) {
	return map[string]interface{}{
		"message": temp["data"],
	}, nil
}

func (consumer *Consumer) getDefaultReplyMsg() (interface{}, e.FinalErrorMessage) {
	msgDetails := map[string]interface{}{ // Sample Reply message
		"name": "Betting",
		"time": time.Now(),
	}

	return msgDetails, nil
}

func (consumer *Consumer) SendMessage(msg WSMessage) {
	msg.Cts = time.Now()
	msgByte, err := json.Marshal(msg)
	if err != nil {
		logger.Debug("Failed to encode message")
	}

	if err2 := consumer.conn.WriteMessage(websocket.TextMessage, msgByte); err != nil {
		logger.Error("WriteMessage: ", err2)
		return
	}
}

func (consumer *Consumer) SendErrorMessage(errorMsg string) {
	var msg = `{"type": "error", "data": {"msg": "` + errorMsg + `"}}`
	consumer.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (consumer *Consumer) IsValidMessage(msg WSMessage) bool {
	rType := map[string]bool{
		"state":     true,
		"bet":       true,
		"selection": true,
		"odds":      true,
		"config":    true,
		"ticket":    true,
		"test":      true,
	}

	if !rType[msg.Type] || len(msg.Type) == 0 {
		return false
	}

	return true
}

func (consumer *Consumer) authenticate(r *http.Request) error {
	var authToken = r.URL.Query().Get("auth_token")
	if authToken == "" {
		return errors.New("Token not found.")
	}
	logger.Debug("Connected with auth_token \"" + authToken + "\"")
	consumer.token = authToken
	esUser, err := consumer.callApiValidateToken()

	if err != nil {
		icache.Touch("auth-token", nil)
		icache.Touch(authToken, nil)
		return err
	} else {
		logger.Debug("SET TOKEN 2 mins expiration!")
		user, err := iUserService.GetMGDetails(esUser.ID)

		if err != nil {
			icache.Touch("auth-token", nil)
			icache.Touch(authToken, nil)
			logger.Error("MG User not found")
			return errors.New("User not found!")
		}
		user.SetRequest(r)
		consumer.user = &user
	}
	consumer.resetConnectionExpiration()
	return nil
}

func (c *Consumer) callApiValidateToken() (*models.ESUser, error) {
	var esUser models.ESUser

	if err := api.NewAPI(settings.EBO_API + "/v1/validate-token/").
		SetIdentifier("callApiValidateToken").
		AddHeaders(map[string]string{
			"User-Agent":    settings.USER_AGENT,
			"Authorization": settings.SERVER_TOKEN,
			"Content-Type":  "application/json",
		}).
		AddBody(map[string]string{
			"token": c.token,
		}).
		Post(&esUser); err != nil {
		return nil, err
	}

	userByte, err := json.Marshal(esUser) // Set User struct
	if err != nil {
		logger.Error("Unable to marshal user from login")
		return nil, err
	}
	icache.Set(c.token, userByte, settings.GetTimeout("timeout"))
	icache.Set("auth-token", c.token, settings.GetTimeout("timeout"))
	return &esUser, nil
}

func (consumer *Consumer) run() {
	consumer.broker.Subscribe(consumer.brokerChan)
	defer consumer.broker.Unsubcribe(consumer.brokerChan)

	var clientChan = make(chan string)
	go consumer.ClientRead(clientChan, consumer.conn)

	var exitLoop = false

	for !exitLoop && !consumer.stop {

		if consumer.isConnectionExpired() {
			exitLoop = true
		}

		select {
		case msg, ok := <-clientChan:
			if !ok {
				exitLoop = true
				break
			}

			var m WSMessage
			err := json.Unmarshal([]byte(msg), &m)

			if err != nil {
				errMsg := getDefaultErrMsg()
				consumer.SendMessage(errMsg)
			} else {
				consumer.OnReceive(m)
			}
		case msg, ok := <-consumer.brokerChan:
			if !ok {
				exitLoop = true
				break
			}

			var m WSMessage
			err := json.Unmarshal([]byte(msg), &m)
			if err != nil {
				m = getDefaultErrMsg()
			}
			consumer.SendMessage(m)
		}
	}
	consumer.stop = true
}

// set consumer expiration
func (consumer *Consumer) resetConnectionExpiration() {
	consumer.expirationTS = time.Now().Unix() + 60 //added 60s TODO: add to constants
}

// invalidate expiration and return if invalidated
func (consumer *Consumer) isConnectionExpired() bool {
	expirationTS := time.Now().Unix()

	if expirationTS > consumer.expirationTS {
		_, err := consumer.callApiValidateToken()

		if err != nil {
			return true
		}
		consumer.resetConnectionExpiration()
		return false
	}
	return false
}

func (consumer *Consumer) sendLoginMessage() {
	if settings.ENVIRONMENT == "live" { //disable login message on live
		return
	}
	isAnonymous := false
	loginMessage := models.JsonResponse{
		IsAnonymous: &isAnonymous,
		MemberCode:  consumer.user.MemberCode,
		QueryParams: &models.QueryParams{
			AuthToken: []string{consumer.token},
		},
		UserID: types.Bytes(fmt.Sprintf("%v%v", consumer.user.EsportsID, consumer.user.EsportsPartnerID)).SHA256(),
		Cts:    float64(time.Now().UnixMicro()) / 1_000_000,
	}

	// 1654156291
	msgByte, _ := json.Marshal(loginMessage)
	consumer.conn.WriteMessage(websocket.TextMessage, msgByte)
}

func (consumer *Consumer) sendTicketState() {
	//added current state
	if state, err := redis.GetPublishState(constants_loltower.Identifier); err != nil {
		logger.Info(constants_loltower.Identifier, " GetPublishState error: ", err.Error())
	} else if state != nil {
		msg := (&response.Response{
			Type: "state",
			Data: state,
		}).JSON()
		msg = fmt.Sprintf(strings.Replace(msg, `"%f"`, "%f", 1), utils.GenerateUnixTS()) //replace %f with current timestamp
		consumer.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}

	res, err := consumer.getTicketState()
	data, msgType := validateMsgResponse("ticket", res, err)
	wsMsg := WSMessage{
		Type: msgType,
		Data: data,
	}

	consumer.SendMessage(wsMsg)
}

func (consumer *Consumer) sendMemberList() {
	event, err := cache.GetEvent(consumer.tableID)

	if err != nil {
		logger.Error("fail to load event on consumer sendMemberList error: ", err.Error())
	} else {
		eventID := event.ID
		prevEventID := event.PrevEventID
		messageData := map[string]any{}
		if err := cache.GetLeaderboard(*eventID, *prevEventID, &messageData); err != nil {
			logger.Error("fail to get leaderboard error: ", err.Error())
		}

		wsMsg := WSMessage{
			Type: "member_list",
			Data: messageData,
		}

		consumer.SendMessage(wsMsg)
	}
}

func (consumer *Consumer) updateTable() error {
	if time.Since(consumer.lastTableUpdate).Seconds() < TABLE_UPDATE_INTERVAL_SEC {
		return nil
	}

	if err := api.NewAPI(settings.MG_CORE_API + "/game-client/mini-game/v2/tables/").
		SetIdentifier("callApiUpdateTable").
		AddHeaders(map[string]string{
			"User-Agent":    settings.USER_AGENT,
			"Authorization": "Token " + consumer.token,
			"Content-Type":  "application/json",
		}).
		GetIgnoreResponse(); err != nil {
		return err
	}

	consumer.lastTableUpdate = time.Now()
	return nil
}
