package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"homecourt-api/games"
	"log"
	"regexp"
	"strings"
	"time"

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

				err = storeData(ctx, queue, parsedData)
				if err != nil {
					log.Printf("error storing message from %s queue: %v", queue, err)
					continue
				}

				// {"event_name":"Atlanta Hawks vs Miami Heat","start_date_time":"2025-02-25T00:30:00Z","min_ticket_price":25,"venue_name":"State Farm Arena"}
				// {"away_team":"Minnesota Timberwolves","home_team":"Sacramento Kings","start_time":"Saturday, Nov 16, 2024 at 3:00am","betting_prices":{"Minnesota Timberwolves":"-105","Sacramento Kings":"-115"}
				d.Ack(false)
			}
		}(queueName, messages)
	}
	<-ctx.Done()
	log.Println("Receiver has been stopped")
}

var Manager games.GamesManager

var TeamAbbreviation = map[string]string{
	"atlanta hawks":          "ATL",
	"boston celtics":         "BOS",
	"brooklyn nets":          "BKN",
	"charlotte hornets":      "CHA",
	"chicago bulls":          "CHI",
	"cleveland cavaliers":    "CLE",
	"dallas mavericks":       "DAL",
	"denver nuggets":         "DEN",
	"detroit pistons":        "DET",
	"golden state warriors":  "GSW",
	"houston rockets":        "HOU",
	"indiana pacers":         "IND",
	"los angeles clippers":   "LAC",
	"la clippers":            "LAC",
	"la lakers":              "LAL",
	"los angeles lakers":     "LAL",
	"memphis grizzlies":      "MEM",
	"miami heat":             "MIA",
	"milwaukee bucks":        "MIL",
	"minnesota timberwolves": "MIN",
	"new orleans pelicans":   "NOP",
	"new york knicks":        "NYK",
	"oklahoma city thunder":  "OKC",
	"orlando magic":          "ORL",
	"philadelphia 76ers":     "PHI",
	"phoenix suns":           "PHX",
	"portland trail blazers": "POR",
	"sacramento kings":       "SAC",
	"san antonio spurs":      "SAS",
	"toronto raptors":        "TOR",
	"utah jazz":              "UTA",
	"washington wizards":     "WAS",
}

func storeData(ctx context.Context, queue string, data map[string]interface{}) error {
	switch queue {
	case "tickets":
		eventName := data["event_name"].(string)
		homeTeam, awayTeam, err := extractTeams(eventName)
		if err != nil {
			return fmt.Errorf("could not extract home team & away team out of tickets message: %v", err)
		}

		date, _ := time.Parse(time.RFC3339, data["start_date_time"].(string))
		formattedDate := date.Format("01.02.2006")
		startTime := data["start_date_time"].(string)
		lowestTicketPrice := fmt.Sprintf("$%.2f", data["min_ticket_price"].(float64))

		gameID := fmt.Sprintf("%s %s %s", TeamAbbreviation[homeTeam], TeamAbbreviation[awayTeam], formattedDate)
		gameKey := fmt.Sprintf("game:%s", gameID)

		exists, err := Manager.GameExists(ctx, gameKey)
		if err != nil {
			return fmt.Errorf("error checking game existence: %v", err)
		}
		if !exists {
			log.Printf("Game %s does not exist", gameID)
			return nil
		}
		log.Print(gameKey)
		fields := map[string]interface{}{
			"lowest_ticket_price": lowestTicketPrice,
		}
		err = Manager.CreateOrUpdateGame(ctx, gameKey, fields)
		if err != nil {
			return err
		}

		zsetKey := fmt.Sprintf("team:%s:upcoming_home_games", TeamAbbreviation[homeTeam])
		gameTime, _ := time.Parse(time.RFC3339, startTime)
		score := gameTime.Unix()
		err = Manager.AddUpcomingGame(ctx, zsetKey, gameID, score)
		if err != nil {
			return err
		}
	case "odds":
		// Extract and normalize team names
		homeTeamName := strings.ToLower(data["home_team"].(string))
		awayTeamName := strings.ToLower(data["away_team"].(string))

		// Map team names to abbreviations
		homeTeamAbbr, ok := TeamAbbreviation[homeTeamName]
		if !ok {
			log.Printf("Unknown home team: %s", homeTeamName)
			return fmt.Errorf("unknown home team")
		}

		awayTeamAbbr, ok := TeamAbbreviation[awayTeamName]
		if !ok {
			log.Printf("Unknown away team: %s", awayTeamName)
			return fmt.Errorf("unknown away team")
		}

		// Parse the date
		dateStr := data["start_time"].(string)
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.Printf("Error parsing date '%s': %v", dateStr, err)
			return fmt.Errorf("invalid date format")
		}
		formattedDate := date.Format("01.02.2006")

		// Extract odds
		odds := data["betting_prices"].(map[string]interface{})
		homeTeamOddsStr, ok := odds[data["home_team"].(string)].(string)
		if !ok {
			log.Printf("odds not found for home team: %s", data["home_team"].(string))
			return fmt.Errorf("odds not found for home team")
		}

		// Construct the game ID and key
		gameID := fmt.Sprintf("%s %s %s", homeTeamAbbr, awayTeamAbbr, formattedDate)
		gameKey := fmt.Sprintf("game:%s", gameID)

		// Check if the game exists
		exists, err := Manager.GameExists(ctx, gameKey)
		if err != nil {
			return fmt.Errorf("error checking game existence: %v", err)
		}
		if !exists {
			log.Printf("Game %s does not exist", gameID)
			return nil
		}

		// Update the game with odds
		fields := map[string]interface{}{
			"home_team_odds": homeTeamOddsStr,
		}

		err = Manager.CreateOrUpdateGame(ctx, gameKey, fields)
		if err != nil {
			log.Printf("Failed to update game: %v", err)
			return err
		}
		log.Printf("Odds updated for game %s", gameID)

	case "injuries":
		panic("unimplemented")

	default:
		return fmt.Errorf("unknown queue: %s", queue)
	}

	return nil
}

func extractTeams(eventName string) (homeTeam, awayTeam string, err error) {
	// Normalize the event name: convert to lowercase and remove special characters
	reg := regexp.MustCompile(`[^a-z0-9\s]+`) // Keep only alphanumeric and spaces
	normalizedEvent := reg.ReplaceAllString(strings.ToLower(eventName), "")

	// Split into words for matching
	words := strings.Fields(normalizedEvent)

	var matchedTeams []string

	// Look for known teams by checking consecutive word combinations
	for i := 0; i < len(words); i++ {
		for j := i + 1; j <= len(words); j++ {
			potentialTeam := strings.Join(words[i:j], " ") // Join words for potential match
			if _, exists := TeamAbbreviation[potentialTeam]; exists {
				matchedTeams = append(matchedTeams, potentialTeam)
				i = j - 1 // Skip ahead to avoid overlapping matches
				break
			}
		}
	}

	// Ensure exactly two teams were matched
	if len(matchedTeams) < 2 {
		return "", "", fmt.Errorf("could not extract exactly two teams from event: %s", eventName)
	}

	// Assign matched teams: first team is away, second is home
	return matchedTeams[0], matchedTeams[1], nil
}
