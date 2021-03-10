package routine

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Try-catch wrapper
func Try(fn func(), catch func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if catch == nil {
				logrus.Errorf("Try panic: %s", err)
				fmt.Fprintln(logrus.StandardLogger().Out, string(debug.Stack()))
			} else {
				catch(err)
			}
			atomic.AddInt64(&Stats.PanicCount, 1)
			Stats.LatestPanicTime = time.Now()
		}
	}()
	fn()
}
