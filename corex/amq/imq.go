package amq

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/golang/protobuf/proto"
)

var (
	_session *Session
)

func init() {
	_session = nil
}

func Init(name string, host string, port int32, username string, password string,
	durable bool, autoDelete bool, exclusive bool, noWait bool) error {
	if _session != nil {
		return errors.New("duplicate create!!!")
	}
	_session = New(name, host, port, username, password, durable, autoDelete, exclusive, noWait)
	return nil
}

func Finish() error {
	if _session == nil {
		return errors.New("invalid amq session")
	}

	return _session.Close()
}

func Push(queue string, msg proto.Message) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		log.Error("json marshal failed: ", err)
		return err
	}
	fmt.Println(string(raw))

	if err := _session.Push(raw, queue); err != nil {
		log.Error("pushSuccess failed: ", err)
		return err
	}
	return nil
}
