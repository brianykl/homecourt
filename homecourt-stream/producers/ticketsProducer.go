package producers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TicketmasterResponse struct {
	Embedded EmbeddedEvents `json:"_embedded"`
}

// EmbeddedEvents holds the embedded events
type EmbeddedEvents struct {
	Events []Event `json:"events"`
}

// Event represents a single event
type Event struct {
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	ID          string       `json:"id"`
	Dates       Dates        `json:"dates"`
	PriceRanges []PriceRange `json:"priceRanges"`
	Embedded    struct {
		Venues []Venue `json:"venues"`
	} `json:"_embedded"`
}

// Dates contains date information
type Dates struct {
	Start Start `json:"start"`
}

// Start contains the start time information
type Start struct {
	LocalDate      string `json:"localDate"`
	LocalTime      string `json:"localTime"`
	DateTime       string `json:"dateTime"`
	DateTBD        bool   `json:"dateTBD"`
	DateTBA        bool   `json:"dateTBA"`
	TimeTBA        bool   `json:"timeTBA"`
	NoSpecificTime bool   `json:"noSpecificTime"`
}

// PriceRange represents a price range for an event
type PriceRange struct {
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

// Venue represents the venue information
type Venue struct {
	Name string `json:"name"`
	// Add other fields as needed
}

// TicketMessage is the simplified message to be published to RabbitMQ
type TicketMessage struct {
	EventName      string  `json:"event_name"`
	StartDateTime  string  `json:"start_date_time"`
	MinTicketPrice float64 `json:"min_ticket_price"`
	VenueName      string  `json:"venue_name"`
}

type TeamInfo struct {
	TeamName string
	City     string
}

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

	var nbaTeams = []TeamInfo{
		{TeamName: "Atlanta Hawks", City: "Atlanta"},
		{TeamName: "Boston Celtics", City: "Boston"},
		{TeamName: "Brooklyn Nets", City: "Brooklyn"},
		{TeamName: "Charlotte Hornets", City: "Charlotte"},
		{TeamName: "Chicago Bulls", City: "Chicago"},
		{TeamName: "Cleveland Cavaliers", City: "Cleveland"},
		{TeamName: "Dallas Mavericks", City: "Dallas"},
		{TeamName: "Denver Nuggets", City: "Denver"},
		{TeamName: "Detroit Pistons", City: "Detroit"},
		{TeamName: "Golden State Warriors", City: "San Francisco"},
		{TeamName: "Houston Rockets", City: "Houston"},
		{TeamName: "Indiana Pacers", City: "Indianapolis"},
		{TeamName: "Los Angeles Lakers", City: "Los Angeles"},
		{TeamName: "Los Angeles Clippers", City: "Los Angeles"},
		{TeamName: "Memphis Grizzlies", City: "Memphis"},
		{TeamName: "Miami Heat", City: "Miami"},
		{TeamName: "Milwaukee Bucks", City: "Milwaukee"},
		{TeamName: "Minnesota Timberwolves", City: "Minneapolis"},
		{TeamName: "New Orleans Pelicans", City: "New Orleans"},
		{TeamName: "New York Knicks", City: "New York"},
		{TeamName: "Oklahoma City Thunder", City: "Oklahoma City"},
		{TeamName: "Orlando Magic", City: "Orlando"},
		{TeamName: "Philadelphia 76ers", City: "Philadelphia"},
		{TeamName: "Phoenix Suns", City: "Phoenix"},
		{TeamName: "Portland Trail Blazers", City: "Portland"},
		{TeamName: "Sacramento Kings", City: "Sacramento"},
		{TeamName: "San Antonio Spurs", City: "San Antonio"},
		{TeamName: "Toronto Raptors", City: "Toronto"},
		{TeamName: "Utah Jazz", City: "Salt Lake City"},
		{TeamName: "Washington Wizards", City: "Washington"},
	}

	teamIndex := 0
	teamCount := len(nbaTeams)

	// need to do some kind of ordering here so that we know which games to look up
	for range ticker.C {

		teamInfo := nbaTeams[teamIndex]
		teamName := teamInfo.TeamName
		cityName := teamInfo.City

		// URL-encode the parameters
		encodedCity := url.QueryEscape(cityName)
		encodedTeam := url.QueryEscape(teamName)

		// Build the API URL with the current city and team name
		apiURL := fmt.Sprintf("https://app.ticketmaster.com/discovery/v2/events.json?apikey=%s&classificationName=NBA&city=%s&keyword=%s", apiKey, encodedCity, encodedTeam)
		// log.Printf(apiURL)
		resp, err := client.Get(apiURL)
		if err != nil {
			log.Printf("failed to get events from ticketmaster: %v", err)
			teamIndex = (teamIndex + 1) % teamCount
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("unexpected status code %d from ticketmaster api", resp.StatusCode)
			resp.Body.Close()
			teamIndex = (teamIndex + 1) % teamCount
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("can't read body from ticketmaster api response: %v", err)
			teamIndex = (teamIndex + 1) % teamCount
			continue
		}

		ticketmasterResponse, err := parseTicketmasterJSON(bodyBytes)
		if err != nil {
			log.Printf("error unmarshalling ticketmaster response: %v", err)
			teamIndex = (teamIndex + 1) % teamCount
			continue
		}

		ticketMessages := extractTicketMedssages(ticketmasterResponse)
		for _, message := range ticketMessages {
			publishMessage(channel, "homecourt_exchange", "tickets", message)
		}
		teamIndex = (teamIndex + 1) % teamCount

	}
}

func extractTicketMedssages(response *TicketmasterResponse) []TicketMessage {
	var messages []TicketMessage

	for _, event := range response.Embedded.Events {
		var message TicketMessage

		// Extract event ID
		message.EventName = event.Name

		// Extract start date and time
		if event.Dates.Start.DateTime != "" {
			message.StartDateTime = event.Dates.Start.DateTime
		} else if event.Dates.Start.LocalDate != "" && event.Dates.Start.LocalTime != "" {
			message.StartDateTime = fmt.Sprintf("%sT%s", event.Dates.Start.LocalDate, event.Dates.Start.LocalTime)
		} else {
			message.StartDateTime = ""
		}

		// Extract minimum ticket price
		if len(event.PriceRanges) > 0 {
			priceRange := event.PriceRanges[0]
			message.MinTicketPrice = priceRange.Min
		} else {
			// Handle events without priceRanges
			message.MinTicketPrice = 0
		}

		// Extract venue name
		if len(event.Embedded.Venues) > 0 {
			message.VenueName = event.Embedded.Venues[0].Name
		} else {
			message.VenueName = ""
		}

		messages = append(messages, message)
	}

	return messages
}

func parseTicketmasterJSON(jsonData []byte) (*TicketmasterResponse, error) {
	var response TicketmasterResponse
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
