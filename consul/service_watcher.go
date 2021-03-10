package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
)

/*
be careful about Service's ready, condition is that we multiple goroutine call Alive, one goroutine will do the
Watch method, but the Watch method is a async method, before the first watch completed(a http call), all user could
not see a single alive node, they have to waiting for the watch complete, so the ready & readyCh come here.
try to comprehend the wait code yourself, baby.
*/

var ServiceMgr sync.Map

type Service struct {
	id     int
	Status string
	Addr   string
	Port   int
}

type ConsulHealth struct {
	Service struct {
		ID      string
		Address string
		Port    int
	}
	Checks []struct {
		ServiceID string
		Status    string
	}
}

func ReadServices(serviceId string) []Service {
	serviceList, ok := ServiceMgr.Load(serviceId)
	if !ok {
		log.Println("not ok list,", serviceList)
		return nil
	}
	serviceList1, ok := serviceList.([]Service)
	if ok {
		return serviceList1
	}
	return nil
}

func ChangeServices(serviceId string, new []Service) {
	ServiceMgr.Store(serviceId, new)
}

func GetRandService(serviceId string) Service {
	serviceList := ReadServices(serviceId)
	if serviceList == nil {
		return Service{}
	}
	length := len(serviceList)
	return serviceList[rand.Intn(length)]
}

func GetService(serviceId string, index int) Service {
	serviceList := ReadServices(serviceId)
	service := Service{}
	for i := 0; i < len(serviceList); i++ {
		if serviceList[i].id == index {
			return serviceList[i]
		}
	}
	return service
}

//
//func (s *Service) waitReady() {
//	if atomic.LoadUint32(&s.ready) == 1 {
//		return
//	}
//
//	<-s.readyCh
//}
//
//func (s *Service) waitReadyTimeout(duration time.Duration) {
//	if atomic.LoadUint32(&s.ready) == 1 {
//		return
//	}
//
//	select {
//	case <-s.readyCh:
//	case <-time.After(duration):
//	}
//}
//
//func (s *Service) setReady() {
//	if atomic.LoadUint32(&s.ready) == 1 {
//		// already ready
//		return
//	}
//
//	if atomic.CompareAndSwapUint32(&s.ready, 0, 1) {
//		// it's me to set ready
//		close(s.readyCh)
//	} else {
//		// do nothing, some one already closed chan
//	}
//}

func getServiceId(cluster string, service string) string {
	return cluster + "-" + service
}

func getCluster(id string) (string, string, int) {
	ss := strings.Split(id, "-")
	index, err := strconv.Atoi(ss[2])
	if err != nil {
		index = 1
	}
	return ss[0], ss[1], index
}

func WatchServer(dc string, cluster string, service string) {
	url := serviceUrl(cluster, service, dc)
	Watch(url, cb, "service")
}

//func (watch *ServiceWatcher) Alive(dc string, cluster string, service string, alter []int) []int {
//	serviceId := getServiceId(cluster, service)
//	// fmt.Println(serviceId)
//	watch.RLock()
//	svs, ok := watch.statusMap[serviceId]
//	if !ok {
//		// if service have not been watched, watch it first,
//		watch.RUnlock()
//		watch.Watch(dc, cluster, service)
//		svs = watch.statusMap[serviceId]
//	} else {
//		watch.RUnlock()
//	}
//	svs.waitReady()
//
//	svs.RLock()
//	defer svs.RUnlock()
//
//	if len(alter) == 0 {
//		alives := make([]int, 0, len(svs.Status))
//		for k := range svs.Status {
//			alives = append(alives, k)
//		}
//		return alives
//	} else {
//		alives := make([]int, 0, len(alter))
//		for _, i := range alter {
//			if _, ok := svs.Status[i]; ok {
//				alives = append(alives, i)
//			}
//		}
//		return alives
//	}
//}

var ServiceNotFound = errors.New("ServiceNotFound")

//func (watch *ServiceWatcher) GetAddress(dc string, cluster string, service string, index int) (Addr string, Port int, err error) {
//	serviceId := getServiceId(cluster, service)
//	var (
//		svs *Service
//		ok  bool
//	)
//	watch.RLock()
//	if svs, ok = watch.statusMap[serviceId]; ok {
//		watch.RUnlock()
//	} else {
//		watch.RUnlock()
//		watch.Watch(dc, cluster, service)
//		svs = watch.statusMap[serviceId]
//	}
//
//	svs.RLock()
//	defer svs.RUnlock()
//
//	if Status, ok := svs.Status[index]; ok {
//		return Status.Addr, Status.Port, nil
//	} else {
//		return "", 0, ServiceNotFound
//	}
//}

func cb(url interface{}, body interface{}) {
	urlS, ok := url.(string)
	if !ok {
		log.Errorf("ServiceCb url to string error: %v", url)
		return
	}
	bodyB, ok := body.([]byte)
	if !ok {
		log.Errorf("ServiceCb body to []byte error: %v", url)
		return
	}
	ServiceCb(urlS, bodyB)
}

func ServiceCb(url string, body []byte) {
	var healthList = &[]ConsulHealth{}
	err := json.Unmarshal(body, healthList)
	if err != nil {
		log.Errorf("defaultServiceCb json unmarshal error: %v", err)
		return
	}
	log.Infof("accept service black query cb, url %v, body %v ", url, healthList)
	if len(*healthList) == 0 {
		return
	}
	var serviceSlice []Service
	var serviceId string
	for _, health := range *healthList {
		if !checkServiceInfo(health) {
			continue
		}
		serviceId = health.Service.ID
		serviceSlice = append(serviceSlice, NewServiceStatusMap(health)...)
	}
	cluster, service, index := getCluster(serviceId)
	ChangeServices(getServiceId(cluster, service), serviceSlice)
	if GetService(getServiceId(cluster, service), index).id == 0 {
		log.Infof("service %s down, del client", service)
		// todo delClient(serviceId)
	}
}

func checkServiceInfo(health ConsulHealth) bool {
	if health.Service.Address == "" {
		return false
	}
	if health.Service.Port == 0 {
		return false
	}
	if len(health.Checks) == 0 {
		return false
	}
	return true
}

func NewServiceStatusMap(health ConsulHealth) []Service {
	var services []Service
	for _, check := range health.Checks {
		if check.ServiceID != "" {
			var index int
			if check.Status != "passing" {
				continue
			}
			_, _, index = getCluster(check.ServiceID)
			service := Service{
				id:     index,
				Addr:   health.Service.Address,
				Port:   health.Service.Port,
				Status: check.Status,
			}
			services = append(services, service)
		}
	}
	return services
}

func serviceUrl(cluster string, service string, dc string) string {
	kvList := []KV{}
	kvList = append(kvList, KV{"dc", dc}, KV{"tag", cluster})
	url := fmt.Sprintf("%shealth/service/%s?", consulUrl, service)
	return UrlAppendParams(url, kvList)
}

type KV struct {
	K string
	V string
}

func UrlAppendParams(uri string, params []KV) string {
	for _, param := range params {
		if param.K == "" || param.V == "" {
		} else {
			uri += fmt.Sprintf("%s=%s&", param.K, param.V)
		}
	}
	return strings.TrimRight(strings.TrimRight(uri, "&"), "?")
}
