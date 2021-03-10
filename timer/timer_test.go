package timer

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTimer(t *testing.T) {
	const tc = 1000
	var counter int64
	for i := 0; i < tc; i++ {
		NewTimer(1*time.Millisecond, func() {
			atomic.AddInt64(&counter, 1)
		})
	}

	<-time.After(5 * time.Millisecond)
	Cron()
	Cron()
	if counter != tc*2 {
		t.Fatalf("expect: %d, got: %d", tc*2, counter)
	}

	if len(manager.timers) != tc {
		t.Fatalf("timers: %d", len(manager.timers))
	}

	if len(manager.createdTimer) != 0 {
		t.Fatalf("createdTimer: %d", len(manager.createdTimer))
	}

	if len(manager.closingTimer) != 0 {
		t.Fatalf("closingTimer: %d", len(manager.closingTimer))
	}
}

func TestNewAfterTimer(t *testing.T) {
	const tc = 1000
	var counter int64
	for i := 0; i < tc; i++ {
		NewAfterTimer(1*time.Millisecond, func() {
			atomic.AddInt64(&counter, 1)
		})
	}

	<-time.After(5 * time.Millisecond)
	Cron()
	if counter != tc {
		t.Fatalf("expect: %d, got: %d", tc, counter)
	}

	if len(manager.timers) != 0 {
		t.Fatalf("timers: %d", len(manager.timers))
	}

	if len(manager.createdTimer) != 0 {
		t.Fatalf("createdTimer: %d", len(manager.createdTimer))
	}

	if len(manager.closingTimer) != 0 {
		t.Fatalf("closingTimer: %d", len(manager.closingTimer))
	}
}
