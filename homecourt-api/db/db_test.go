package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

// --- Helper Functions ---

// loadEnv loads environment variables from the specified file.
// If a *testing.T is provided, it uses it to report fatal errors.
func loadEnv(t *testing.T, filePath string) {
	err := godotenv.Load(filePath)
	if err != nil {
		if t != nil {
			t.Fatalf("Error loading %s file: %v", filePath, err)
		} else {
			log.Fatalf("Error loading %s file: %v", filePath, err)
		}
	}
}

// getConnectionString constructs the PostgreSQL connection string from environment variables.
func getConnectionString(t *testing.T) string {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" || dbSSLMode == "" {
		t.Fatal("One or more required environment variables are missing (DB_HOST, DB_USER, DB_PASS, DB_NAME, DB_SSLMODE)")
	}

	// Log connection details (Be cautious with logging sensitive information in production)
	t.Logf("DB_HOST: %s", dbHost)
	t.Logf("DB_USER: %s", dbUser)
	t.Logf("DB_NAME: %s", dbName)
	t.Logf("DB_SSLMODE: %s", dbSSLMode)

	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbName, dbSSLMode)

	return connString
}

// --- Test Setup ---

var dbManager DBManager

func TestMain(m *testing.M) {
	// Initialize a dummy *testing.T for loading environment variables
	dummyT := &testing.T{}

	// Load environment variables from .env.local file
	loadEnv(dummyT, "../.env.local")

	// Get connection string
	connString := getConnectionString(dummyT)

	// Initialize DBManager
	var err error
	dbManager, err = NewDBManager(connString)
	if err != nil {
		log.Fatalf("Failed to initialize DB Manager: %v", err)
	}
	defer dbManager.Close()

	// Run tests
	exitVal := m.Run()

	// Exit with the appropriate code
	os.Exit(exitVal)
}

// --- Teams Tests ---

func TestTeamsCRUD(t *testing.T) {
	ctx := context.Background()

	// Create a new team
	newTeam := &Team{
		Name:    "Test Team",
		City:    "Test City",
		League:  "Test League",
		LogoURL: "https://example.com/logo.png",
	}

	// Insert Team
	err := dbManager.InsertTeam(ctx, newTeam)
	if err != nil {
		t.Fatalf("InsertTeam failed: %v", err)
	}
	t.Logf("Inserted Team with ID: %d", newTeam.TeamID)

	// Get Team
	retrievedTeam, err := dbManager.GetTeam(ctx, newTeam.TeamID)
	if err != nil {
		t.Fatalf("GetTeam failed: %v", err)
	}
	t.Logf("Retrieved Team: %+v", retrievedTeam)

	// Verify Inserted Data
	if retrievedTeam.Name != newTeam.Name {
		t.Errorf("Team Name mismatch: got %s, want %s", retrievedTeam.Name, newTeam.Name)
	}
	if retrievedTeam.City != newTeam.City {
		t.Errorf("Team City mismatch: got %s, want %s", retrievedTeam.City, newTeam.City)
	}
	if retrievedTeam.League != newTeam.League {
		t.Errorf("Team League mismatch: got %s, want %s", retrievedTeam.League, newTeam.League)
	}
	if retrievedTeam.LogoURL != newTeam.LogoURL {
		t.Errorf("Team LogoURL mismatch: got %s, want %s", retrievedTeam.LogoURL, newTeam.LogoURL)
	}

	// Update Team
	retrievedTeam.City = "Updated City"
	retrievedTeam.League = "Updated League"
	err = dbManager.UpdateTeam(ctx, retrievedTeam)
	if err != nil {
		t.Fatalf("UpdateTeam failed: %v", err)
	}
	t.Log("Team updated successfully.")

	// Get Updated Team
	updatedTeam, err := dbManager.GetTeam(ctx, newTeam.TeamID)
	if err != nil {
		t.Fatalf("GetTeam after update failed: %v", err)
	}
	t.Logf("Updated Team: %+v", updatedTeam)

	// Verify Updated Data
	if updatedTeam.City != "Updated City" {
		t.Errorf("Team City not updated: got %s, want %s", updatedTeam.City, "Updated City")
	}
	if updatedTeam.League != "Updated League" {
		t.Errorf("Team League not updated: got %s, want %s", updatedTeam.League, "Updated League")
	}

	// Delete Team
	err = dbManager.DeleteTeam(ctx, newTeam.TeamID)
	if err != nil {
		t.Fatalf("DeleteTeam failed: %v", err)
	}
	t.Log("Team deleted successfully.")

	// Verify Deletion
	_, err = dbManager.GetTeam(ctx, newTeam.TeamID)
	if err == nil {
		t.Errorf("Expected error when getting deleted team, but got none")
	} else {
		t.Logf("Confirmed team deletion: %v", err)
	}
}

