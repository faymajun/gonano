package task

import (
	"fmt"
	"runtime"

	"github.com/faymajun/gonano/core/scheduler"

	"github.com/pkg/errors"
)

const enableTaskSource = false

// 表示一个异步任务, 异步任务执行完后, 会调用ResultHandler
type Future interface {
	Then(ResultHandler) Task
}

// 可以执行的任务
type Task interface {
	Run() error
}

var ErrAsyncTaskHasSet = errors.New("异步任务已设置")

// 异步任务
type AsyncJob func() (interface{}, error)
type ResultHandler func(interface{}, error)

// 执行一个异步任务, 然后再TaskManager线程中执行所有的同步任务
type task struct {
	source string        // 任务来源
	job    AsyncJob      // 异步作业
	then   ResultHandler // 作业执行完成后的回调函数
}

func (t *task) async(j AsyncJob) *task {
	if t.job != nil {
		panic(ErrAsyncTaskHasSet)
	}
	t.job = j
	return t
}

func New(job AsyncJob) Future {
	t := &task{}
	if enableTaskSource {
		_, file, line, _ := runtime.Caller(1)
		t.source = fmt.Sprintf("%s:%d", file, line)
	}
	return t.async(job)
}

func (t *task) Then(h ResultHandler) Task {
	t.then = h
	return t
}

// 执行任务
func (t *task) Run() error {
	if t.job == nil {
		return fmt.Errorf("没有指定异步任务, Source=%s", t.source)
	}

	if t.then == nil {
		return fmt.Errorf("没有指定异步后续任务, Source=%s", t.source)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					scheduler.PushTask(func() { t.then(nil, e) })
				} else {
					scheduler.PushTask(func() { t.then(nil, fmt.Errorf("aync task error: %v", err)) })
				}
			}
		}()

		r, err := t.job()
		scheduler.PushTask(func() { t.then(r, err) })
	}()

	return nil
}
