package core

import (
	"sync/atomic"
	"time"
)

// 带超时机制的WaitGroup
type WaitGroup struct {
	count int64
}

func (self *WaitGroup) Add(delta int) {
	atomic.AddInt64(&self.count, int64(delta))
}

func (self *WaitGroup) Done() {
	atomic.AddInt64(&self.count, -1)
}

// timeout/秒 超时时间 ret:是否强制结束
func (self *WaitGroup) WaitTimeout(timeout int) bool {
	start := time.Now()
	for atomic.LoadInt64(&self.count) > 0 {
		if time.Since(start).Seconds() >= float64(timeout) {
			return true
		}
		time.Sleep(time.Millisecond * 1)
	}
	return false
}
