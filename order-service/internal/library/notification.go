package library

import (
	"encoding/json"
	"fmt"
	"log"
	"savannah-store/order-service/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)



// SendSMS pushes SMS notification to RabbitMQ
func Notification(rabbitConn *amqp.Connection, notify models.Notification) error {
	// confirm if mandatory fields are present
	if notify.Type == "" || notify.To == "" || notify.Message == "" {
		return fmt.Errorf("missing mandatory fields in notification")
	}
	log.Printf("Preparing to send notification: %+v", notify)
	err :=pushToQueue(rabbitConn, "notification", notify)
	if err != nil {
		log.Printf("Error pushing notification to queue: %v", err)
		return fmt.Errorf("error pushing notification to queue: %v", err)
	}
	return err
}


// pushToQueue serializes and publishes message to RabbitMQ
func pushToQueue(rabbitConn *amqp.Connection, queueName string, payload interface{}) error {
	ch, err := rabbitConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	err = ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Notification pushed to queue %s: %s", queueName, string(body))
	return nil
}
