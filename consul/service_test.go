package consul

import (
	"fmt"
	"github.com/faymajun/gonano/util"
	"net/url"
	"strings"
	"testing"
	"time"
)

const URI = "http://127.0.0.1:2000/v1/health/service/game-gateway-server?dc=dc1"

func TestHealth(t *testing.T) {
	uri, err := url.Parse(URI)
	if err != nil {
		t.Log(err.Error())
		return
	}
	paths := strings.Split(uri.Path, "/")
	data := make(map[string]string)
	for _, d := range strings.Split(uri.RawQuery, "&") {
		dd := strings.Split(d, "=")
		data[dd[0]] = strings.Join(dd[1:], "=")
	}
	fmt.Println(paths[len(paths)-1], data["tag"])
}

func TestRegisterService(t *testing.T) {
	RegisterService("dc1", "game-gateway-server", 1, util.LocalIPString(), 2222)
	time.Sleep(time.Hour)
}
