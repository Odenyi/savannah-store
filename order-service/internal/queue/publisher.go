package queue

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare("user.events", "topic", true, false, false, false, nil); err != nil {
		log.Printf("exchange declare failed: %v", err)
	}
	return &Publisher{ch: ch}, nil
}

func (p *Publisher) Publish(routing string, payload interface{}) error {
	b, _ := json.Marshal(payload)
	return p.ch.Publish("user.events", routing, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        b,
	})
}
