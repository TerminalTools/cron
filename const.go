package cron

import (
	"time"
)

// default
const (
	defaultInterval = time.Second
)

// ticker
const (
	minInterval                  = time.Millisecond * 100 // 定时器最小间隔
	tickerMinIntervalCoefficient = 0.5                    // 两次tick间最小间隔系数
)
