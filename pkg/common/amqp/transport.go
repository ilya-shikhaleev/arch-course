package amqp

import (
	"log"

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
	forRetrieve        bool
}

func (c *channel) Name() string {
	return orderDomainEventsExchangeName
}

func (c *channel) Send(msgBody string, eventType string) error {
	log.Println("sent", msgBody, " to", orderDomainEventsRoutingPrefix+eventType)
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  jsonrpc.ContentType,
		Body:         []byte(msgBody),
	}
	routingKey := orderDomainEventsRoutingPrefix + eventType
	return c.writeChannel.Publish(orderDomainEventsExchangeName, routingKey, false, false, msg)
}

func (c *channel) Receive() chan string {
	return c.messageReceiveChan
}

func (c *channel) Connect(conn *amqp.Connection) error {
	c.writeChannel = nil

	c.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	c.writeChannel = channel

	if c.forRetrieve {
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
				log.Println("read message from rabbit", string(msg.Body))
				c.messageReceiveChan <- string(msg.Body)
			}
		}()
	}

	return err
}

func NewOrderDomainEventsChannel(forRetrieve bool) *channel {
	return &channel{messageReceiveChan: make(chan string), forRetrieve: forRetrieve}
}
