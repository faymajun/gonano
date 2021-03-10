// consul blocking queries https://github.com/hashicorp/consul/blob/0d6aff29f63ffa50f2bb0ee28be2d138dab338f9/website/pages/api-docs/features/blocking.mdx
package consul

import (
	"bytes"
	"fmt"
	"github.com/faymajun/gonano/util"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var WatchMgr sync.Map

const notWatch = -1
const unReadyWatch = 0
const readyWatch = 1
const queryWait = "5m"

func checkWatch(path string) int {
	watch, ok := WatchMgr.Load(path)
	if ok && watch == readyWatch {
		return readyWatch
	}
	if !ok {
		return notWatch
	}
	return unReadyWatch
}

func testWatch(path string) bool {
	switch checkWatch(path) {
	case readyWatch:
		return true
	case unReadyWatch:
		time.Sleep(500 * time.Millisecond)
		return testWatch(path)
	case notWatch:
		addWatch(path)
		return false
	}
	return false
}

func addWatch(path string) {
	WatchMgr.Store(path, unReadyWatch)
}

func addReadyWatch(path string) {
	WatchMgr.Store(path, readyWatch)
}

func Watch(path string, callback func(interface{}, interface{}), cbType string) error {
	if testWatch(path) {
		return nil
	}
	query(path, callback, cbType)
	return nil
}

func query(url string, callback func(interface{}, interface{}), cbType string) {
	// It is always safe to use an index of 1 to wait for updates when the data being requested doesn't exist yet
	watchPath := url + fmt.Sprintf("&index=1&wait=%s", queryWait)
	if body, head, err := util.HttpGet(watchPath); err == nil {
		newIndex := consulIndex(head)
		handleCallback(callback, url, body, cbType)
		addReadyWatch(url)
		go blockQuery(url, newIndex, body, callback, cbType)
	} else {
		log.Errorf("consul err: %v", err)
	}
}

func blockQuery(url string, index int, oldBody []byte, callback func(interface{}, interface{}), cbType string) {
	for {
		watchPath := url + fmt.Sprintf("&index=%d&wait=%s", index, queryWait)
		if body, head, err := util.HttpGet(watchPath); err == nil {
			newIndex := consulIndex(head)
			//if newIndex > index { // 已经改变
			//	handleCallback(callback, url, body, cbType)
			//	index = newIndex
			//} else if newIndex == index { // 没有变化
			//	index = newIndex
			//} else {
			//	//* If a raft snapshot is restored on the servers with older version of the data.
			//	//* KV list operations where an item with the highest index is removed.
			//	//* A Consul upgrade changes the way watches work to optimize them with more granular indexes.
			//	index = 0 // 异常：立即重新获取
			//}
			if newIndex > index {
				index = newIndex
			} else {
				index = 0
			}
			if bytes.Compare(oldBody, body) != 0 {
				handleCallback(callback, url, body, cbType)
				oldBody = body
			}
		} else {
			log.Errorf("error get info %#v", err)
			index = 0
		}
	}
}

func handleCallback(callback func(interface{}, interface{}), url string, newConfig []byte, cbType string) {
	// we should guarantee that user provided callback wont destroy our config system
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("consul watch callback panic: %v", r)
		}
	}()
	switch cbType {
	case "kv":
		defaultKvCb(url, newConfig, callback)
	case "service":
		callback(url, newConfig)
	}
}

func consulIndex(head http.Header) int {
	index := head.Get("X-Consul-Index")
	if index == "" {
		return 1
	}
	indexInt, err := strconv.Atoi(index)
	if err != nil {
		return 1
	}
	return indexInt
}
