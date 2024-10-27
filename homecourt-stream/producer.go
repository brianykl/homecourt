package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	ampq "github.com/rabbitmq/amqp091-go"
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
	conn, err := ampq.Dial("amqp://guest:guest@localhost:5672/")
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

	liveTicketTicker := time.NewTicker(30 * time.Second)      // Every 30 seconds
	oddsTicker := time.NewTicker(45 * time.Second)            // Every 45 seconds
	playerInjuriesTicker := time.NewTicker(5 * time.Minute)   // Every 5 minutes
	playerTransfersTicker := time.NewTicker(10 * time.Minute) // Every 10 minutes
	defer liveTicketTicker.Stop()
	defer oddsTicker.Stop()
	defer playerInjuriesTicker.Stop()
	defer playerTransfersTicker.Stop()

	log.Println("producer started. waiting for tickers...")

	for {
		select {
		case t := <-liveTicketTicker.C:
			// Create dummy LiveTicketPrice data
			data := LiveTicketPrice{
				TeamID:    "team123",
				GameID:    "game456",
				Price:     150.00,
				Currency:  "USD",
				Timestamp: t.Format(time.RFC3339),
			}

			publishMessage(channel, "homecourt_exchange", "live_ticket_prices", data)

		case t := <-oddsTicker.C:
			// Create dummy Odds data
			data := Odds{
				GameID:    "game456",
				Odds:      1.95,
				Timestamp: t.Format(time.RFC3339),
			}

			publishMessage(channel, "homecourt_exchange", "odds", data)

		case t := <-playerInjuriesTicker.C:
			// Create dummy PlayerInjury data
			data := PlayerInjury{
				PlayerID:  "player789",
				Status:    "Out",
				GameID:    "game456",
				Timestamp: t.Format(time.RFC3339),
			}

			publishMessage(channel, "homecourt_exchange", "player_injuries", data)

		case t := <-playerTransfersTicker.C:
			// Create dummy PlayerTransfer data
			data := PlayerTransfer{
				PlayerID:  "player321",
				FromTeam:  "team123",
				ToTeam:    "team456",
				Status:    "signed",
				Timestamp: t.Format(time.RFC3339),
			}

			publishMessage(channel, "homecourt_exchange", "player_transfers", data)
		}
	}
}

func publishMessage(channel *ampq.Channel, exchange, routingKey string, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data for %s: %v", routingKey, err)
		return
	}

	err = channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		ampq.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: ampq.Persistent,
		},
	)
	if err != nil {
		log.Printf("Error publishing message to %s: %v", routingKey, err)
		return
	}

	log.Printf("Published message to %s: %s", routingKey, string(body))
}
