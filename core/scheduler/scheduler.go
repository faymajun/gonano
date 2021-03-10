package scheduler

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/faymajun/gonano/colorized"
	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/message"
	"github.com/faymajun/gonano/tags"
	"github.com/faymajun/gonano/timer"

	log "github.com/sirupsen/logrus"
)

// 1个8核服务器，开4台战斗服
const queueBacklog = 4096

var scheduler = NewScheduler()

type Scheduler struct {
	started, stoped int32
	chDead          chan struct{}
	chFunc          chan func()
	chPackets       chan *packet.RecvMessage
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		chDead:    make(chan struct{}),
		chFunc:    make(chan func(), queueBacklog),
		chPackets: make(chan *packet.RecvMessage, queueBacklog),
	}
}

func Process(pack *packet.RecvMessage) {
	var (
		session = pack.Session
		handler = pack.Handler
		pbMsg   = pack.Payload
	)

	if tags.DEBUG {
		if msgid := handler.MsgID(); msgid != message.MSGID_ReqHeartbeatE &&
			msgid != message.MSGID_ResHeartbeatE {
			println(fmt.Sprintf(colorized.Yellow("++> %s, %+v, %s"),
				handler.MsgID().String(), pbMsg, session.RemoteAddr()))
		}
	}

	if err := handler.Handle(session, pbMsg); err != nil {
		log.Errorf("Handle %v error: %v", handler.MsgID(), err)
	}
}

func (sched *Scheduler) Start() {
	if atomic.AddInt32(&sched.started, 1) != 1 {
		return
	}

	ticker := time.NewTicker(timer.Precision())
	defer func() {
		close(sched.chDead)
		ticker.Stop()
	}()

loop:
	for {
		select {
		case fn, ok := <-sched.chFunc:
			if !ok {
				log.Warnf("Scheduler function queue was closed")
				sched.chFunc = nil
				if sched.chPackets == nil {
					break loop
				}
				continue
			}
			if fn == nil {
				log.Errorf("Scheduler receive a nil function")
				continue
			}
			routine.Try(fn, nil)

		case pack, ok := <-sched.chPackets:
			if !ok {
				log.Warnf("Scheduler packets queue was closed")
				sched.chPackets = nil
				if sched.chFunc == nil {
					break loop
				}
				continue
			}
			routine.Try(func() { Process(pack) }, nil)

		case <-ticker.C:
			timer.Cron()
		}
	}
}

func (sched *Scheduler) Stop() {
	if atomic.LoadInt32(&sched.started) == 0 {
		return
	}

	if atomic.AddInt32(&sched.stoped, 1) != 1 {
		return
	}

	log.Infof("Scheduler ready for stopping...")
	close(sched.chFunc)
	close(sched.chPackets)

	select {
	case <-sched.chDead:
		log.Infof("Scheduler was stopped gracefully")

	case <-time.After(60 * time.Second):
		log.Infof("Stop scheduler timeout, will be shutdown immediately")
	}
}

func (sched *Scheduler) PushTask(fn func()) {
	if atomic.LoadInt32(&sched.stoped) != 0 {
		return
	}
	routine.Try(func() { sched.chFunc <- fn }, nil)

}

func (sched *Scheduler) PushPacket(pack *packet.RecvMessage) {
	if atomic.LoadInt32(&sched.stoped) != 0 {
		return
	}
	routine.Try(func() { sched.chPackets <- pack }, nil)

}
