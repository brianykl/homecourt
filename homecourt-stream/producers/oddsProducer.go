package producers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OddsResponse struct {
	Games []Game `json:"games"`
}

// Game represents a single game in the OddsResponse.
type Game struct {
	Teams       Teams        `json:"teams"`
	Start       string       `json:"start"`
	Sportsbooks []Sportsbook `json:"sportsbooks"`
}

// Teams contains the home and away teams for a game.
type Teams struct {
	Away Team `json:"away"`
	Home Team `json:"home"`
}

// Team represents a team with a name.
type Team struct {
	Name string `json:"name"`
}

// Sportsbook represents a sportsbook offering odds on a game.
type Sportsbook struct {
	Odds []Odd `json:"odds"`
}

// Odd represents the betting odds for a team.
type Odd struct {
	Selection string `json:"selection"` // "Home" or "Away"
	Price     string `json:"price"`     // e.g., "-190", "+155"
}

// OddsMessage is the simplified message to be published to RabbitMQ.
type OddsMessage struct {
	AwayTeam      string            `json:"away_team"`
	HomeTeam      string            `json:"home_team"`
	StartTime     string            `json:"start_time"`
	BettingPrices map[string]string `json:"betting_prices"` // Map of team name to price
}

// Helper functions

// parseOddsJSON takes a JSON byte slice and unmarshals it into an OddsResponse struct.
func parseOddsJSON(jsonData []byte) (*OddsResponse, error) {
	var response OddsResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// extractOddsMessages processes the OddsResponse and returns a slice of OddsMessage.
func extractOddsMessages(response *OddsResponse) []OddsMessage {
	var messages []OddsMessage

	for _, game := range response.Games {
		// Parse the game start time
		gameTime, err := time.Parse(time.RFC3339, game.Start)
		if err != nil {
			log.Printf("Error parsing game start time: %v", err)
			continue
		}
		formattedStartTime := gameTime.Format("Monday, Jan 2, 2006 at 3:04pm")

		// Initialize the OddsMessage
		message := OddsMessage{
			AwayTeam:      game.Teams.Away.Name,
			HomeTeam:      game.Teams.Home.Name,
			StartTime:     formattedStartTime,
			BettingPrices: make(map[string]string),
		}

		// Process the odds from the first sportsbook
		if len(game.Sportsbooks) > 0 {
			sportsbook := game.Sportsbooks[0]
			for _, odd := range sportsbook.Odds {
				selection := strings.ToLower(odd.Selection)
				if selection == "home" {
					message.BettingPrices[message.HomeTeam] = odd.Price
				} else if selection == "away" {
					message.BettingPrices[message.AwayTeam] = odd.Price
				}
			}
		} else {
			log.Printf("No sportsbooks found for game: %s vs %s", message.AwayTeam, message.HomeTeam)
		}

		messages = append(messages, message)
	}

	return messages
}

func HandleOdds(channel *amqp.Channel) {
	apiKey := os.Getenv("ODDSBLAZEKEY")
	if apiKey == "" {
		log.Fatal("ODDSBLAZEKEY hasn't been set")
	}

	ticker := time.NewTicker(8 * time.Second) // rate limit is 10 api calls per minute
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := fmt.Sprintf("https://data.oddsblaze.com/v1/odds/espn_bet_nba.json?key=%s&market=Moneyline&live=false", apiKey)

	for range ticker.C {
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Printf("failed to get odds from oddsblaze: %v", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("unexpected status code %d for oddsblaze api", resp.StatusCode)
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("can't read body from oddsblaze api response: %v", err)
			continue
		}

		oddsResponse, err := parseOddsJSON(bodyBytes)
		if err != nil {
			log.Printf("error unmarshalling odds: %v", err)
			continue
		}

		oddsMessages := extractOddsMessages(oddsResponse)
		for _, message := range oddsMessages {
			publishMessage(channel, "homecourt_exchange", "odds", message)
		}

	}

}
