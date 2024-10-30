package producers

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func publishMessage(channel *amqp.Channel, exchange, routingKey string, data interface{}) {
	body, _ := json.Marshal(data)
	// failOnError(err, fmt.Sprintf("error marshalling data for %s: %v", routingKey, err))

	_ = channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	// failOnError(err, fmt.Sprintf("error publishing message to %s: %v", routingKey, err))

	log.Printf("Published message to %s: %s", routingKey, string(body))
}
