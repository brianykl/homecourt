package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Function to handle errors
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Data structures for dummy data
type LiveTicketPrice struct {
	TeamID    string  `json:"team_id"`
	GameID    string  `json:"game_id"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Timestamp string  `json:"timestamp"`
}

type Odds struct {
	GameID    string  `json:"game_id"`
	Odds      float64 `json:"odds"`
	Timestamp string  `json:"timestamp"`
}

type PlayerInjury struct {
	PlayerID  string `json:"player_id"`
	Status    string `json:"status"`
	GameID    string `json:"game_id"`
	Timestamp string `json:"timestamp"`
}

type PlayerTransfer struct {
	PlayerID  string `json:"player_id"`
	FromTeam  string `json:"from_team"`
	ToTeam    string `json:"to_team"`
	Status    string `json:"status"` // e.g., "signed", "waived"
	Timestamp string `json:"timestamp"`
}

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "failed to connect to rabbitmq")
	defer conn.Close()

	channel, err := conn.Channel()
	failOnError(err, "failed to open a channel")
	defer channel.Close()

	// Declare Exchange
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
	}
	log.Println("homecourt_exchange and queues set up successfully")
	log.Println("producer started. waiting for tickers...")

	go handleOdds(channel)
	log.Println("producer started. running in background...")
	select {}

}

func publishMessage(channel *amqp.Channel, exchange, routingKey string, data interface{}) {
	body, err := json.Marshal(data)
	failOnError(err, fmt.Sprintf("error marshalling data for %s: %v", routingKey, err))

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
	failOnError(err, fmt.Sprintf("error publishing message to %s: %v", routingKey, err))

	log.Printf("Published message to %s: %s", routingKey, string(body))
}

func handleOdds(channel *amqp.Channel) {
	ticker := time.NewTicker(6 * time.Second) // limited at 10 api calls a min
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 10 * time.Second, // Adjust the timeout as needed
	}

	for range ticker.C {
		resp, err := client.Get("https://api.example.com/ticket_prices?team=team123")
		if resp.StatusCode != http.StatusOK {
			log.Printf("Unexpected status code: %d for odds API", resp.StatusCode)
			continue
		}
		failOnError(err, "failed to get odds from odds blaze")
		body, err := io.ReadAll(resp.Body)
		failOnError(err, "cant read body from odds blaze api call")
		resp.Body.Close()

		var data Odds
		err = json.Unmarshal(body, &data)
		failOnError(err, "error marshalling odds")
		publishMessage(channel, "homecourt_exchange", "odds", data)
	}
}
