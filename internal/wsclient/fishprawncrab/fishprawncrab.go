package wsclient_fishprawncrab

import (
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
)

func NewPublishBroker() wsclient.MessagePublishBroker {
	return wsclient.NewMessagePublishBroker(constants_fishprawncrab.Identifier, constants_fishprawncrab.WebsocketChannel, constants_fishprawncrab.WebsocketRestartTimeout)
}

func NewReceiverBroker() wsclient.MessageReceiverBroker {
	return wsclient.NewMessageReceiverBroker(constants_fishprawncrab.Identifier, constants_fishprawncrab.WebsocketChannel, constants_fishprawncrab.WebsocketRestartTimeout)
}

func NewWSClient(broker wsclient.MessageReceiverBroker) wsclient.Client {
	return wsclient.NewWSClient(constants_fishprawncrab.Identifier, broker, func(handler wsclient.OnConnectHandler) {
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.StateType}))
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.TicketType}))
	})
}
