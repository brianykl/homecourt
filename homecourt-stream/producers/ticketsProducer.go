package producers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func HandleTickets(channel *amqp.Channel) {
	apiKey := os.Getenv("TICKETMASTERKEY")
	if apiKey == "" {
		log.Fatal("TICKETMASTERKEY hasn't been set")
	}

	apiSecret := os.Getenv("TICKETMASTERSECRET")
	if apiSecret == "" {
		log.Fatal("TICKETMASTERSECRET hasn't been set")
	}

	ticker := time.NewTicker(20 * time.Second) // rate limit is 3.5 api calls per minute
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// need to do some kind of ordering here so that we know which games to look up
	apiURL := fmt.Sprintf("https://data.oddsblaze.com/v1/odds/espn_bet_nba.json?key=%s&market=Moneyline&live=false", apiKey)
	for range ticker.C {
		log.Printf(apiURL)
		client.Get(apiURL)
	}
	publishMessage(channel, "homecourt-exchange", "tickets", nil)
}
