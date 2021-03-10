package routine

import "time"

var Stats struct {
	PanicCount      int64
	LatestPanicTime time.Time
	RoutineCount    int64
}