// --- Players Tests ---

func TestPlayersCRUD(t *testing.T) {
	ctx := context.Background()

	// First, create a team to associate with the player
	team := &Team{
		Name:    "Player Test Team",
		City:    "Player Test City",
		League:  "Player Test League",
		LogoURL: "https://example.com/player_logo.png",
	}

	err := dbManager.InsertTeam(ctx, team)
	if err != nil {
		t.Fatalf("InsertTeam for player test failed: %v", err)
	}
	t.Logf("Inserted Team for Player with ID: %d", team.TeamID)

	// Create a new player
	newPlayer := &Player{
		TeamID: team.TeamID,
		Name:   "Test Player",
		Status: "Active",
	}

	// Insert Player
	err = dbManager.InsertPlayer(ctx, newPlayer)
	if err != nil {
		t.Fatalf("InsertPlayer failed: %v", err)
	}
	t.Logf("Inserted Player with ID: %d", newPlayer.PlayerID)

	// Get Player
	retrievedPlayer, err := dbManager.GetPlayer(ctx, newPlayer.PlayerID)
	if err != nil {
		t.Fatalf("GetPlayer failed: %v", err)
	}
	t.Logf("Retrieved Player: %+v", retrievedPlayer)

	// Verify Inserted Data
	if retrievedPlayer.Name != newPlayer.Name {
		t.Errorf("Player Name mismatch: got %s, want %s", retrievedPlayer.Name, newPlayer.Name)
	}
	if retrievedPlayer.Status != newPlayer.Status {
		t.Errorf("Player Status mismatch: got %s, want %s", retrievedPlayer.Status, newPlayer.Status)
	}
	if retrievedPlayer.TeamID != team.TeamID {
		t.Errorf("Player TeamID mismatch: got %d, want %d", retrievedPlayer.TeamID, team.TeamID)
	}

	// Update Player
	retrievedPlayer.Status = "Injured"
	err = dbManager.UpdatePlayer(ctx, retrievedPlayer)
	if err != nil {
		t.Fatalf("UpdatePlayer failed: %v", err)
	}
	t.Log("Player updated successfully.")

	// Get Updated Player
	updatedPlayer, err := dbManager.GetPlayer(ctx, newPlayer.PlayerID)
	if err != nil {
		t.Fatalf("GetPlayer after update failed: %v", err)
	}
	t.Logf("Updated Player: %+v", updatedPlayer)

	// Verify Updated Data
	if updatedPlayer.Status != "Injured" {
		t.Errorf("Player Status not updated: got %s, want %s", updatedPlayer.Status, "Injured")
	}

	// Delete Player
	err = dbManager.DeletePlayer(ctx, newPlayer.PlayerID)
	if err != nil {
		t.Fatalf("DeletePlayer failed: %v", err)
	}
	t.Log("Player deleted successfully.")

	// Verify Deletion
	_, err = dbManager.GetPlayer(ctx, newPlayer.PlayerID)
	if err == nil {
		t.Errorf("Expected error when getting deleted player, but got none")
	} else {
		t.Logf("Confirmed player deletion: %v", err)
	}

	// Clean up: Delete the team
	err = dbManager.DeleteTeam(ctx, team.TeamID)
	if err != nil {
		t.Fatalf("DeleteTeam for player test failed: %v", err)
	}
	t.Log("Player Test Team deleted successfully.")
}

// --- Games Tests ---

