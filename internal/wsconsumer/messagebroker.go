package wsconsumer

import (
	"context"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/cache"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	r "bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background() //restart pr6

type MessagePublishBroker interface {
	Publish(msg string)
}

type MessageRecieverBroker interface {
	Subscribe(c chan<- string)
	Unsubcribe(c chan<- string)
	Start()
	Stop()
}

type messageBroker struct {
	subscribers map[chan<- string]chan<- string
	closeChan   chan uint
	redis       r.Redis
	pubSub      *redis.PubSub
	retries     int8
}

func NewMessagePublishBroker() MessagePublishBroker {
	broker := messageBroker{}

	broker.ResetPublishTimeout()
	return &broker
}

func NewMessageRecieverBroker() MessageRecieverBroker {
	broker := messageBroker{
		subscribers: map[chan<- string]chan<- string{},
		closeChan:   make(chan uint),
	}

	broker.ResetRecieverTimeout()
	return &broker
}

//Publish
func (broker *messageBroker) Publish(msg string) {
	if broker.ShouldRestartPublish() {
		logger.Info("messagebroker publish restart client on publish")
		broker.ClientClose()
		broker.ResetPublishTimeout()
		slack.SendMessage(":bangbang: *WEB-SOCKET MESSAGES* :bangbang:", "Restarting publish broker", slack.LootboxHealthCheck)
	}
	if err := broker.GetClient().Publish(ctx, settings.LOL_TOWER_CHANNEL, msg).Err(); err != nil {
		logger.Error("messagebroker Publish error: ", err.Error())
	}
}

func (broker *messageBroker) ResetPublishTimeout() {
	key := settings.LOL_TOWER_CHANNEL + "-publish-restart-timeout"

	timeOut := time.Now().Unix() + constants.LOL_TOWER_GAME_DURATION
	cache.Main.Set(key, *types.Int(timeOut).String().Ptr(), constants.LOL_TOWER_GAME_DURATION*time.Second)
}

func (broker *messageBroker) ShouldRestartPublish() bool {
	key := settings.LOL_TOWER_CHANNEL + "-publish-restart-timeout"
	cTS, err := cache.Main.GetOrig(key)

	if err != nil {
		logger.Info("messagebroker ShouldRestartPublish key: (", cTS, ")")
		logger.Error("messagebroker ShouldRestartPublish error: ", err.Error())
	}

	if cTS == "" {
		return true
	}
	return types.String(cTS).Int().Int64() < time.Now().Unix()
}

//Receiver
func (broker *messageBroker) Subscribe(sub chan<- string) {
	broker.subscribers[sub] = sub
}

func (broker *messageBroker) Unsubcribe(sub chan<- string) {
	delete(broker.subscribers, sub)
}

func (broker *messageBroker) publishToChannel(msg string) {
	for _, c := range broker.subscribers {
		c <- msg
	}
}

func (broker *messageBroker) Start() {
	stopRestarter := false

	go func() {
		for !stopRestarter {
			if broker.ShouldRestartReciever() {
				broker.GetPubSub().Close()
				broker.ClientClose()
				broker.ResetRecieverTimeout()
				slack.SendMessage(":bangbang: *WEB-SOCKET MESSAGES* :bangbang:", "Restarting receiver broker", slack.LootboxHealthCheck)
			}
			time.Sleep(5 * time.Second)
		}
		logger.Info("messagebroker restarter exited")
	}()

	go func() {
		defer broker.ClientClose()

	broker_loop:
		for {
			select {
			case <-broker.closeChan:
				logger.Info("messagebroker receiver close channel triggered")
				break broker_loop
			default:
				msg, err := broker.GetPubSub().ReceiveMessage(ctx)

				if err != nil {
					logger.Error("messagebroker publish message error: ", err)
					continue
				}
				broker.publishToChannel(msg.Payload)
				go func() { //resets ResetPublishTimeout and ResetRecieverTimeout when received message from publish
					broker.ResetPublishTimeout()
					broker.ResetRecieverTimeout()
				}()
			}
		}
		stopRestarter = true
		logger.Info("messagebroker receiver loop exited")
	}()

}

func (broker *messageBroker) Stop() {
	close(broker.closeChan)
}

func (broker *messageBroker) ResetRecieverTimeout() {
	key := settings.LOL_TOWER_CHANNEL + "-reciever-restart-timeout"

	timeOut := time.Now().Unix() + constants.LOL_TOWER_GAME_DURATION
	cache.Main.Set(key, *types.Int(timeOut).String().Ptr(), constants.LOL_TOWER_GAME_DURATION*time.Second)
}

func (broker *messageBroker) ShouldRestartReciever() bool {
	key := settings.LOL_TOWER_CHANNEL + "-reciever-restart-timeout"
	cTS, err := cache.Main.GetOrig(key)

	if err != nil {
		logger.Info("messagebroker receiver ShouldRestartReciever key: (", cTS, ")")
		logger.Error("messagebroker receiver ShouldRestartReciever error: ", err.Error())
	}
	if cTS == "" {
		return true
	}
	return types.String(cTS).Int().Int64() < time.Now().Unix()
}

//common broker implementation
func (broker *messageBroker) ClientClose() {
	if broker.retries >= constants.MAX_WS_RETRIES {
		panic("messagebroker MAX_WS_RETRIES (" + types.Int(constants.MAX_WS_RETRIES).String() + ") reached")
	}
	broker.retries++
	broker.GetClient().Close()
	broker.redis = nil  //set nil to restart redis
	broker.pubSub = nil //set nil to restart pubsub
}

func (broker *messageBroker) GetClient() *redis.Client {
	if broker.redis == nil {
		logger.Info("messagebroker initialize redis client")
		broker.redis = r.NewRedis()
	}
	return broker.redis.GetClient()
}

func (broker *messageBroker) GetPubSub() *redis.PubSub {
	if broker.pubSub == nil {
		logger.Info("messagebroker initialize redis pubSub")
		broker.pubSub = broker.GetClient().Subscribe(ctx, settings.LOL_TOWER_CHANNEL)
	}
	return broker.pubSub
}
