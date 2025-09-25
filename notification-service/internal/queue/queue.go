package queue

import (
	"database/sql"
	"errors"
	
	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// Queue holds connections for DB, Redis, and RabbitMQ
type Queue struct {
	DB           *sql.DB
	RabbitMQConn *amqp.Connection
	RedisConn    *redis.Client
}

// RabbitMQConnection represents a RabbitMQ connection and channel
type RabbitMQConnection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	err     chan error
}

// NewRabbitConnection creates a simple RabbitMQ connection for a queue
func NewRabbitConnection(queueName string) *RabbitMQConnection {
	return &RabbitMQConnection{
		queue: queueName,
		err:   make(chan error),
	}
}

// InitQueue sets up the notification queue and starts consuming
func (q *Queue) InitQueue() {
	rmq := NewRabbitConnection("notification_queue")

	if err := rmq.Connect(q.RabbitMQConn); err != nil {
		panic(err)
	}

	if err := rmq.DeclareQueue(); err != nil {
		panic(err)
	}

	go func() {
		if err := rmq.Consume(q.processNotificationQueue); err != nil {
			logrus.Errorf("Error consuming queue: %v", err)
		}
	}()

	select {} // keep process alive
}

// Connect opens a channel on RabbitMQ
func (r *RabbitMQConnection) Connect(conn *amqp.Connection) error {
	r.conn = conn
	ch, err := r.conn.Channel()
	if err != nil {
		return err
	}
	r.channel = ch

	go func() {
		<-r.channel.NotifyClose(make(chan *amqp.Error))
		r.err <- errors.New("channel closed")
	}()

	return nil
}

// DeclareQueue declares the notification queue
func (r *RabbitMQConnection) DeclareQueue() error {
	_, err := r.channel.QueueDeclare(r.queue, true, false, false, false, nil)
	return err
}

// Consume starts consuming messages from the queue
func (r *RabbitMQConnection) Consume(handler func([]byte) error) error {
	deliveries, err := r.channel.Consume(r.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range deliveries {
		if err := handler(d.Body); err != nil {
			d.Nack(false, false)
		} else {
			d.Ack(false)
		}
	}

	return nil
}


