package wsclient

import (
	"context"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	_redis "bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/go-redis/redis/v8"
)

type MessagePublishBroker interface {
	GetIdentifier() string
	Publish(msg string)
}
type MessageReceiverBroker interface {
	GetIdentifier() string
	Subscribe(c chan<- string)
	Unsubcribe(c chan<- string)
	Start()
	Stop()
}

var ctx = context.Background()

type messageBroker struct {
	identifier  string
	channel     string
	timeoutMS   int64
	subscribers map[chan<- string]chan<- string
	closeChan   chan uint
	sendCount   int64
	sentCount   int64
	redis       _redis.Redis
	pubSub      *redis.PubSub
	retries     int8
}

func NewMessagePublishBroker(identifier string, channel string, timeoutMS int64) MessagePublishBroker {
	broker := messageBroker{
		identifier: identifier,
		channel:    channel,
		timeoutMS:  timeoutMS,
	}

	broker.ResetPublishTimeout()
	return &broker
}

func NewMessageReceiverBroker(identifier string, channel string, timeoutMS int64) MessageReceiverBroker {
	broker := messageBroker{
		identifier:  identifier,
		channel:     channel,
		subscribers: map[chan<- string]chan<- string{},
		timeoutMS:   timeoutMS,
		closeChan:   make(chan uint),
	}

	broker.ResetRecieverTimeout()
	return &broker
}

func (broker *messageBroker) GetIdentifier() string {
	return string(broker.identifier + " messageBroker")
}

// Publish
func (broker *messageBroker) Publish(msg string) {
	if broker.ShouldRestartPublish() {
		logger.Info(broker.GetIdentifier(), " restart client on publish")
		broker.ClientClose()
		broker.ResetPublishTimeout()
		go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(string(broker.identifier))+"gameloop", "> *RESTARTING PUBLISH BROKER*"), slack.LootboxHealthCheck)
	}
	go broker.PublishToWS(msg)
	if err := broker.GetClient().Publish(ctx, broker.channel, msg).Err(); err != nil {
		logger.Error(broker.GetIdentifier(), " publish error: ", err.Error())
	}
}

func (broker *messageBroker) ResetPublishTimeout() {
	key := broker.channel + "-publish-restart-timeout"

	timeOut := time.Now().UnixMilli() + broker.timeoutMS
	_redis.Cache().Set(key, *types.Int(timeOut).String().Ptr(), time.Duration(broker.timeoutMS)*time.Millisecond)
}

func (broker *messageBroker) ShouldRestartPublish() bool {
	key := broker.channel + "-publish-restart-timeout"
	cTS, err := _redis.Cache().Get(key)

	if err != nil {
		logger.Info(broker.GetIdentifier(), " ShouldRestartPublish key: (", cTS, ")")
		logger.Error(broker.GetIdentifier(), " ShouldRestartPublish error: ", err.Error())
	}

	if cTS == "" {
		return true
	}
	return types.String(cTS).Int().Int64() < time.Now().UnixMilli()
}

func (broker *messageBroker) PublishToWS(jsonMsg string) {
	if err := api.NewAPI(settings.GetMGWSBaseAPIURL().String() + "v2/broadcast/" + broker.Slug() + "/").
		SetIdentifier(broker.GetIdentifier() + " PublishToWS").
		AddHeaders(map[string]string{
			"User-Agent":    settings.GetUserAgent().String(),
			"Authorization": settings.GetServerToken().String(),
			"Content-Type":  "application/json",
		}).
		AddBody(jsonMsg).
		Post(nil); err != nil {
		logger.Error(broker.GetIdentifier()+" PublishToWS", err.Error())
	}
}

func (broker *messageBroker) Slug() string {
	switch broker.identifier {
	case constants_loltower.Identifier:
		return constants_loltower.WebsocketSlug
	case constants_lolcouple.Identifier:
		return constants_lolcouple.WebsocketSlug
	case constants_fifashootup.Identifier:
		return constants_fifashootup.WebsocketSlug
	case constants_fishprawncrab.Identifier:
		return constants_fishprawncrab.WebsocketSlug
	default:
		panic("unsupported broker identifier: " + broker.identifier)
	}
}

