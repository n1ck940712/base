package wsconsumer

import (
	"time"
)

type DummyBroker struct {
	intervalMsec uint
	subscribers  map[chan<- string]chan<- string
	closeChan    chan uint
}

func NewDummyBroker(intervalMsec uint) *DummyBroker {
	return &DummyBroker{intervalMsec, map[chan<- string]chan<- string{}, make(chan uint)}
}

func (broker *DummyBroker) Subscribe(sub chan<- string) {
	broker.subscribers[sub] = sub
}

func (broker *DummyBroker) Unsubcribe(sub chan<- string) {
	delete(broker.subscribers, sub)
}

func (broker *DummyBroker) Publish(msg string) {
	for _, c := range broker.subscribers { //TODO check if this cause race condition or segfault cause we access this on possibly multiple threads/process
		c <- msg
	}
}

func (broker *DummyBroker) Start() {
	var exitLoop = false
	for !exitLoop {
		select {
		case <-broker.closeChan:
			exitLoop = true
		default:
			time.Sleep(time.Duration(broker.intervalMsec) * time.Millisecond)
			broker.Publish(`{"type": "dummy_message"}`)
		}
	}
}

func (broker *DummyBroker) Stop() {
	close(broker.closeChan)
}
