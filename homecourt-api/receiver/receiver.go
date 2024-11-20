package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"homecourt-api/games"
	"log"
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
	"Atlanta Hawks":          "ATL",
	"Boston Celtics":         "BOS",
	"Brooklyn Nets":          "BKN",
	"Charlotte Hornets":      "CHA",
	"Chicago Bulls":          "CHI",
	"Cleveland Cavaliers":    "CLE",
	"Dallas Mavericks":       "DAL",
	"Denver Nuggets":         "DEN",
	"Detroit Pistons":        "DET",
	"Golden State Warriors":  "GSW",
	"Houston Rockets":        "HOU",
	"Indiana Pacers":         "IND",
	"Los Angeles Clippers":   "LAC",
	"Los Angeles Lakers":     "LAL",
	"Memphis Grizzlies":      "MEM",
	"Miami Heat":             "MIA",
	"Milwaukee Bucks":        "MIL",
	"Minnesota Timberwolves": "MIN",
	"New Orleans Pelicans":   "NOP",
	"New York Knicks":        "NYK",
	"Oklahoma City Thunder":  "OKC",
	"Orlando Magic":          "ORL",
	"Philadelphia 76ers":     "PHI",
	"Phoenix Suns":           "PHX",
	"Portland Trail Blazers": "POR",
	"Sacramento Kings":       "SAC",
	"San Antonio Spurs":      "SAS",
	"Toronto Raptors":        "TOR",
	"Utah Jazz":              "UTA",
	"Washington Wizards":     "WAS",
}

func storeData(ctx context.Context, queue string, data map[string]interface{}) error {
	switch queue {
	case "tickets":
		eventName := data["event_name"].(string)
		teams := strings.Split(eventName, " vs ")
		if len(teams) != 2 {
			return fmt.Errorf("unexpected event name format")
		}
		homeTeam := TeamAbbreviation[teams[0]]
		awayTeam := TeamAbbreviation[teams[1]]
		venueName := data["venue_name"].(string)
		date, _ := time.Parse(time.RFC3339, data["start_date_time"].(string))
		formattedDate := date.Format("01.02.2006")
		startTime := data["start_date_time"].(string)
		lowestTicketPrice := fmt.Sprintf("$%.2f", data["min_ticket_price"].(float64))

		gameID := fmt.Sprintf("%s_%s_%s", homeTeam, awayTeam, formattedDate)
		gameKey := fmt.Sprintf("game:%s", gameID)

		fields := map[string]interface{}{
			"home_team":           homeTeam,
			"away_team":           awayTeam,
			"venueName":           venueName,
			"start_time":          startTime,
			"lowest_ticket_price": lowestTicketPrice,
		}
		err := Manager.CreateOrUpdateGame(ctx, gameKey, fields)
		if err != nil {
			return err
		}

		zsetKey := fmt.Sprintf("team:%s:upcoming_home_games", homeTeam)
		gameTime, _ := time.Parse(time.RFC3339, startTime)
		score := gameTime.Unix()
		err = Manager.AddUpcomingGame(ctx, zsetKey, gameID, score)
		if err != nil {
			return err
		}
	case "odds":
		panic("unimplemented")
	case "injuries":
		panic("unimplemented")
	default:
		return fmt.Errorf("unknown queue: %s", queue)
	}
	return nil
}
