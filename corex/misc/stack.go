package misc

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime/debug"
)

func BackTrace(name string) {
	log.Errorf("goroutine[%s] is exiting...\n", name)
	buf := bytes.NewBuffer(debug.Stack())
	fmt.Fprintf(os.Stderr, buf.String())
	log.Error(buf.String())
}
