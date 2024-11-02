package producers

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func publishMessage(channel *amqp.Channel, exchange, routingKey string, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("error marshalling data for %s: %v", routingKey, err)
		return
	}

	err = channel.Publish(
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
	if err != nil {
		log.Printf("error publishing message to %s: %v", routingKey, err)
		return
	}

	log.Printf("published message to %s: %s", routingKey, string(body))
}
