package amqp

import (
	stderrors "errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/ispringteam/go-patterns/infrastructure/log"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Config contains parameters for AMQP connection
type Config struct {
	User     string
	Password string
	Host     string
}

func NewAMQPConnection(cfg *Config, logger log.Logger) *Connection {
	return &Connection{cfg: cfg, logger: logger}
}

type Channel interface {
	Name() string
	Send(msgBody string, eventType string) error
	Receive() chan string
	Connect(conn *amqp.Connection) error
}

var (
	errNilAMQPConnection    = stderrors.New("amqp connection is empty")
	errClosedAMQPConnection = stderrors.New("amqp connection is closed")
)

type Connection struct {
	cfg      *Config
	amqpConn *amqp.Connection
	logger   log.Logger
	channels []Channel
}

/*
* Start recreates AMQP connection, its channels and queues
* loosely based on https://github.com/isayme/go-amqp-reconnect/blob/master/rabbitmq/rabbitmq.go
* but using the fact that connErrorChan and all AMQP channels and queues are closed when the connection is closed
* also we shouldn't try to reconnect from goroutines other than NotifyClose listener to avoid races
* it's better to fail requests until the connection is fully restored than dealing with multiple active connections
 */
func (c *Connection) Start() error {
	url := fmt.Sprintf("amqp://%s:%s@%s/", c.cfg.User, c.cfg.Password, c.cfg.Host)

	err := backoff.Retry(func() error {
		connection, cErr := amqp.Dial(url)
		c.amqpConn = connection
		return errors.Wrap(cErr, "failed to connect to amqp")
	}, newBackOff())

	if err == nil {
		if err = c.validateConnection(c.amqpConn); err != nil {
			return err
		}

		for _, channel := range c.channels {
			if err = channel.Connect(c.amqpConn); err != nil {
				return err
			}
		}

		connErrorChan := c.amqpConn.NotifyClose(make(chan *amqp.Error))
		go c.processConnectErrors(connErrorChan)
	}
	return err
}

func (c *Connection) Close() error {
	return c.amqpConn.Close()
}

func (c *Connection) AddChannel(channel Channel) {
	c.channels = append(c.channels, channel)
}

func (c *Connection) validateConnection(conn *amqp.Connection) error {
	if conn == nil {
		return errors.WithStack(errNilAMQPConnection)
	}
	if conn.IsClosed() {
		return errors.WithStack(errClosedAMQPConnection)
	}
	return nil
}

// channel will be closed then the connection is closed so this function will exit, no need for custom graceful shutdown
func (c *Connection) processConnectErrors(ch chan *amqp.Error) {
	err := <-ch
	if err == nil {
		return
	}

	c.logger.Error(err, "AMQP connection error, trying to reconnect")
	for {
		err := c.Start()
		if err == nil {
			c.logger.Info("AMQP connection restored")
			break
		} else {
			c.logger.Error(err, "failed to reconnect to AMQP")
		}
	}
}

func newBackOff() backoff.BackOff {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = 60 * time.Second
	exponentialBackOff.MaxInterval = 5 * time.Second
	return exponentialBackOff
}
