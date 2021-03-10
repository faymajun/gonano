package grace

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var signalChan chan os.Signal
var hooks = make([]func(), 0)
var hookLock sync.Mutex

func init() {
	signalChan = make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP)
	signal.Notify(signalChan,
		os.Interrupt,
		os.Kill,
		syscall.SIGALRM, // 时钟定时信号
		// syscall.SIGHUP,   // 终端连接断开
		syscall.SIGINT,  // Ctrl-C
		syscall.SIGTERM, // 结束程序
		// syscall.SIGQUIT,  // Ctrl-/
	)
	go func() {
		for _ = range signalChan {
			for _, hook := range hooks {
				hook()
			}
			os.Exit(0)
		}
	}()
}

func OnInterrupt(fn func()) {
	hookLock.Lock()
	defer hookLock.Unlock()

	// control+c,etc
	// 控制终端关闭，守护进程不退出
	hooks = append(hooks, fn)
}
