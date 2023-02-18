package wsclient_loltower

import (
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
)

func NewPublishBroker() wsclient.MessagePublishBroker {
	return wsclient.NewMessagePublishBroker(constants_loltower.Identifier, constants_loltower.WebsocketChannel, constants_loltower.WebsocketRestartTimeout)
}

func NewReceiverBroker() wsclient.MessageReceiverBroker {
	return wsclient.NewMessageReceiverBroker(constants_loltower.Identifier, constants_loltower.WebsocketChannel, constants_loltower.WebsocketRestartTimeout)
}

func NewWSClient(broker wsclient.MessageReceiverBroker) wsclient.Client {
	return wsclient.NewWSClient(constants_loltower.Identifier, broker, func(handler wsclient.OnConnectHandler) {
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.StateType}))
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.TicketType}))
		handler.ProcessMessage(utils.JSON(map[string]any{"type": process.MemberListType}))
	})
}
