package wsclient_lolcouple

import (
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
)

func NewPublishBroker() wsclient.MessagePublishBroker {
	return wsclient.NewMessagePublishBroker(constants_lolcouple.Identifier, constants_lolcouple.WebsocketChannel, constants_lolcouple.WebsocketRestartTimeout)
}

func NewReceiverBroker() wsclient.MessageReceiverBroker {
	return wsclient.NewMessageReceiverBroker(constants_lolcouple.Identifier, constants_lolcouple.WebsocketChannel, constants_lolcouple.WebsocketRestartTimeout)
}

func NewWSClient(broker wsclient.MessageReceiverBroker) wsclient.Client {
	return wsclient.NewWSClient(constants_lolcouple.Identifier, broker, func(handler wsclient.OnConnectHandler) {
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.StateType}))
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.TicketType}))
	})
}
