package cron

import (
	"errors"
	"sync"
	"time"
)

type Ticker struct {
	*time.Ticker
	tick                chan time.Time
	latestTickTime      time.Time        // 最新一次触发tick的时间
	interval            time.Duration    // 时间间隔
	guide               int64            // 采集时刻
	correction          bool             // 是否补正
	correctionThreshold int64            // 补正阈值
	firstTrigger        bool             // 是否在启动定时前执行任务
	offset              offset           // 偏移时间
	now                 func() time.Time // 获取当前时间方法
	tickerLock          *sync.RWMutex
}

func NewTicker(options *Options) (ticker *Ticker, newErr error) {
	ticker = &Ticker{
		interval:     options.GetInterval(),
		offset:       options.GetOffset(),
		correction:   options.GetCorrection(),
		firstTrigger: options.GetFirstTrigger(),
		tick:         make(chan time.Time, 1),
		now:          options.GetNowFunc(),
		tickerLock:   &sync.RWMutex{},
	}

	if ticker.interval == 0 {
		return nil, errors.New("time interval cannot be zero")
	}

	ticker.initInterval()

	if ticker.interval <= ticker.offset.Duration() {
		return nil, errors.New("the offset must be less than the interval")
	}

	ticker.initThreshold()
	ticker.initGuide()
	ticker.initTicker()
	return ticker, nil
}

// initGuide 初始化采集时刻, 按最小粒度上靠
func (self *Ticker) initGuide() {
	var granularity time.Duration
	if self.interval/time.Second < 1 { // 毫秒级别
		granularity = minInterval
	} else if self.interval/time.Minute < 1 { // 秒级
		granularity = time.Second
	} else if self.interval/time.Hour < 1 { // 分钟级
		granularity = time.Minute
	} else { // 小时级
		granularity = time.Hour
	}

	var (
		quotient  = self.interval / granularity
		remainder = self.interval % granularity
	)
	if remainder != 0 {
		quotient = quotient + 1 // interval=>quotient: 0=>0, 1=>1, 100=>1, 101=>2
	}
	self.guide = int64(quotient * granularity)
}

// initThreshold 初始化补正阈值
func (self *Ticker) initThreshold() {
	var correctionThreshold int64
	if self.interval/time.Second < 1 { // 毫秒级别
		correctionThreshold = int64(time.Millisecond) * 10
	} else if self.interval/time.Minute < 1 { // 秒级
		correctionThreshold = int64(time.Millisecond) * 100
	} else if self.interval/time.Hour < 1 { // 分钟级
		correctionThreshold = int64(time.Second)
	} else { // 小时级
		correctionThreshold = int64(time.Minute)
	}
	self.correctionThreshold = correctionThreshold
}

// initInterval 初始化定时器间隔(有限制最小间隔)
func (self *Ticker) initInterval() {
	if !self.correction { // 不要求校正
		return
	}

	if self.interval < minInterval {
		self.interval = minInterval
	}
}

// initTicker 初始化定时器
func (self *Ticker) initTicker() {
	self.Ticker = time.NewTicker(self.interval)
	go self.startTime()
}

// doCorrection 时间补正, 整点执行定时任务
func (self *Ticker) doCorrection() (trigger bool) {
	var (
		now     = self.now()
		nowNano = now.UnixNano()
	)
	var (
		excessTime     = nowNano - int64(self.offset.Duration())
		correctionTime = excessTime % self.guide
	)

	if correctionTime < self.correctionThreshold {
		return false
	}

	var (
		waitTime = time.Duration(self.guide - correctionTime)
		wait     = time.NewTimer(waitTime)
	)

	<-wait.C
	self.Ticker.Reset(self.interval)
	return true
}

// first 在启动定时前执行一次任务
func (self *Ticker) first() {
	if !self.firstTrigger {
		return
	}
	if self.correction {
		self.doCorrection()
	}
	self.recvTick(time.Now())
}

// startTime 启动定时器
func (self *Ticker) startTime() {
	self.first()
	for now := range self.Ticker.C {
		if self.correction {
			self.doCorrection()
		}
		self.recvTick(now)
	}
}

func (self *Ticker) recvTick(tick time.Time) {
	self.tickerLock.Lock()
	defer self.tickerLock.Unlock()

	if self.correction {
		var (
			snice               = time.Since(self.latestTickTime)
			interval            = time.Duration(float64(self.interval) * tickerMinIntervalCoefficient)
			correctionThreshold = time.Duration(self.correctionThreshold)
		)
		// 启动时间补正后, 两次tick之间必须大于补正阈值
		if snice < correctionThreshold {
			return
		}
		// 启动时间补正后, 两次tick之间必须大于定时器间隔
		if snice < interval {
			return
		}
	}

	self.latestTickTime = time.Now()
	self.tick <- tick
}

// Tick 返回定时器
func (self *Ticker) Tick() <-chan time.Time {
	return self.tick
}
