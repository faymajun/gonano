package rabbitmq

import (
	"fmt"

	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var logger = logrus.WithField("component", "rabbitmq")
var Producer = &producer{queues: make(map[string]bool)}

func failOnError(err error, msg string) {
	if err != nil {
		logger.Fatalf("%s: %s", msg, err)
	}
}

type producer struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	queues map[string]bool
}

// 初始化生产者
func InitProducer(addr string) {
	logger.Infof("RabbitMQ producer init. addr:%s", addr)
	conn, err := amqp.Dial(addr)
	failOnError(err, "Failed to connect to RabbitMQ")
	Producer.conn = conn

	ch, err := Producer.conn.Channel()
	failOnError(err, "Failed to open a channel")
	Producer.ch = ch
}

func AddQueues(name string) error {
	if Producer.queues[name] {
		return fmt.Errorf("%s: %s", name, "duplicated to declare a queue")
	}
	_, err := Producer.ch.QueueDeclare(
		name,  // name
		false, // durable            // 聊天消息不用持久化
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("%s: %s", err, "Failed to declare a queue")
	}
	Producer.queues[name] = true
	return nil
}

func PublishMsg(name string, msgid message.MSGID, pbmsg proto.Message) error {
	packet := packet.SendMessage{MsgID: msgid, Payload: pbmsg}
	payload, err := packet.Serialize()
	if err != nil {
		return fmt.Errorf("RabbitMQ PublishMsg serialize msg:%d failed:%s", msgid, err)
	}
	body := coder.Encode(msgid, payload)

	err = Producer.ch.Publish(
		"",    // exchange
		name,  // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Body: body,
		})

	return err
}

func StopProducer() {
	logger.Info("RabbitMQ producer Close!!!")
	Producer.ch.Close()
	Producer.conn.Close()
}
