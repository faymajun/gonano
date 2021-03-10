package rabbitmq

import (
	"fmt"
	"github.com/faymajun/gonano/core/net"
	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/core/scheduler"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type agent struct {
	chDead chan struct{}
	conn   *amqp.Connection
	ch     *amqp.Channel
}

var Agent *agent

func InitAgent(address string, name string) {
	logger.Infof("RabbitMQ agent init. addr:%s, name:%s", address, name)
	conn, err := amqp.Dial(address)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		name,     // name
		"fanout", // type
		false,    // durable             // 聊天消息不用持久化
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a Exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		name,   // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to register a QueueBind")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	Agent = &agent{conn: conn, ch: ch, chDead: make(chan struct{})}

Loop:
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				logger.Infoln("RabbitMQ channel close")
				break Loop
			}

			// 解析消息
			if len(msg.Body) <= 4 {
				errs := errors.New("RabbitMQ agent msg error!")
				logger.Error(errs)
			} else {
				pack, err := coder.Decode(msg.Body[net.MsgHeadSize:])
				if err != nil {
					logger.Errorf("RabbitMQ HandleMessage decode data err=%s", err)
				} else {
					scheduler.PushPacket(&pack)
				}
			}
		case <-Agent.chDead:
			break Loop
		}
	}
	logger.Info("RabbitMQ agent Loop Quit!!!")
}

func (a *agent) PublishMsg(name string, msgid message.MSGID, pbmsg proto.Message) error {
	packet := packet.SendMessage{MsgID: msgid, Payload: pbmsg}
	payload, err := packet.Serialize()
	if err != nil {
		return fmt.Errorf("RabbitMQ PublishMsg serialize msg:%d failed:%s", msgid, err)
	}
	body := coder.Encode(msgid, payload)

	err = a.ch.Publish(
		name,  // exchange
		"",    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Body: body,
		})

	return err
}

func (a *agent) Close() {
	logger.Info("RabbitMQ agent Close!!!")
	a.ch.Close()
	a.conn.Close()
	close(a.chDead)
}