func TestGamesCRUD(t *testing.T) {
	ctx := context.Background()

	// Create two teams to associate with the game
	homeTeam := &Team{
		Name:    "Home Test Team",
		City:    "Home City",
		League:  "Home League",
		LogoURL: "https://example.com/home_logo.png",
	}

	err := dbManager.InsertTeam(ctx, homeTeam)
	if err != nil {
		t.Fatalf("InsertTeam for home team failed: %v", err)
	}
	t.Logf("Inserted Home Team with ID: %d", homeTeam.TeamID)

	awayTeam := &Team{
		Name:    "Away Test Team",
		City:    "Away City",
		League:  "Away League",
		LogoURL: "https://example.com/away_logo.png",
	}

	err = dbManager.InsertTeam(ctx, awayTeam)
	if err != nil {
		t.Fatalf("InsertTeam for away team failed: %v", err)
	}
	t.Logf("Inserted Away Team with ID: %d", awayTeam.TeamID)

	// Create a new game with average and lowest ticket prices
	newGame := &Game{
		HomeTeamID:         homeTeam.TeamID,
		AwayTeamID:         awayTeam.TeamID,
		ScheduledDate:      time.Now().Add(24 * time.Hour).UTC().Truncate(time.Microsecond), // Truncate to match PostgreSQL precision
		Venue:              "Test Arena",
		AverageTicketPrice: decimal.NewFromFloat(50.00),
		LowestTicketPrice:  decimal.NewFromFloat(25.00),
	}

	// Insert Game
	err = dbManager.InsertGame(ctx, newGame)
	if err != nil {
		t.Fatalf("InsertGame failed: %v", err)
	}
	t.Logf("Inserted Game with ID: %d", newGame.GameID)

	// Retrieve Game
	retrievedGame, err := dbManager.GetGame(ctx, newGame.GameID)
	if err != nil {
		t.Fatalf("GetGame failed: %v", err)
	}
	t.Logf("Retrieved Game: %+v", retrievedGame)

	// Verify Inserted Data
	if retrievedGame.HomeTeamID != newGame.HomeTeamID {
		t.Errorf("Game HomeTeamID mismatch: got %d, want %d", retrievedGame.HomeTeamID, newGame.HomeTeamID)
	}
	if retrievedGame.AwayTeamID != newGame.AwayTeamID {
		t.Errorf("Game AwayTeamID mismatch: got %d, want %d", retrievedGame.AwayTeamID, newGame.AwayTeamID)
	}

	// Compare ScheduledDate with truncated time to match precision
	expectedTime := newGame.ScheduledDate
	actualTime := retrievedGame.ScheduledDate.Truncate(time.Microsecond)
	if !expectedTime.Equal(actualTime) {
		t.Errorf("Game ScheduledDate mismatch: got %v, want %v", actualTime, expectedTime)
	}

	if retrievedGame.Venue != newGame.Venue {
		t.Errorf("Game Venue mismatch: got %s, want %s", retrievedGame.Venue, newGame.Venue)
	}

	// Verify AverageTicketPrice
	if !retrievedGame.AverageTicketPrice.Equal(newGame.AverageTicketPrice) {
		t.Errorf("Game AverageTicketPrice mismatch: got %s, want %s", retrievedGame.AverageTicketPrice, newGame.AverageTicketPrice)
	}

	// Verify LowestTicketPrice
	if !retrievedGame.LowestTicketPrice.Equal(newGame.LowestTicketPrice) {
		t.Errorf("Game LowestTicketPrice mismatch: got %s, want %s", retrievedGame.LowestTicketPrice, newGame.LowestTicketPrice)
	}

	// Update Game with new venue and updated ticket prices
	retrievedGame.Venue = "Updated Arena"
	retrievedGame.AverageTicketPrice = decimal.NewFromFloat(55.00)
	retrievedGame.LowestTicketPrice = decimal.NewFromFloat(20.00)
	retrievedGame.ScheduledDate = retrievedGame.ScheduledDate.Add(1 * time.Hour) // Optional: update date

	err = dbManager.UpdateGame(ctx, retrievedGame)
	if err != nil {
		t.Fatalf("UpdateGame failed: %v", err)
	}
	t.Log("Game updated successfully.")

	// Retrieve Updated Game
	updatedGame, err := dbManager.GetGame(ctx, newGame.GameID)
	if err != nil {
		t.Fatalf("GetGame after update failed: %v", err)
	}
	t.Logf("Updated Game: %+v", updatedGame)

	// Verify Updated Data
	if updatedGame.Venue != "Updated Arena" {
		t.Errorf("Game Venue not updated: got %s, want %s", updatedGame.Venue, "Updated Arena")
	}

	// Verify Updated AverageTicketPrice
	expectedAvgPrice := decimal.NewFromFloat(55.00)
	if !updatedGame.AverageTicketPrice.Equal(expectedAvgPrice) {
		t.Errorf("Game AverageTicketPrice not updated: got %s, want %s", updatedGame.AverageTicketPrice, expectedAvgPrice)
	}

	// Verify Updated LowestTicketPrice
	expectedLowPrice := decimal.NewFromFloat(20.00)
	if !updatedGame.LowestTicketPrice.Equal(expectedLowPrice) {
		t.Errorf("Game LowestTicketPrice not updated: got %s, want %s", updatedGame.LowestTicketPrice, expectedLowPrice)
	}

	// (Optional) Verify Updated ScheduledDate if modified
	// expectedDate := retrievedGame.ScheduledDate
	// actualDate := updatedGame.ScheduledDate.Truncate(time.Microsecond)
	// if !expectedDate.Equal(actualDate) {
	//     t.Errorf("Game ScheduledDate mismatch after update: got %v, want %v", actualDate, expectedDate)
	// }

	// Delete Game
	err = dbManager.DeleteGame(ctx, newGame.GameID)
	if err != nil {
		t.Fatalf("DeleteGame failed: %v", err)
	}
	t.Log("Game deleted successfully.")

	// Verify Deletion
	_, err = dbManager.GetGame(ctx, newGame.GameID)
	if err == nil {
		t.Errorf("Expected error when getting deleted game, but got none")
	} else {
		t.Logf("Confirmed game deletion: %v", err)
	}

	// Clean up: Delete the teams
	err = dbManager.DeleteTeam(ctx, homeTeam.TeamID)
	if err != nil {
		t.Fatalf("DeleteTeam for home team failed: %v", err)
	}
	t.Log("Home Test Team deleted successfully.")

	err = dbManager.DeleteTeam(ctx, awayTeam.TeamID)
	if err != nil {
		t.Fatalf("DeleteTeam for away team failed: %v", err)
	}
	t.Log("Away Test Team deleted successfully.")
}

