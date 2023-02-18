package wsclient

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/mutex"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

var (
	consumersConnData = mutex.NewData(make(map[int64]string))
)

type Client interface {
	Serve(w http.ResponseWriter, r *http.Request)
}

type wsclient struct {
	identifier              string
	broker                  MessageReceiverBroker
	ws                      WS
	user                    *models.User
	processRequest          process.ProcessRequest
	connID                  *string
	activityExpirationTS    int64 //when expired dc client - no activity while connected
	updateTableExpirationTS int64 //wheb expired can call UpdateTableIfNeeded
	onConnectCallback       onConnectCallback
}

func NewWSClient(identifier string, broker MessageReceiverBroker, onConnectCallback onConnectCallback) Client {
	return &wsclient{
		identifier:        identifier,
		broker:            broker,
		onConnectCallback: onConnectCallback,
	}
}

func (wsc *wsclient) Serve(w http.ResponseWriter, r *http.Request) {
	wsc.ws = NewWS(w, r)

	if wsc.ws.GetError() != nil {
		logger.Error(wsc.identifier, " upgrader upgrade error: ", wsc.ws.GetError().Error())
		return
	}
	defer wsc.ws.Close()
	if err := wsc.Authenticate(); err != nil {
		logger.Error(wsc.identifier, " authenticate error: ", err.Error())
		wsc.SendErrorMessage("authenticate", "invalid token")
		return
	}
	wsc.UpdateTableIfNeeded()
	wsc.OnConnectMessages()
	wsc.Run()
}

func (wsc *wsclient) SendErrorMessage(errorType string, msg string) {
	response := process.CreateResponse("error", response.ErrorWithMessage(errorType, msg))

	wsc.ws.WriteMessage(response.JSON())
}

func (wsc *wsclient) Run() {
	wsc.ConnectConnID()
	defer wsc.DisconnectConnID()
	sigTermChan := make(chan os.Signal, 1)
	closeChan := make(chan bool)
	clientWriteStop := false
	clientChan := make(chan string)
	brokerReadStop := false
	brokerChan := make(chan string)

	signal.Notify(sigTermChan, syscall.SIGINT, syscall.SIGTERM)
	wsc.broker.Subscribe(brokerChan) //subscribe for write channel
	defer close(closeChan)
	defer close(clientChan)
	defer close(brokerChan)
	defer wsc.broker.Unsubcribe(brokerChan) //unsubscribe for write channel
	defer close(sigTermChan)
	go wsc.RunClientWrite(&clientWriteStop, clientChan, closeChan) //write client channel from read message
	go wsc.RunBrokerRead(&brokerReadStop, brokerChan, closeChan)   //read channel and send message

client_read:
	for {
		select {
		case sig, ok := <-sigTermChan:
			if !ok {
				break client_read
			}
			logger.Info(wsc.identifier, " terminate signal (", sig, ") received, exiting...")
			break client_read
		case <-closeChan:
			break client_read
		case msg, ok := <-clientChan:

			if !ok {
				break client_read
			}
			if err := wsc.Authenticate(); err != nil {
				wsc.SendErrorMessage("authenticate", "invalid token")
				break client_read
			}
			if !wsc.IsConnIDValid() {
				wsc.SendErrorMessage("authenticate", "invalid token")
				break client_read
			}
			wsc.UpdateTableIfNeeded()
			response := wsc.GetProcess().ProcessRequest(msg)

			wsc.ws.WriteMessage(response.JSON())
		}
	}
	clientWriteStop = true
	brokerReadStop = true
	logger.Info(wsc.identifier, " client read exited")
}

func (wsc *wsclient) RunClientWrite(writeStop *bool, clientChan chan<- string, closeChan chan<- bool) {
client_write:
	for !*writeStop {
		msg, err := wsc.ws.ReadMessage()

		if *writeStop {
			logger.Info(wsc.identifier, " client write stopped")
			break client_write
		}
		if err != nil {
			logger.Error(wsc.identifier, " client write read error: ", err.Error())
			closeChan <- true
			break client_write
		}

		if msg != "" {
			clientChan <- msg
		}
	}
	*writeStop = true
	logger.Info(wsc.identifier, " client write exited")
}

func (wsc *wsclient) RunBrokerRead(readStop *bool, brokerChan <-chan string, closeChan chan<- bool) {
	for !*readStop {
		msg, ok := <-brokerChan

		if !ok {
			break
		}
		if *readStop {
			logger.Info(wsc.identifier, " broker read stopped")
			break
		}
		if !wsc.IsConnIDValid() {
			wsc.SendErrorMessage("authenticate", "invalid token")
			closeChan <- true
			break
		}
		if wsc.IsActivityExpired() {
			wsc.SendErrorMessage("authenticate", "no activity for 60s")
			closeChan <- true
			break
		}
		wsc.ws.WriteMessage(msg)

	}
	*readStop = true
	logger.Info(wsc.identifier, " broker read exited")
}

func (wsc *wsclient) OnConnectMessages() {
	if settings.ENVIRONMENT != "live" {
		login := response.Response{
			UserID:      types.Bytes(fmt.Sprintf("%v%v", wsc.user.EsportsID, wsc.user.EsportsPartnerID)).SHA256(),
			IsAnonymous: types.Bool(false).Ptr(),
			MemberCode:  wsc.user.MemberCode,
			QueryParams: &response.GenericMap{
				"auth_token": []string{
					*wsc.user.AuthToken,
				},
			},
		}
		wsc.ws.WriteMessage(login.JSON())
	}
	wsc.onConnectCallback(wsc) //trigger on connect
}

func (wsc *wsclient) ConnectConnID() {
	connID := *wsc.user.AuthToken
	consumersConnData.Lock()
	defer consumersConnData.Unlock()
	consumersConnData.Data[wsc.user.EsportsID] = connID //overwrite connID
	wsc.connID = &connID
}

func (wsc *wsclient) DisconnectConnID() {
	consumersConnData.Lock()
	defer consumersConnData.Unlock()
	if connID, ok := consumersConnData.Data[wsc.user.EsportsID]; ok { //existing key
		if *wsc.connID == connID { //connID with equal disconnect
			delete(consumersConnData.Data, wsc.user.EsportsID)
		}
	}
}

func (wsc *wsclient) IsConnIDValid() bool {
	consumersConnData.Lock()
	defer consumersConnData.Unlock()
	if connID, ok := consumersConnData.Data[wsc.user.EsportsID]; ok {
		return connID == *wsc.connID
	}
	return false
}
