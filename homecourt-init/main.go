package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

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
	"LA Clippers":            "LAC",
	"LA Lakers":              "LAL",
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

func main() {
	err := godotenv.Load(".env.local")
	icsURL := os.Getenv("CALENDAR_SECRET")
	if err != nil {
		log.Fatalf("Error downloading .ics file: %v", err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", icsURL, nil)
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "HomecourtScheduler/1.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error downloading file: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("failed to download file: status code %d", resp.StatusCode)
	}

	outFile, err := os.Create("homecourt-schedule.ics")
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Fatalf("error saving file: %v", err)
	}
	// log.Printf("we did it :D")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	content, err := os.ReadFile("homecourt-schedule.ics")
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	reader := bytes.NewReader(content)
	calendar, err := ics.ParseCalendar(reader)
	if err != nil {
		log.Fatalf("failed to parse calendar: %v", err)
	}

	for _, event := range calendar.Events() {
		summary := event.GetProperty(ics.ComponentPropertySummary).Value
		startTime := event.GetProperty(ics.ComponentPropertyDtStart).Value
		location := event.GetProperty(ics.ComponentPropertyLocation).Value

		cleanSummary := strings.TrimSpace(strings.TrimPrefix(summary, "🏀"))
		teams := strings.Split(cleanSummary, "@")
		if len(teams) != 2 {
			log.Printf("unexpected format: %s", summary)
			continue
		}

		awayTeam := strings.TrimSpace(teams[0])
		homeTeam := strings.TrimSpace(teams[1])

		datetimeLayout := "20060102T150405Z"
		date, err := time.Parse(datetimeLayout, startTime)
		if err != nil {
			log.Printf("failed to parse date %s: %v", date, err)
			continue
		}
		formattedDate := date.Format("01.02.2006")
		time, err := time.Parse(datetimeLayout, startTime)
		formattedTime := time.Format("Jan 2, 2006")

		if err != nil {
			log.Printf("Failed to parse start time '%s': %v", startTime, err)
			continue
		}

		gameID := fmt.Sprintf("%s %s %s", TeamAbbreviation[homeTeam], TeamAbbreviation[awayTeam], formattedDate)
		gameKey := fmt.Sprintf("game:%s", gameID)
		fields := map[string]interface{}{
			"home_team":  TeamAbbreviation[homeTeam],
			"away_team":  TeamAbbreviation[awayTeam],
			"venueName":  location,
			"start_time": formattedTime,
		}
		err = redisClient.HSet(ctx, gameKey, fields).Err()
		if err != nil {
			log.Printf("failed to create or update game %s: %v", gameKey, err)
			continue
		}
		log.Printf("stored game: %s", gameKey)

		upcomingGamesKey := fmt.Sprintf("team:%s:upcoming_home_games", TeamAbbreviation[homeTeam])
		score := time.Unix()
		// this might potentially be gameid instead of gamekey....
		err = redisClient.ZAdd(ctx, upcomingGamesKey, redis.Z{
			Score:  float64(score),
			Member: gameID,
		}).Err()
		if err != nil {
			log.Printf("failed to add game to upcoming games list")
			continue
		}
		log.Printf("added game to upcoming games list")
	}

	// define types for what im storing in redis
	// parse and store into redis
	// can consider flushing redis each time we do this
	// end script
	log.Printf("finished loading games from schedule")
}