// --- Odds Tests ---

func TestOddsCRUD(t *testing.T) {
	ctx := context.Background()

	// Create two teams to associate with the game
	homeTeam := &Team{
		Name:    "Odds Home Team",
		City:    "Odds Home City",
		League:  "Odds League",
		LogoURL: "https://example.com/odds_home_logo.png",
	}

	err := dbManager.InsertTeam(ctx, homeTeam)
	if err != nil {
		t.Fatalf("InsertTeam for home team failed: %v", err)
	}
	t.Logf("Inserted Home Team with ID: %d", homeTeam.TeamID)

	awayTeam := &Team{
		Name:    "Odds Away Team",
		City:    "Odds Away City",
		League:  "Odds League",
		LogoURL: "https://example.com/odds_away_logo.png",
	}

	err = dbManager.InsertTeam(ctx, awayTeam)
	if err != nil {
		t.Fatalf("InsertTeam for away team failed: %v", err)
	}
	t.Logf("Inserted Away Team with ID: %d", awayTeam.TeamID)

	// Create a new game
	newGame := &Game{
		HomeTeamID:         homeTeam.TeamID,
		AwayTeamID:         awayTeam.TeamID,
		ScheduledDate:      time.Now().Add(48 * time.Hour).UTC().Truncate(time.Microsecond),
		Venue:              "Odds Arena",
		AverageTicketPrice: decimal.NewFromFloat(60.00),
		LowestTicketPrice:  decimal.NewFromFloat(30.00),
	}

	err = dbManager.InsertGame(ctx, newGame)
	if err != nil {
		t.Fatalf("InsertGame failed: %v", err)
	}
	t.Logf("Inserted Game with ID: %d", newGame.GameID)

	// Create new odds
	newOdds := &Odds{
		GameID:       newGame.GameID,
		HomeTeamOdds: decimal.NewFromFloat(1.95),
		AwayTeamOdds: decimal.NewFromFloat(1.85),
		UpdatedAt:    time.Now().UTC(),
	}

	// Insert Odds
	err = dbManager.InsertOdds(ctx, newOdds)
	if err != nil {
		t.Fatalf("InsertOdds failed: %v", err)
	}
	t.Logf("Inserted Odds with ID: %d", newOdds.OddsID)

	// Get Odds
	retrievedOdds, err := dbManager.GetOdds(ctx, newOdds.OddsID)
	if err != nil {
		t.Fatalf("GetOdds failed: %v", err)
	}
	t.Logf("Retrieved Odds: %+v", retrievedOdds)

	// Verify Inserted Data
	if !retrievedOdds.HomeTeamOdds.Equal(newOdds.HomeTeamOdds) {
		t.Errorf("Odds HomeTeamOdds mismatch: got %s, want %s", retrievedOdds.HomeTeamOdds, newOdds.HomeTeamOdds)
	}
	if !retrievedOdds.AwayTeamOdds.Equal(newOdds.AwayTeamOdds) {
		t.Errorf("Odds AwayTeamOdds mismatch: got %s, want %s", retrievedOdds.AwayTeamOdds, newOdds.AwayTeamOdds)
	}

	// Update Odds
	retrievedOdds.HomeTeamOdds = decimal.NewFromFloat(2.05)
	retrievedOdds.AwayTeamOdds = decimal.NewFromFloat(1.75)
	retrievedOdds.UpdatedAt = time.Now().UTC()
	err = dbManager.UpdateOdds(ctx, retrievedOdds)
	if err != nil {
		t.Fatalf("UpdateOdds failed: %v", err)
	}
	t.Log("Odds updated successfully.")

	// Get Updated Odds
	updatedOdds, err := dbManager.GetOdds(ctx, newOdds.OddsID)
	if err != nil {
		t.Fatalf("GetOdds after update failed: %v", err)
	}
	t.Logf("Updated Odds: %+v", updatedOdds)

	// Verify Updated Data
	expectedHomeOdds := decimal.NewFromFloat(2.05)
	if !updatedOdds.HomeTeamOdds.Equal(expectedHomeOdds) {
		t.Errorf("Odds HomeTeamOdds not updated: got %s, want %s", updatedOdds.HomeTeamOdds, expectedHomeOdds)
	}

	expectedAwayOdds := decimal.NewFromFloat(1.75)
	if !updatedOdds.AwayTeamOdds.Equal(expectedAwayOdds) {
		t.Errorf("Odds AwayTeamOdds not updated: got %s, want %s", updatedOdds.AwayTeamOdds, expectedAwayOdds)
	}

	// Delete Odds
	err = dbManager.DeleteOdds(ctx, newOdds.OddsID)
	if err != nil {
		t.Fatalf("DeleteOdds failed: %v", err)
	}
	t.Log("Odds deleted successfully.")

	// Verify Deletion
	_, err = dbManager.GetOdds(ctx, newOdds.OddsID)
	if err == nil {
		t.Errorf("Expected error when getting deleted odds, but got none")
	} else {
		t.Logf("Confirmed odds deletion: %v", err)
	}

	// Clean up: Delete the game and teams
	err = dbManager.DeleteGame(ctx, newGame.GameID)
	if err != nil {
		t.Fatalf("DeleteGame failed: %v", err)
	}
	t.Log("Game deleted successfully.")

	err = dbManager.DeleteTeam(ctx, homeTeam.TeamID)
	if err != nil {
		t.Fatalf("DeleteTeam for home team failed: %v", err)
	}
	t.Log("Home Test Team deleted successfully.")

	err = dbManager.DeleteTeam(ctx, awayTeam.TeamID)
	if err != nil {
		t.Fatalf("DeleteTeam for away team failed: %v", err)
	}
	t.Log("Away Test Team deleted successfully.")
}
