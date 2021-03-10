package coroutine

import (
	"errors"
	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/core/scheduler"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

const RoutineQueLen = 64 //队列长度
var logger = logrus.WithField("component", "net")

type Coroutine struct {
	chRead chan *packet.RecvMessage // 读通道
	chFunc chan func()              // 处理函数channel
	close  int32                    // 状态
	userId int64                    // 用户id
}

func NewCoroutine(queCap int, userId int64) *Coroutine {
	c := &Coroutine{
		chRead: make(chan *packet.RecvMessage, queCap),
		chFunc: make(chan func(), queCap),
		close:  0,
		userId: userId,
	}
	routine.Go(func() {
		c.process()
	})
	return c
}

func TestCoroutine(chRead chan *packet.RecvMessage, chFunc chan func()) *Coroutine {
	c := &Coroutine{
		chRead: chRead,
		chFunc: chFunc,
		close:  0,
		userId: 0,
	}
	return c
}

func (c *Coroutine) isClose() bool {
	return atomic.LoadInt32(&c.close) == 1
}

func (c *Coroutine) Close() error {
	if !atomic.CompareAndSwapInt32(&c.close, 0, 1) {
		return errors.New("duplication close")
	}

	close(c.chRead)
	close(c.chFunc)
	return nil
}

func (c *Coroutine) process() {
loop:
	for {
		select {
		case fn, ok := <-c.chFunc:
			if !ok {
				c.chFunc = nil
				if c.chRead == nil {
					break loop
				}
				continue
			}
			if fn == nil {
				logger.Errorf("Coroutine receive a nil function")
				continue
			}
			routine.Try(fn, nil)

		case pack, ok := <-c.chRead:
			if !ok {
				c.chRead = nil
				if c.chFunc == nil {
					break loop
				}
				continue
			}
			routine.Try(func() { scheduler.Process(pack) }, nil)

		}
	}
	c.Close()
}

func (c *Coroutine) PushTask(fn func(), permitDefeat bool) error {
	if c.isClose() {
		return nil
	}

	if permitDefeat && len(c.chFunc) >= (cap(c.chFunc)*7/10) {
		logger.Warnf("permitDefeat PushTask chFunc buffer excced, session close", c.userId)
		return nil
	}

	if len(c.chFunc) >= cap(c.chFunc) {
		logger.Warnf("PushTask chFunc buffer excced, session close", c.userId)
	}

	routine.Try(func() { c.chFunc <- fn }, nil)
	return nil
}

func (c *Coroutine) PushPacket(packet *packet.RecvMessage) error {
	if c.isClose() {
		return nil
	}

	routine.Try(func() { c.chRead <- packet }, nil)
	return nil
}
