package consul

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func RegisterService(cluster string, service string, index int, addr string, port int) error {
	logrus.Infof("consul register service: %s, %s, %d, %s:%d", cluster, service, index, addr, port)
	url := consulUrl + "agent/service/register"
	id := makeServiceId(cluster, service, index)
	args := &map[string]interface{}{
		"Name":    service,
		"Id":      id,
		"Port":    port,
		"Address": addr,
		"Tags":    []string{cluster, strconv.Itoa(index), fmt.Sprintf("gRPC.port=%d", port)},
		"Meta": map[string]string{
			"cluster": cluster,
			"index":   strconv.Itoa(index),
		},
		"Check": map[string]string{
			"Name":                           id,
			"CheckID":                        id,
			"DeregisterCriticalServiceAfter": "8h",
			"Interval":                       "5s",
			"Timeout":                        "1s",
			"Tcp":                            fmt.Sprintf("%s:%d", addr, port),
			"Status":                         "passing",
		},
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	} else {
		return fmt.Errorf("consul return http code: %d", resp.StatusCode)
	}
}
