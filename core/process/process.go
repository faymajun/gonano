package process

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/faymajun/gonano/config"
	"github.com/faymajun/gonano/core/net"
	"github.com/faymajun/gonano/core/routine"
	"github.com/faymajun/gonano/core/scheduler"
	"github.com/faymajun/gonano/mongo"
	"github.com/faymajun/gonano/redis"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	status struct {
		started int32
		stoped  int32
	}

	logger = logrus.WithField("component", "process")
)

func Action(serve func(), exit func()) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		if atomic.AddInt32(&status.started, 1) != 1 {
			return fmt.Errorf("Server process has be started: %d", status.started)
		}

		rand.Seed(time.Now().UnixNano())

		//initConfig(ctx)
		loadLogConf()

		timeString := time.Now().Format("2006-01-02.15.04.05")
		for _, v := range ctx.Args() {
			if "cpuprof" == v {
				f, err := os.Create(fmt.Sprintf("cpuprof-%s", timeString))
				if err != nil {
					panic(err)
				}
				pprof.StartCPUProfile(f)
				defer pprof.StopCPUProfile()
			}

			if "memprof" == v {
				f, err := os.Create(fmt.Sprintf("memprof-%s", timeString))
				if err != nil {
					panic(err)
				}
				defer pprof.WriteHeapProfile(f)
			}
		}

		// 全局逻辑线程
		routine.Go(scheduler.Start)

		serve()

		// 等待退出信号
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGINT)

		select {
		case sig, _ := <-stopChan:
			logger.Infof("<<<==================>>>")
			logger.Infof("<<<stop process by:%v>>>", sig)
			logger.Infof("<<<==================>>>")
			break
		}

		if atomic.LoadInt32(&status.started) == 0 || atomic.AddInt32(&status.stoped, 1) != 1 {
			return fmt.Errorf("Server stop duplication")
		}

		// 设置关服标记
		config.SetClosing()

		signal.Stop(stopChan)
		close(stopChan)

		net.StopTcpServer()  // 关闭服务器监听，阻止新连接
		net.StopTcpClients() // 断开与其他服务的连接
		net.StopTcpSession() // 关闭客户端连接

		logger.Infof("Starting execute server shutdown hooks")
		exit()
		logger.Infof("Server shutdown hooks executed completed")

		scheduler.Stop()    // 关闭主派发协程-主协程内容为空往下执行
		routine.StopAll()   // 关闭所有的协程
		routine.Pool.Stop() // 协程池任务关闭

		// 关闭redis数据库
		redis.StopRedis()
		// 关闭日志mongo 连接
		mongo.StopMongo()

		logger.Infof("Server shutdown Finish!!!")
		return nil
	}
}
