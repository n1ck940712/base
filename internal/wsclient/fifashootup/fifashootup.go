package wsclient_fifashootup

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
)

func NewPublishBroker() wsclient.MessagePublishBroker {
	return wsclient.NewMessagePublishBroker(constants_fifashootup.Identifier, constants_fifashootup.WebsocketChannel, constants_fifashootup.WebsocketRestartTimeout)
}

func NewReceiverBroker() wsclient.MessageReceiverBroker {
	return wsclient.NewMessageReceiverBroker(constants_fifashootup.Identifier, constants_fifashootup.WebsocketChannel, constants_fifashootup.WebsocketRestartTimeout)
}

func NewWSClient(broker wsclient.MessageReceiverBroker) wsclient.Client {
	return wsclient.NewWSClient(constants_fifashootup.Identifier, broker, func(handler wsclient.OnConnectHandler) {
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.StateType}))
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.TicketType}))
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.GameDataType}))
	})
}
