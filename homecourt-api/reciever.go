package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Reciever() {
	// Connect to RabbitMQ
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}

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
	queues := []string{"live_ticket_prices", "odds", "player_injuries", "player_transfers"}
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
			true,
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
			}
		}(queueName, messages)
	}
	log.Println("homecourt_exchange and queues set up successfully")

	log.Printf(" [x] waiting for messages. to exit press ctrl + c")
	select {}
}
