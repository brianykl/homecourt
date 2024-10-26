package main

import (
	"context"
	"log"
	"time"

	ampq "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := ampq.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "failed to connect to rabbitmq")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer conn.Close()

	queue, err := channel.QueueDeclare(
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "hello world"
	err = channel.PublishWithContext(
		ctx,
		"",
		queue.Name,
		false,
		false,
		ampq.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "failed to publish a message")
	log.Printf(" [x] sent %s\n", body)

}
