package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Table creation queries
const (
	createTeamsTable = `
	CREATE TABLE IF NOT EXISTS teams (
		team_id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		city VARCHAR(100),
		league VARCHAR(50),
		logo_url VARCHAR(255)
	);`

	createPlayersTable = `
	CREATE TABLE IF NOT EXISTS players (
		player_id SERIAL PRIMARY KEY,
		team_id INTEGER NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		status VARCHAR(20) NOT NULL CHECK (status IN ('Active', 'Injured'))
	);`

	createGamesTable = `
	CREATE TABLE IF NOT EXISTS games (
		game_id SERIAL PRIMARY KEY,
		home_team_id INTEGER NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
		away_team_id INTEGER NOT NULL REFERENCES teams(team_id) ON DELETE CASCADE,
		scheduled_date TIMESTAMP NOT NULL,
		venue VARCHAR(100),
		CHECK (home_team_id <> away_team_id)
	);`

	createOddsTable = `
	CREATE TABLE IF NOT EXISTS odds (
		odds_id SERIAL PRIMARY KEY,
		game_id INTEGER NOT NULL REFERENCES games(game_id) ON DELETE CASCADE,
		home_team_odds DECIMAL(5,2) NOT NULL,
		away_team_odds DECIMAL(5,2) NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
)

func main() {
	// Load environment variables (if using .env files)
	// Uncomment the following lines if you're using a .env file

	err := godotenv.Load("../.env.local")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve database connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT") // e.g., "5432"
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE") // e.g., "disable", "require"

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" || dbSSLMode == "" {
		log.Fatal("One or more required environment variables are missing (DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, DB_SSLMODE)")
	}

	// Construct the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open a DB connection: %v", err)
	}
	defer db.Close()

	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Successfully connected to the database.")

	// Create a context
	ctx := context.Background()

	// Execute table creation in a transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// List of table creation queries in the correct order
	queries := []string{
		createTeamsTable,
		createPlayersTable,
		createGamesTable,
		createOddsTable,
	}

	for _, query := range queries {
		_, err := tx.ExecContext(ctx, query)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute query: %v\nError: %v", query, err)
		}
		log.Println("Executed query successfully.")
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("Database schema initialized successfully.")
}
