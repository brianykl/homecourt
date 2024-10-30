package producers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Odds struct {
}

func HandleOdds(channel *amqp.Channel) {
	oddsBlazeKey := os.Getenv("ODDSBLAZEKEY")
	ticker := time.NewTicker(6 * time.Second) // limited at 10 api calls a min
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 10 * time.Second, // Adjust the timeout as needed
	}

	for range ticker.C {
		resp, _ := client.Get(fmt.Sprintf("https://data.oddsblaze.com/v1/odds/espn_bet_nba.json?key=%s&market=Moneyline&live=false", oddsBlazeKey))
		if resp.StatusCode != http.StatusOK {
			log.Printf("Unexpected status code: %d for odds API", resp.StatusCode)
			continue
		}
		// failOnError(err, "failed to get odds from odds blaze")
		body, _ := io.ReadAll(resp.Body)
		// failOnError(err, "cant read body from odds blaze api call")
		resp.Body.Close()

		var data Odds
		_ = json.Unmarshal(body, &data)
		// failOnError(err, "error marshalling odds")
		publishMessage(channel, "homecourt_exchange", "odds", data)
	}
}
