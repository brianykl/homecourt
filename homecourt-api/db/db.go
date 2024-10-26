package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// --- Struct Definitions ---

// Team represents a sports team.
type Team struct {
	TeamID  int
	Name    string
	City    string
	League  string
	LogoURL string
}

// Player represents a player on a team.
type Player struct {
	PlayerID int
	TeamID   int
	Name     string
	Status   string // e.g., Active, Injured
}

// Game represents a game between two teams.
type Game struct {
	GameID             int             `json:"game_id"`
	HomeTeamID         int             `json:"home_team_id"`
	AwayTeamID         int             `json:"away_team_id"`
	ScheduledDate      time.Time       `json:"scheduled_date"`
	Venue              string          `json:"venue"`
	AverageTicketPrice decimal.Decimal `json:"average_ticket_price"`
	LowestTicketPrice  decimal.Decimal `json:"lowest_ticket_price"`
}

// Odds represents the betting odds for a game.
type Odds struct {
	OddsID       int
	GameID       int
	HomeTeamOdds decimal.Decimal
	AwayTeamOdds decimal.Decimal
	UpdatedAt    time.Time
}

// --- Interface Definitions ---

// DBManager defines the interface for database operations.
type DBManager interface {
	// Teams
	InsertTeam(ctx context.Context, team *Team) error
	GetTeam(ctx context.Context, teamID int) (*Team, error)
	UpdateTeam(ctx context.Context, team *Team) error
	DeleteTeam(ctx context.Context, teamID int) error

	// Players
	InsertPlayer(ctx context.Context, player *Player) error
	GetPlayer(ctx context.Context, playerID int) (*Player, error)
	UpdatePlayer(ctx context.Context, player *Player) error
	DeletePlayer(ctx context.Context, playerID int) error

	// Games
	InsertGame(ctx context.Context, game *Game) error
	GetGame(ctx context.Context, gameID int) (*Game, error)
	UpdateGame(ctx context.Context, game *Game) error
	DeleteGame(ctx context.Context, gameID int) error

	// Odds
	InsertOdds(ctx context.Context, odds *Odds) error
	GetOdds(ctx context.Context, oddsID int) (*Odds, error)
	UpdateOdds(ctx context.Context, odds *Odds) error
	DeleteOdds(ctx context.Context, oddsID int) error

	Close() error
}

// --- Implementation ---

type postgresDBManager struct {
	client *sql.DB
}

// NewDBManager initializes a new DBManager with a PostgreSQL connection.
func NewDBManager(connDetails string) (DBManager, error) {
	db, err := sql.Open("postgres", connDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %v", err)
	}

	return &postgresDBManager{client: db}, nil
}

// --- Teams CRUD Operations ---

// InsertTeam inserts a new team into the teams table.
func (ptm *postgresDBManager) InsertTeam(ctx context.Context, team *Team) error {
	query := `INSERT INTO teams (name, city, league, logo_url) VALUES ($1, $2, $3, $4) RETURNING team_id`

	err := ptm.client.QueryRowContext(ctx, query, team.Name, team.City, team.League, team.LogoURL).Scan(&team.TeamID)
	if err != nil {
		return fmt.Errorf("failed to insert team: %v", err)
	}

	return nil
}

// GetTeam retrieves a team by its ID.
func (ptm *postgresDBManager) GetTeam(ctx context.Context, teamID int) (*Team, error) {
	query := `SELECT team_id, name, city, league, logo_url FROM teams WHERE team_id = $1`

	row := ptm.client.QueryRowContext(ctx, query, teamID)

	team := &Team{}
	err := row.Scan(&team.TeamID, &team.Name, &team.City, &team.League, &team.LogoURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team with ID %d not found", teamID)
		}
		return nil, fmt.Errorf("failed to get team: %v", err)
	}

	return team, nil
}

// UpdateTeam updates an existing team’s details.
func (ptm *postgresDBManager) UpdateTeam(ctx context.Context, team *Team) error {
	query := `UPDATE teams SET name = $1, city = $2, league = $3, logo_url = $4 WHERE team_id = $5`

	res, err := ptm.client.ExecContext(ctx, query, team.Name, team.City, team.League, team.LogoURL, team.TeamID)
	if err != nil {
		return fmt.Errorf("failed to update team: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for team update: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no team found with ID %d to update", team.TeamID)
	}

	return nil
}

// DeleteTeam removes a team from the teams table.
func (ptm *postgresDBManager) DeleteTeam(ctx context.Context, teamID int) error {
	query := `DELETE FROM teams WHERE team_id = $1`

	res, err := ptm.client.ExecContext(ctx, query, teamID)
	if err != nil {
		return fmt.Errorf("failed to delete team: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for team deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no team found with ID %d to delete", teamID)
	}

	return nil
}

