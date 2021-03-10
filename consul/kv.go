package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/faymajun/gonano/util"
	"github.com/tidwall/gjson"
	"net/http"
	"sync"
)

const consulUrl = "http://127.0.0.1:2000/v1/"

var KvMap sync.Map

func Values(Key string) []byte {
	url := fmt.Sprintf("%skv/%s?raw=true", consulUrl, Key)
	if value := ReadCfgData(url); value == nil {
		kvWatch(url, nil)
	}
	return ReadCfgData(url)
}

func Field(Key, field string) gjson.Result {
	url := fmt.Sprintf("%skv/%s?raw=true", consulUrl, Key)
	if value := ReadCfgData(url); value == nil {
		kvWatch(url, nil)
	}
	values := ReadCfgData(url)
	return gjson.GetBytes(values, field)
}

func StaticCfg() []byte {
	url := StaticCfgUrl()
	if value := ReadCfgData(url); value == nil {
		kvWatch(url, nil)
	}
	return ReadCfgData(url)
}

func DynamicCfg(service string, key string, callback func(interface{}, interface{})) []byte {
	cluster, _, _, _ := GetWhoAmI()
	url := DynamicCfgUrl(cluster, service, key)
	if value := ReadCfgData(url); value == nil {
		kvWatch(url, callback)
	}
	return ReadCfgData(url)
}

func DynamicCfgCustom(url string, callback func(interface{}, interface{})) []byte {
	url1 := DynamicCfgUrlCustom(url)
	if value := ReadCfgData(url1); value == nil {
		kvWatch(url1, callback)
	}
	return ReadCfgData(url1)
}

func SetDynamicCfg(k string, data interface{}) (err error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	cluster, service, _, _ := GetWhoAmI()
	path := DynamicCfgUrl(cluster, service, k)
	url := consulUrl + path
	req, err := http.NewRequest("PUT", url, bytes.NewReader(raw))
	if err != nil {
		return
	}
	if resp, err := http.DefaultClient.Do(req); err != nil {
		return err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return nil
		} else {
			return fmt.Errorf("consul return http code: %d", resp.StatusCode)
		}
	}
}

func ReadDataDirectly(service string, key string) ([]byte, error) {
	cluster, _, _, _ := GetWhoAmI()
	path := DynamicCfgUrl(cluster, service, key)
	value := ReadCfgData(path)
	if value == nil {
		body, _, err := util.HttpGet(path)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return value, nil
}

func WhoAmI(cluster string, service string, index int) {
	MySelf.Cluster = cluster
	MySelf.Service = service
	MySelf.Index = index
}

type dcBody struct {
	Config struct {
		Datacenter string `json:"Datacenter"`
	} `json:"Config"`
}

func Dc() (string, error) {
	if MySelf.Dc != "" {
		return MySelf.Dc, nil
	}
	url := consulUrl + "agent/self"
	body, _, err := util.HttpGet(url)
	if err != nil {
		return "", err
	}
	var dcInfo = &dcBody{}
	err = json.Unmarshal(body, dcInfo)
	if err != nil {
		return "", err
	}
	MySelf.Dc = dcInfo.Config.Datacenter
	return dcInfo.Config.Datacenter, nil
}

func GetWhoAmI() (string, string, int, string) {
	return MySelf.Cluster, MySelf.Service, MySelf.Index, MySelf.Dc
}

func ReadCfgData(path string) []byte {
	cfg, _ := KvMap.Load(path)
	cfg1, ok := cfg.([]byte)
	if ok {
		return cfg1
	}
	return nil
}

func ChangeCfgData(path string, new []byte) {
	KvMap.Store(path, new)
}

func StaticCfgUrl() string {
	cluster, service, index, _ := GetWhoAmI()
	return fmt.Sprintf("%skv/app_static_cfg/%s/%s/%d?raw=true", consulUrl, cluster, service, index)
}

func DynamicCfgUrl(cluster string, service string, key string) string {
	return fmt.Sprintf("%skv/app_dynamic_cfg/%s/%s/%s?raw=true", consulUrl, cluster, service, key)
}

func DynamicCfgUrlCustom(url string) string {
	return fmt.Sprintf("%s%s?raw=true", consulUrl, url)
}

func kvWatch(path string, callback func(interface{}, interface{})) {
	body, head, err := util.HttpGet(path)
	if err != nil {
		log.Errorf("kvMatch error, path=%s, error=%v", path, err)
		return
	}
	defaultKvCb(path, body, callback)
	Index := consulIndex(head)
	go blockQuery(path, Index, body, callback, "kv")
}

func defaultKvCb(path string, body []byte, customCb func(interface{}, interface{})) {
	ChangeCfgData(path, body)
	if customCb != nil {
		customCb(path, body)
	}
}
