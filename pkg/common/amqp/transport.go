package amqp

import (
	"github.com/go-kit/kit/transport/http/jsonrpc"
	"github.com/streadway/amqp"
)

const (
	orderDomainEventsQueueName     = "order_domain_event"
	orderDomainEventsExchangeName  = "domain_event"
	orderDomainEventsExchangeType  = "topic"
	orderDomainEventsRoutingKey    = "#"
	orderDomainEventsRoutingPrefix = "order."
)

type channel struct {
	conn               *amqp.Connection
	writeChannel       *amqp.Channel
	messageReceiveChan chan string
}

func (t *channel) Name() string {
	return orderDomainEventsExchangeName
}

func (t *channel) Send(msgBody string, eventType string) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  jsonrpc.ContentType,
		Body:         []byte(msgBody),
	}
	routingKey := orderDomainEventsRoutingPrefix + eventType
	return t.writeChannel.Publish(orderDomainEventsExchangeName, routingKey, false, false, msg)
}

func (t *channel) Receive() chan string {
	return t.messageReceiveChan
}

func (t *channel) Connect(conn *amqp.Connection) error {
	t.writeChannel = nil

	t.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	t.writeChannel = channel

	err = channel.ExchangeDeclare(orderDomainEventsExchangeName, orderDomainEventsExchangeType, true, false, false, false, nil)
	if err != nil {
		return err
	}

	readQueue, err := channel.QueueDeclare(orderDomainEventsQueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = channel.QueueBind(readQueue.Name, orderDomainEventsRoutingKey, orderDomainEventsExchangeName, false, nil)
	if err != nil {
		return err
	}

	readChan, err := channel.Consume(readQueue.Name, "", true, false, false, false, nil)

	go func() {
		for msg := range readChan {
			t.messageReceiveChan <- string(msg.Body)
		}
	}()

	return err
}

func NewOrderDomainEventsChannel() *channel {
	return &channel{messageReceiveChan: make(chan string)}
}
