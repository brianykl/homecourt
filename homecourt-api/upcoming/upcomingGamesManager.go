package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Game represents the structure of a game stored in Redis.
type Game struct {
	GameID            string `json:"game_id"`
	HomeTeamOdds      string `json:"home_team_odds"`
	Venue             string `json:"venue"`
	Time              string `json:"time"`
	LowestTicketPrice string `json:"lowest_ticket_price"`
	InjuredPlayers    []struct {
		Team       string `json:"team"`
		PlayerName string `json:"player_name"`
		Status     string `json:"status"`
	} `json:"injured_players"`
}

// UpcomingGamesManager defines the methods for managing upcoming games and game details.
type UpcomingGamesManager interface {
	// Upcoming games per team (ZSET of game IDs)
	AddUpcomingGame(ctx context.Context, teamID string, gameID string, startTime time.Time) error
	GetUpcomingGames(ctx context.Context, teamID string, count int) ([]string, error)
	DeleteUpcomingGames(ctx context.Context, teamID string) error

	// Individual game details
	StoreGame(ctx context.Context, gameID string, game Game) error
	GetGame(ctx context.Context, gameID string) (Game, error)
	UpdateGame(ctx context.Context, gameID string, fields map[string]interface{}) error
	DeleteGame(ctx context.Context, gameID string) error
}

// redisUpcomingGamesManager manages the Redis connection and operations.
type redisUpcomingGamesManager struct {
	client *redis.Client
}

// AddUpcomingGame implements UpcomingGamesManager.
func (r *redisUpcomingGamesManager) AddUpcomingGame(ctx context.Context, teamID string, gameID string, startTime time.Time) error {
	zsetKey := fmt.Sprintf("team:%s:upcoming_home_games", teamID)
	score := float64(startTime.Unix())

	err := r.client.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: gameID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add game to ZSET: %v", err)
	}

	return nil
}

// GetUpcomingGames implements UpcomingGamesManager.
func (r *redisUpcomingGamesManager) GetUpcomingGames(ctx context.Context, teamID string, count int) ([]string, error) {
	zsetKey := fmt.Sprintf("team:%s:upcoming_home_games", teamID)
	now := time.Now().Unix()

	gameIDs, err := r.client.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%d", now),
		Max:    "+inf",
		Offset: 0,
		Count:  int64(count),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming games: %v", err)
	}

	return gameIDs, nil
}

// DeleteUpcomingGames implements UpcomingGamesManager.
// may want to also delete the actual game data for each deleted game down the line as well
func (r *redisUpcomingGamesManager) DeleteUpcomingGames(ctx context.Context, teamID string) error {
	zsetKey := fmt.Sprintf("team:%s:upcoming_home_games", teamID)
	now := time.Now().Unix()

	// Remove games with scores less than now
	err := r.client.ZRemRangeByScore(ctx, zsetKey, "0", fmt.Sprintf("(%d", now)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete past games for team %s: %v", teamID, err)
	}

	return nil
}

// StoreGame implements UpcomingGamesManager.
func (r *redisUpcomingGamesManager) StoreGame(ctx context.Context, gameID string, game Game) error {
	hashKey := fmt.Sprintf("game:%s", gameID)
	gameData, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("failed to marshal game data: %v", err)
	}

	err = r.client.Set(ctx, hashKey, gameData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store gameData: %v", err)
	}
	return nil
}

// GetGame implements UpcomingGamesManager.
func (r *redisUpcomingGamesManager) GetGame(ctx context.Context, gameID string) (Game, error) {
	hashKey := fmt.Sprintf("game:%s", gameID)

	var game Game
	gameData, err := r.client.Get(ctx, hashKey).Result()
	if err != nil {
		if err == redis.Nil {
			return game, fmt.Errorf("game %s does not exist", gameID)
		}
		return game, fmt.Errorf("failed to get game data: %v", err)
	}

	err = json.Unmarshal([]byte(gameData), &game)
	if err != nil {
		return game, fmt.Errorf("failed to unmarshal game data: %v", err)
	}

	return game, nil
}

// UpdateGame implements UpcomingGamesManager.
func (r *redisUpcomingGamesManager) UpdateGame(ctx context.Context, gameID string, fields map[string]interface{}) error {
	// Get the current game data
	game, err := r.GetGame(ctx, gameID)
	if err != nil {
		return err
	}

	// Update the fields
	for key, value := range fields {
		switch key {
		case "game_id":
			game.GameID = value.(string)
		case "home_team_odds":
			game.HomeTeamOdds = value.(string)
		case "venue":
			game.Venue = value.(string)
		case "time":
			game.Time = value.(string)
		case "lowest_ticket_price":
			game.LowestTicketPrice = value.(string)
		case "injured_players":
			// Ensure value is a slice of injured players
			if players, ok := value.([]struct {
				Team       string `json:"team"`
				PlayerName string `json:"player_name"`
				Status     string `json:"status"`
			}); ok {
				game.InjuredPlayers = players
			} else {
				return fmt.Errorf("invalid type for injured_players")
			}
		default:
			return fmt.Errorf("unknown field %s", key)
		}
	}

	// Store the updated game
	return r.StoreGame(ctx, gameID, game)
}

// DeleteGame implements UpcomingGamesManager.
func (r *redisUpcomingGamesManager) DeleteGame(ctx context.Context, gameID string) error {
	hashKey := fmt.Sprintf("game:%s", gameID)
	err := r.client.Del(ctx, hashKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete game %s: %v", gameID, err)
	}
	return nil
}

// NewUpcomingGamesManager initializes a new Redis client and returns an UpcomingGamesManager.
func NewUpcomingGamesManager(addr string) (UpcomingGamesManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &redisUpcomingGamesManager{client: client}, nil
}
