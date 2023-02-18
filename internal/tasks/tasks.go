package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/service"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsconsumer"
	"github.com/hibiken/asynq"
)

// A list of task types.
const (
	TypeEventBetting     = "event:betting"
	TypeEventStopBetting = "event:stopbetting"
	TypeEventResult      = "event:result"
)

type EventPayload struct {
	EventID int64
}

var (
	eventService service.EventService = service.NewEvent()
)

//----------------------------------------------
// Write a function NewXXXTask to create a task.
// A task consists of a type and a payload.
//----------------------------------------------

func NewBettingTask(event models.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventBetting, payload), nil
}

func NewStopBettingTask(event models.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventStopBetting, payload), nil
}

func NewShowResultTask(event models.Event) (*asynq.Task, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeEventResult, payload), nil
}

//---------------------------------------------------------------
// Write a function HandleXXXTask to handle the input task.
// Note that it satisfies the asynq.HandlerFunc interface.
//
// Handler doesn't need to be a function. You can define a type
//---------------------------------------------------------------

func HandleBettingTask(ctx context.Context, t *asynq.Task) error {
	var p models.Event
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	logger.Info("execute betting task: ", p.ID, p, time.Now().String())

	// update status to 1 - active
	p.Status = 1
	eventService.UpdateEventStatus(*p.ID, p)

	endTime := p.StartDatetime.Add(7*time.Second).UnixNano() / int64(time.Millisecond)

	// broadcast redis pub/sub
	message := map[string]interface{}{
		"type": "state",
		"data": map[string]interface{}{"name": "BETTING", "end": endTime},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		logger.Error("unexpected JSON format")
	}

	wsconsumer.NewMessagePublishBroker().Publish(string(payload))

	return nil
}

func HandleStopBettingTask(ctx context.Context, t *asynq.Task) error {
	var p models.Event
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// update status to 1 - for settlement
	p.Status = 3
	eventService.UpdateEventStatus(*p.ID, p)

	// betting time 7 sec + 3 sec stop betting buffer
	endTime := p.StartDatetime.Add(10*time.Second).UnixNano() / int64(time.Millisecond)

	logger.Info("execute stop betting task: ", p.ID, time.Now().String(), endTime)

	// broadcast redis pub/sub
	message := map[string]interface{}{
		"type": "state",
		"data": map[string]interface{}{"name": "STOP_BETTING", "end": endTime},
	}

	payload, err := json.Marshal(message)
	if err != nil {
		logger.Error("unexpected JSON format")
	}

	wsconsumer.NewMessagePublishBroker().Publish(string(payload))

	return nil
}

func HandleShowResult(ctx context.Context, t *asynq.Task) error {
	var p models.Event
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	// update status to 4 - settlement in progress? where to set this?
	// p.Status = 4
	// eventService.UpdateEvent(p)

	// update status to 5 - settled
	p.Status = 5
	eventService.UpdateEventStatus(*p.ID, p)

	// betting time 7 sec + 3 sec stop betting buffer + 4 sec show result
	endTime := p.StartDatetime.Add(14*time.Second).UnixNano() / int64(time.Millisecond)
	logger.Info("execute stop betting task: ", p.ID, time.Now().String(), endTime)

	// broadcast redis pub/sub
	message := map[string]interface{}{
		"type": "state",
		"data": map[string]interface{}{"name": "SHOW_RESULT", "end": endTime},
	}
	m1, err1 := json.Marshal(message)
	if err1 != nil {
		logger.Error("unexpected JSON format")
	}

	wsconsumer.NewMessagePublishBroker().Publish(string(m1))

	bomb := eventService.GetEventLolTowerBomb(*p.ID)
	// broadcast redis pub/sub bomb result
	message2 := map[string]interface{}{
		"type": "bomb",
		"data": map[string]interface{}{"bomb": bomb},
	}

	m2, err := json.Marshal(message2)
	if err != nil {
		logger.Error("unexpected JSON format")
	}

	wsconsumer.NewMessagePublishBroker().Publish(string(m2))

	return nil
}

/*
0 - (Enabled),
1 - (Active),
2 - (Disabled),
3 - (For Settlement),
4 - (Settlement in Progress),
5 - (Settled),
6 - (Cancelled)

{
    “type”:"games.tower.state",
    “data”: {"name":"BETTING","end":1644235865875}
}
{
    “type”:"games.tower.state",
    “data”: {"name":"STOP_BETTING","end":1644235868875}
}
{
    “type”: "games.tower.state",
    “data”: {"name": "SHOW_RESULT", "end": 1644237630414}
}
{
    “type”: "games.tower.bomb",
    “data”: {"bomb": "1,3"}
}
*/
