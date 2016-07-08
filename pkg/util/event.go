package util

import (
	"time"
)

type EventBus struct {
	stopChan chan bool
	evtChan  chan string
	handlers map[string]func()
	timeout  time.Duration
}

func NewEventBus(bufferSize int, timeout time.Duration) *EventBus {
	return &EventBus{
		stopChan: make(chan bool),
		evtChan:  make(chan string, bufferSize),
		timeout:  timeout,
	}
}

func (self *EventBus) Registry(evt string, handler func()) {
	self.handlers[evt] = handler
}

func (self *EventBus) PublishEvent(evt string) {
	self.evtChan <- evt
}

func (self *EventBus) Stop() {
	self.stopChan <- true
}

func (self *EventBus) EventLoop() {
	for {
		select {
		case <-self.stopChan:
			return
		case evt := <-self.evtChan:
			fun, ok := self.handlers[evt]

			if ok {
				fun()
			}
		}
	}
}
