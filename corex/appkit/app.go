package appkit

import (
	"context"
	log "github.com/sirupsen/logrus"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"
)

func New(name string, usage string, version string, loop func(ctx context.Context)) {
	app := &cli.App{
		Name:    name,
		Usage:   usage,
		Version: version,

		Action: func(c *cli.Context) error {

			//// c.Bool("pprof")
			//if c.Bool("pprof") {
			//	go http.ListenAndServe(":6060", nil)
			//}

			ctx, cancelFunc := context.WithCancel(context.TODO())
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, os.Kill, syscall.SIGTERM, os.Interrupt, syscall.SIGINT, syscall.SIGHUP)
			log.Infof("start server [%s]", name)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				loop(ctx)
			}()
			select {
			case <-sc:
				cancelFunc()
				log.Infof("stop server [%s]", name)
			}
			wg.Wait()
			return nil
		},
	}
	app.Run(os.Args)
}
