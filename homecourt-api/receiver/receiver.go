package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Receiver(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("reciever panicked: %v", r)
		}
	}()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "failed to connect to rabbitmq")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer channel.Close()

	err = channel.ExchangeDeclare(
		"homecourt_exchange", // name
		"direct",             // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	failOnError(err, "failed to declare exchange")
	log.Printf("homecourt_exchange declared successfully")

	// Declare Queues and Bindings
	queues := []string{"tickets", "odds", "injuries"}
	for _, queueName := range queues {
		_, err := channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		failOnError(err, fmt.Sprintf("failed to declare queue: %s", queueName))
		log.Printf("successfully declared queue: %s", queueName)

		err = channel.QueueBind(
			queueName,            // queue name
			queueName,            // routing key
			"homecourt_exchange", // exchange
			false,
			nil,
		)
		failOnError(err, fmt.Sprintf("failed to bind queue %s to homecourt_exchange", queueName))
		log.Printf("successfully bound queue %s to homecourt_exchange with routing key %s", queueName, queueName)

		messages, err := channel.Consume(
			queueName,
			queueName,
			false,
			false,
			false,
			false,
			nil,
		)
		failOnError(err, fmt.Sprintf("failed to register consumer %s", queueName))
		log.Printf("successfully registered consumer for queue: %s", queueName)

		go func(queue string, messages <-chan amqp.Delivery) {
			for d := range messages {
				log.Printf(" [x] %s", d.Body)

				var parsedData map[string]interface{}
				err := json.Unmarshal(d.Body, &parsedData)
				if err != nil {
					log.Printf("error parsing message from %s queue: %v", queue, err)
					continue
				}

				// err = storeData(queue, parsedData)
				//
				// {"away_team":"Minnesota Timberwolves","home_team":"Sacramento Kings","start_time":"Saturday, Nov 16, 2024 at 3:00am","betting_prices":{"Minnesota Timberwolves":"-105","Sacramento Kings":"-115"}

				d.Ack(false)
			}
		}(queueName, messages)
	}
	<-ctx.Done()
	log.Println("Receiver has been stopped")
}

func storeData(queue string, data map[string]interface{}) error {
	panic("unimplemented")
}