// --- Players CRUD Operations ---

// InsertPlayer inserts a new player into the players table.
func (ptm *postgresDBManager) InsertPlayer(ctx context.Context, player *Player) error {
	query := `INSERT INTO players (team_id, name, status) VALUES ($1, $2, $3) RETURNING player_id`

	err := ptm.client.QueryRowContext(ctx, query, player.TeamID, player.Name, player.Status).Scan(&player.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to insert player: %v", err)
	}

	return nil
}

// GetPlayer retrieves a player by their ID.
func (ptm *postgresDBManager) GetPlayer(ctx context.Context, playerID int) (*Player, error) {
	query := `SELECT player_id, team_id, name, status FROM players WHERE player_id = $1`

	row := ptm.client.QueryRowContext(ctx, query, playerID)

	player := &Player{}
	err := row.Scan(&player.PlayerID, &player.TeamID, &player.Name, &player.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player with ID %d not found", playerID)
		}
		return nil, fmt.Errorf("failed to get player: %v", err)
	}

	return player, nil
}

// UpdatePlayer updates an existing player's details.
func (ptm *postgresDBManager) UpdatePlayer(ctx context.Context, player *Player) error {
	query := `UPDATE players SET team_id = $1, name = $2, status = $3 WHERE player_id = $4`

	res, err := ptm.client.ExecContext(ctx, query, player.TeamID, player.Name, player.Status, player.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to update player: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for player update: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no player found with ID %d to update", player.PlayerID)
	}

	return nil
}

// DeletePlayer removes a player from the players table.
func (ptm *postgresDBManager) DeletePlayer(ctx context.Context, playerID int) error {
	query := `DELETE FROM players WHERE player_id = $1`

	res, err := ptm.client.ExecContext(ctx, query, playerID)
	if err != nil {
		return fmt.Errorf("failed to delete player: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for player deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no player found with ID %d to delete", playerID)
	}

	return nil
}

// --- Games CRUD Operations ---

// InsertGame inserts a new game into the games table.
func (ptm *postgresDBManager) InsertGame(ctx context.Context, game *Game) error {
	query := `
        INSERT INTO games (home_team_id, away_team_id, scheduled_date, venue, average_ticket_price, lowest_ticket_price)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING game_id
    `

	// Convert decimal.Decimal to string for insertion
	err := ptm.client.QueryRowContext(ctx, query,
		game.HomeTeamID,
		game.AwayTeamID,
		game.ScheduledDate.Truncate(time.Microsecond), // Truncate to match PostgreSQL precision
		game.Venue,
		game.AverageTicketPrice.String(),
		game.LowestTicketPrice.String(),
	).Scan(&game.GameID)
	if err != nil {
		return fmt.Errorf("failed to insert game: %v", err)
	}

	return nil
}

// GetGame retrieves a game by its ID.
func (ptm *postgresDBManager) GetGame(ctx context.Context, gameID int) (*Game, error) {
	query := `
        SELECT game_id, home_team_id, away_team_id, scheduled_date, venue, average_ticket_price, lowest_ticket_price
        FROM games
        WHERE game_id = $1
    `

	row := ptm.client.QueryRowContext(ctx, query, gameID)

	game := &Game{}
	var avgPriceStr, lowPriceStr string
	err := row.Scan(
		&game.GameID,
		&game.HomeTeamID,
		&game.AwayTeamID,
		&game.ScheduledDate,
		&game.Venue,
		&avgPriceStr,
		&lowPriceStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game with ID %d not found", gameID)
		}
		return nil, fmt.Errorf("failed to get game: %v", err)
	}

	// Parse the decimal strings into decimal.Decimal
	game.AverageTicketPrice, err = decimal.NewFromString(avgPriceStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse average_ticket_price: %v", err)
	}

	game.LowestTicketPrice, err = decimal.NewFromString(lowPriceStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lowest_ticket_price: %v", err)
	}

	return game, nil
}

// UpdateGame updates an existing game's details.
func (ptm *postgresDBManager) UpdateGame(ctx context.Context, game *Game) error {
	query := `
        UPDATE games
        SET home_team_id = $1,
            away_team_id = $2,
            scheduled_date = $3,
            venue = $4,
            average_ticket_price = $5,
            lowest_ticket_price = $6
        WHERE game_id = $7
    `

	// Convert decimal.Decimal to string for updating
	res, err := ptm.client.ExecContext(ctx, query,
		game.HomeTeamID,
		game.AwayTeamID,
		game.ScheduledDate.Truncate(time.Microsecond), // Truncate to match PostgreSQL precision
		game.Venue,
		game.AverageTicketPrice.String(),
		game.LowestTicketPrice.String(),
		game.GameID,
	)
	if err != nil {
		return fmt.Errorf("failed to update game: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for game update: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no game found with ID %d to update", game.GameID)
	}

	return nil
}

