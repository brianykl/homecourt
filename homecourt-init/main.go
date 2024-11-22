package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

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
		endTime := event.GetProperty(ics.ComponentPropertyDtEnd).Value
		location := event.GetProperty(ics.ComponentPropertyLocation).Value
		startTimeParsed, err := time.Parse("20060102T150405Z", startTime)
		if err != nil {
			log.Printf("Failed to parse start time '%s': %v", startTime, err)
			continue
		}
	}
	// define types for what im storing in redis
	// parse and store into redis
	// can consider flushing redis each time we do this
	// end script
}
