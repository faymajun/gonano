package routine

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/tags"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("routine", "pool")
var Pool *workerPool

const MaxWorkersCount = 256 * 1024

func init() {
	WorkPoolStart()
}

func WorkPoolStart() {
	Pool = &workerPool{MaxWorkersCount: MaxWorkersCount}
	Pool.Start()
}

type workerPool struct {
	MaxWorkersCount       int
	LogAllErrors          bool
	MaxIdleWorkerDuration time.Duration
	lock                  sync.Mutex
	workersCount          int
	mustStop              bool
	ready                 []*workerChan
	stopCh                chan struct{}
	workerChanPool        sync.Pool
}

var chId uint64

type workerChan struct {
	lastUseTime time.Time
	fun         chan func()
	id          uint64
}

func (wp *workerPool) Start() {
	if wp.stopCh != nil {
		panic("BUG: workerPool already started")
	}
	wp.stopCh = make(chan struct{})
	stopCh := wp.stopCh
	go func() {
		var scratch []*workerChan
		for {
			wp.clean(&scratch)
			select {
			case <-stopCh:
				return
			default:
				time.Sleep(wp.getMaxIdleWorkerDuration())
			}
		}
	}()
}

func (wp *workerPool) Stop() {
	if wp.stopCh == nil {
		panic("BUG: workerPool wasn't started")
	}
	close(wp.stopCh)
	wp.stopCh = nil
	wp.lock.Lock()
	ready := wp.ready
	for i, ch := range ready {
		ch.fun <- nil
		ready[i] = nil
	}
	wp.ready = ready[:0]
	wp.mustStop = true
	wp.lock.Unlock()

	wp.waitStop()
}

func (wp *workerPool) waitStop() {
	logger.Infof("Stop all worker start...")
	start := time.Now()
	for wp.workersCount > 0 {
		if time.Since(start).Seconds() >= float64(60) { // 60秒等待所有的协程任务结束
			logger.Infof("Stop all worker timeout, will be shutdown immediately")
			return
		}
		time.Sleep(time.Millisecond * 1)
	}
	logger.Infof("Stop all worker gracefully")

}

func (wp *workerPool) getMaxIdleWorkerDuration() time.Duration {
	if wp.MaxIdleWorkerDuration <= 0 {
		return 10 * time.Second
	}
	return wp.MaxIdleWorkerDuration
}

func (wp *workerPool) clean(scratch *[]*workerChan) {
	maxIdleWorkerDuration := wp.getMaxIdleWorkerDuration()
	currentTime := time.Now()
	wp.lock.Lock()
	ready := wp.ready
	n := len(ready)
	i := 0
	for i < n && currentTime.Sub(ready[i].lastUseTime) > maxIdleWorkerDuration {
		i++
	}
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		m := copy(ready, ready[i:])
		for i = m; i < n; i++ {
			ready[i] = nil
		}
		wp.ready = ready[:m]
	}
	wp.lock.Unlock()
	tmp := *scratch
	for i, ch := range tmp {
		ch.fun <- nil
		tmp[i] = nil
	}
}

func (wp *workerPool) Serve(fn func()) bool {
	ch := wp.getCh()
	if ch == nil {
		return false
	}
	ch.fun <- fn
	return true
}

var workerChanCap = func() int {
	if runtime.GOMAXPROCS(0) == 1 {
		return 0
	}
	return 1
}()

func (wp *workerPool) getCh() *workerChan {
	var ch *workerChan
	createWorker := false
	wp.lock.Lock()
	ready := wp.ready
	n := len(ready) - 1
	if n < 0 {
		if wp.workersCount < wp.MaxWorkersCount {
			createWorker = true
			wp.workersCount++
		}
	} else {
		ch = ready[n]
		ready[n] = nil
		wp.ready = ready[:n]
	}
	wp.lock.Unlock()
	if ch == nil {
		if !createWorker {
			return nil
		}
		vch := wp.workerChanPool.Get()
		if vch == nil {
			id := atomic.AddUint64(&chId, 1)
			vch = &workerChan{
				fun: make(chan func(), workerChanCap),
				id:  id,
			}

			if tags.DEBUG {
				logger.Debugf("Starting gorontine count:%d, id:%d", wp.workersCount, id)
			}
		}
		ch = vch.(*workerChan)
		go func() {
			wp.workerFunc(ch)
			wp.workerChanPool.Put(vch)
		}()
	}
	return ch
}

func (wp *workerPool) release(ch *workerChan) bool {
	ch.lastUseTime = time.Now()
	wp.lock.Lock()
	if wp.mustStop {
		wp.lock.Unlock()
		return false
	}
	wp.ready = append(wp.ready, ch)
	wp.lock.Unlock()
	return true
}

func (wp *workerPool) workerFunc(ch *workerChan) {
	for fn := range ch.fun {
		if fn == nil {
			break
		}
		Try(fn, nil)
		if !wp.release(ch) {
			break
		}
	}
	wp.lock.Lock()
	wp.workersCount--
	if tags.DEBUG {
		logger.Debugf("Goroutine terminated count:%d), id:%d", wp.workersCount, ch.id)
	}
	wp.lock.Unlock()
}