// DeleteGame removes a game from the games table.
func (ptm *postgresDBManager) DeleteGame(ctx context.Context, gameID int) error {
	query := `DELETE FROM games WHERE game_id = $1`

	res, err := ptm.client.ExecContext(ctx, query, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete game: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for game deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no game found with ID %d to delete", gameID)
	}

	return nil
}

// --- Odds CRUD Operations ---

// InsertOdds inserts new odds into the odds table.
func (ptm *postgresDBManager) InsertOdds(ctx context.Context, odds *Odds) error {
	query := `INSERT INTO odds (game_id, home_team_odds, away_team_odds, updated_at) VALUES ($1, $2, $3, $4) RETURNING odds_id`

	err := ptm.client.QueryRowContext(ctx, query, odds.GameID, odds.HomeTeamOdds.String(), odds.AwayTeamOdds.String(), odds.UpdatedAt).Scan(&odds.OddsID)
	if err != nil {
		return fmt.Errorf("failed to insert odds: %v", err)
	}

	return nil
}

// GetOdds retrieves odds by their ID.
func (ptm *postgresDBManager) GetOdds(ctx context.Context, oddsID int) (*Odds, error) {
	query := `SELECT odds_id, game_id, home_team_odds, away_team_odds, updated_at FROM odds WHERE odds_id = $1`

	row := ptm.client.QueryRowContext(ctx, query, oddsID)

	odds := &Odds{}
	var homeOddsStr, awayOddsStr string
	err := row.Scan(&odds.OddsID, &odds.GameID, &homeOddsStr, &awayOddsStr, &odds.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("odds with ID %d not found", oddsID)
		}
		return nil, fmt.Errorf("failed to get odds: %v", err)
	}

	odds.HomeTeamOdds, err = decimal.NewFromString(homeOddsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse home_team_odds: %v", err)
	}

	odds.AwayTeamOdds, err = decimal.NewFromString(awayOddsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse away_team_odds: %v", err)
	}

	return odds, nil
}

// UpdateOdds updates existing odds in the odds table.
func (ptm *postgresDBManager) UpdateOdds(ctx context.Context, odds *Odds) error {
	query := `UPDATE odds SET home_team_odds = $1, away_team_odds = $2, updated_at = $3 WHERE odds_id = $4`

	res, err := ptm.client.ExecContext(ctx, query, odds.HomeTeamOdds.String(), odds.AwayTeamOdds.String(), odds.UpdatedAt, odds.OddsID)
	if err != nil {
		return fmt.Errorf("failed to update odds: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for odds update: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no odds found with ID %d to update", odds.OddsID)
	}

	return nil
}

// DeleteOdds removes odds from the odds table.
func (ptm *postgresDBManager) DeleteOdds(ctx context.Context, oddsID int) error {
	query := `DELETE FROM odds WHERE odds_id = $1`

	res, err := ptm.client.ExecContext(ctx, query, oddsID)
	if err != nil {
		return fmt.Errorf("failed to delete odds: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected for odds deletion: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no odds found with ID %d to delete", oddsID)
	}

	return nil
}

// --- Utility Functions ---

// Close closes the database connection.
func (ptm *postgresDBManager) Close() error {
	return ptm.client.Close()
}

// --- Example Usage ---

/*
func main() {
	connDetails := "user=youruser password=yourpassword dbname=yourdb sslmode=disable"
	dbManager, err := NewDBManager(connDetails)
	if err != nil {
		log.Fatalf("Failed to initialize DB Manager: %v", err)
	}
	defer dbManager.Close()

	ctx := context.Background()

	// Insert a new team
	newTeam := &Team{
		Name:    "Los Angeles Lakers",
		City:    "Los Angeles",
		League:  "NBA",
		LogoURL: "https://example.com/lakers.png",
	}
	err = dbManager.InsertTeam(ctx, newTeam)
	if err != nil {
		log.Fatalf("InsertTeam failed: %v", err)
	}
	fmt.Printf("Inserted Team with ID: %d\n", newTeam.TeamID)

	// Get the team
	team, err := dbManager.GetTeam(ctx, newTeam.TeamID)
	if err != nil {
		log.Fatalf("GetTeam failed: %v", err)
	}
	fmt.Printf("Retrieved Team: %+v\n", team)

	// Update the team
	team.City = "LA"
	err = dbManager.UpdateTeam(ctx, team)
	if err != nil {
		log.Fatalf("UpdateTeam failed: %v", err)
	}
	fmt.Println("Team updated successfully.")

	// Delete the team
	err = dbManager.DeleteTeam(ctx, team.TeamID)
	if err != nil {
		log.Fatalf("DeleteTeam failed: %v", err)
	}
	fmt.Println("Team deleted successfully.")
}
*/
