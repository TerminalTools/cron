package cron

import (
	"errors"
	"sync"

	"github.com/TerminalTools/once-do"
)

type Cron struct {
	done   chan struct{}
	ticker *Ticker
	lock   *sync.RWMutex
	task   *once.Once
	once   *once.Once
}

// NewCron 新增定时任务
func NewCron(options *Options) (object *Cron, newErr error) {
	if options == nil {
		return nil, errors.New("options cannot be empty")
	}

	if options.GetTask() == nil {
		return nil, errors.New("task cannot be empty")
	}

	ticker, newTickerErr := NewTicker(options)
	if newTickerErr != nil {
		return nil, newTickerErr
	}

	task, newTaskErr := once.NewOnce(options.GetTask())
	if newTaskErr != nil {
		return nil, newTaskErr
	}

	object = &Cron{
		done:   make(chan struct{}),
		lock:   &sync.RWMutex{},
		ticker: ticker,
		task:   task,
	}

	onceDo, newOnceErr := once.NewOnce(object.start)
	if newOnceErr != nil {
		return nil, newOnceErr
	}
	object.once = onceDo

	return object, nil
}

// Start 启动定时任务
func (self *Cron) Start() {
	go self.once.Do()
}

func (self *Cron) start() {
	for {
		select {
		case <-self.ticker.tick:
			go self.task.Do()
		case <-self.done:
			return
		}
	}
}

// Stop 停止定时任务
func (self *Cron) Stop() {
	self.lock.Lock()
	defer self.lock.Unlock()

	close(self.done)
	self.done = make(chan struct{})
}
