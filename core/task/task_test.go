package task

import (
	"fmt"
)

const testCount = 100000

type MockExecutor struct {
	count  int
	chTask chan func()
}

func (m *MockExecutor) Run(fn func()) {
	m.chTask <- fn
}

func (m *MockExecutor) start() {
	for {
		select {
		case f, ok := <-m.chTask:
			if !ok {
				return
			}
			f()
			fmt.Println("get", m.count)
			if m.count == testCount {
				return
			}
		}
	}
}
