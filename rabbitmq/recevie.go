package rabbitmq

import (
	"github.com/faymajun/gonano/core/net"
	"github.com/faymajun/gonano/core/scheduler"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

var Consumer = &consumer{}
var coder = net.OrdinaryCodecFactory()

// 消费者
type consumer struct {
	chDead chan struct{}
}

func InitConsumer(address string, name string) {
	logger.Infof("RabbitMQ consumer init. addr:%s, name:%s", address, name)
	conn, err := amqp.Dial(address)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		name,  // name
		false, // durable             // 聊天消息不用持久化
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

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

	Consumer.chDead = make(chan struct{})

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
				errs := errors.New("RabbitMQ receiver msg error!")
				logger.Error(errs)
			} else {
				pack, err := coder.Decode(msg.Body[net.MsgHeadSize:])
				if err != nil {
					logger.Errorf("RabbitMQ HandleMessage decode data err=%s", err)
				} else {
					scheduler.PushPacket(&pack)
				}
			}
		case <-Consumer.chDead:
			break Loop
		}
	}

}

func (c *consumer) Close() {
	logger.Info("RabbitMQ consumer Close!!!")
	close(c.chDead)
}