// Receiver
func (broker *messageBroker) Subscribe(sub chan<- string) {
	broker.subscribers[sub] = sub
}

func (broker *messageBroker) Unsubcribe(sub chan<- string) {
	delete(broker.subscribers, sub)
}

func (broker *messageBroker) Start() {
	stopRestarter := false
	restarted := false

	go func() {
		for !stopRestarter {
			if broker.ShouldRestartReciever() {
				broker.GetPubSub().Close()
				broker.ClientClose()
				broker.ResetRecieverTimeout()
				restarted = true
				go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(string(broker.identifier))+"websocket", "> *RESTARTING RECEIVER BROKER*"), slack.LootboxHealthCheck)
			}
			time.Sleep(5 * time.Second)
		}
		logger.Info(broker.GetIdentifier(), " restarter exited")
	}()

	go func() {
		defer broker.ClientClose()

	broker_loop:
		for {
			select {
			case <-broker.closeChan:
				logger.Info(broker.GetIdentifier(), " receiver close channel triggered")
				break broker_loop
			default:
				msg, err := broker.GetPubSub().ReceiveMessage(ctx)

				if err != nil {
					logger.Error(broker.GetIdentifier(), " receiver publish error: ", err.Error())
					continue
				}
				broker.publishToChannel(msg.Payload)
				go func() { //resets ResetPublishTimeout and ResetRecieverTimeout when received message from publish
					if restarted {
						go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(string(broker.identifier))+"websocket", "> *PUBLISH AND RECEIVER SUCESSFULLY RESTARTED*"), slack.LootboxHealthCheck)
					}
					restarted = false
					broker.ResetPublishTimeout()
					broker.ResetRecieverTimeout()
				}()
			}
		}
		stopRestarter = true
		logger.Info(broker.GetIdentifier(), " receiver loop exited")
	}()
}

func (broker *messageBroker) Stop() {
	close(broker.closeChan)
}

func (broker *messageBroker) ResetRecieverTimeout() {
	key := broker.channel + "-reciever-restart-timeout"

	timeOut := time.Now().UnixMilli() + broker.timeoutMS
	_redis.Cache().Set(key, *types.Int(timeOut).String().Ptr(), time.Duration(broker.timeoutMS)*time.Millisecond)
}

func (broker *messageBroker) ShouldRestartReciever() bool {
	key := broker.channel + "-reciever-restart-timeout"
	cTS, err := _redis.Cache().Get(key)

	if err != nil {
		logger.Info(broker.GetIdentifier(), " receiver ShouldRestartReciever key: (", cTS, ")")
		logger.Error(broker.GetIdentifier(), " receiver ShouldRestartReciever error: ", err.Error())
	}
	if cTS == "" {
		return true
	}
	return types.String(cTS).Int().Int64() < time.Now().UnixMilli()
}

func (broker *messageBroker) publishToChannel(msg string) {
	for _, sChan := range broker.subscribers {
		go func(c chan<- string, msg string) {
			broker.sendCount++
			c <- msg
			broker.sentCount++
		}(sChan, msg)
	}
}

// common broker implementation
func (broker *messageBroker) ClientClose() {
	if broker.retries >= constants.MAX_WS_RETRIES {
		panic(types.String(broker.GetIdentifier()) + " MAX_WS_RETRIES (" + types.Int(constants.MAX_WS_RETRIES).String() + ") reached")
	}
	broker.retries++
	broker.GetClient().Close()
	broker.redis = nil  //set nil to restart redis
	broker.pubSub = nil //set nil to restart pubsub
}

func (broker *messageBroker) GetClient() *redis.Client {
	if broker.redis == nil {
		logger.Info(broker.GetIdentifier(), " initialize redis client")
		broker.redis = _redis.NewRedis()
	}
	return broker.redis.GetClient()
}

func (broker *messageBroker) GetPubSub() *redis.PubSub {
	if broker.pubSub == nil {
		logger.Info(broker.GetIdentifier(), " initialize redis pubSub")
		broker.pubSub = broker.GetClient().Subscribe(ctx, broker.channel)
	}
	return broker.pubSub
}
