package routine

import (
	"fmt"
	"github.com/faymajun/gonano/core"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/faymajun/gonano/tags"

	log "github.com/sirupsen/logrus"
)

var (
	stopGoChan  = make(chan struct{})
	waitAll     core.WaitGroup // 等待所有goroutine
	waitRedis   sync.WaitGroup
	gorontineId uint64
)

func StopAll() {
	log.Infof("Stop all gorontine start...")
	close(stopGoChan)
	timeout := waitAll.WaitTimeout(60)
	if timeout {
		log.Infof("Stop all gorontine timeout, will be shutdown immediately")
	} else {
		log.Infof("Stop all gorontine gracefully")
	}
}

// 协程封装
func run(fn func()) {
	if tags.DEBUG {
		_, file, line, _ := runtime.Caller(2)
		i := strings.LastIndex(file, "/") + 1
		i = strings.LastIndex(file[:i-1], "/") + 1
		createdAt := fmt.Sprintf("%s:%d", file[i:], line)

		currentId := atomic.AddUint64(&gorontineId, 1)
		currentCount := atomic.AddInt64(&Stats.RoutineCount, 1)
		log.Debugf("Starting Goroutine(id:%d, count:%d, line:%s)",
			currentId, currentCount, createdAt)

		waitAll.Add(1)
		go func() {
			Try(fn, nil)
			remainCount := atomic.AddInt64(&Stats.RoutineCount, -1)
			log.Debugf("Goroutine terminated(id:%d, count:%d, line:%s)", currentId, remainCount, createdAt)

			waitAll.Done()
		}()
	} else {
		atomic.AddUint64(&gorontineId, 1)
		atomic.AddInt64(&Stats.RoutineCount, 1)

		waitAll.Add(1)
		go func() {
			Try(fn, nil)
			atomic.AddInt64(&Stats.RoutineCount, -1)
			waitAll.Done()
		}()
	}
}

func Go(fn func()) {
	if !Pool.Serve(fn) { //复用goroutine
		logger.Warnf("警告!,繁忙goroutine已达%d", MaxWorkersCount)
		run(fn)
	}
}

func GoChan(fn func(cstop chan struct{})) {
	run(func() { fn(stopGoChan) })
}
