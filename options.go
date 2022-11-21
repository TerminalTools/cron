package cron

import (
	"time"
)

type Options struct {
	interval     time.Duration    // 时间间隔
	correction   bool             // 是否补正
	firstTrigger bool             // 是否在启动定时前执行任务
	offset                        // 时间偏移量
	task         func()           // 定时任务
	now          func() time.Time // 获取当前时间的方法
}

func NewOptions(task func()) (object *Options) {
	object = &Options{
		task: task,
	}
	return object
}

func (self *Options) SetInterval(interval time.Duration) {
	self.interval = interval
}

func (repl Options) GetInterval() time.Duration {
	if repl.interval == 0 {
		return defaultInterval
	}

	return repl.interval
}

func (self *Options) SetCorrection(correction bool) {
	self.correction = correction
}

func (repl Options) GetCorrection() bool {
	return repl.correction
}

func (self *Options) SetFirstTrigger(firstTrigger bool) {
	self.firstTrigger = firstTrigger
}

func (repl Options) GetFirstTrigger() bool {
	return repl.firstTrigger
}

func (self *Options) SetOffset(second, minute, hour int64) {
	self.offset.second, self.offset.minute, self.offset.hour = second, minute, hour
}

func (repl Options) GetOffset() offset {
	return repl.offset
}

func (self *Options) SetTask(task func()) {
	self.task = task
}

func (repl Options) GetTask() func() {
	return repl.task
}

func (self *Options) SetNowFunc(function func() time.Time) {
	self.now = function
}

func (repl Options) GetNowFunc() func() time.Time {
	if repl.now == nil {
		return time.Now
	}

	return repl.now
}

type offset struct {
	second int64 // 偏移秒数
	minute int64 // 偏移分钟数
	hour   int64 // 偏移小时数
}

func (repl offset) Duration() time.Duration {
	return time.Second*time.Duration(repl.second) + time.Minute*time.Duration(repl.minute) + time.Hour*time.Duration(repl.hour)
}
