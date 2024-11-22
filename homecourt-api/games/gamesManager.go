package games

import (
	"context"
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

// GamesManager defines the methods for managing  games and game details.
type GamesManager interface {
	//  games per team (ZSET of game IDs)

	CreateOrUpdateGame(ctx context.Context, gameKey string, fields map[string]interface{}) error
	AddUpcomingGame(ctx context.Context, zsetKey, gameID string, score int64) error
	GetUpcomingGames(ctx context.Context, teamID string, count int64) ([]string, error)
	GameExists(ctx context.Context, gameID string) (bool, error)
	GetGame(ctx context.Context, gameID string) (map[string]string, error)
	RemovePastGames(ctx context.Context, teamID string) error

	// AddGame(ctx context.Context, teamID string, gameID string, startTime time.Time) error
	// GetGames(ctx context.Context, teamID string, count int) ([]string, error)
	// DeleteGames(ctx context.Context, teamID string) error

	// // Individual game details
	// StoreGame(ctx context.Context, gameID string, game Game) error
	// GetGame(ctx context.Context, gameID string) (Game, error)
	// UpdateGame(ctx context.Context, gameID string, fields map[string]interface{}) error
	// DeleteGame(ctx context.Context, gameID string) error
}

// redisGamesManager manages the Redis connection and operations.
type redisGamesManager struct {
	client *redis.Client
}

func NewGamesManager(addr string) (GamesManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &redisGamesManager{client: client}, nil
}

func (r *redisGamesManager) CreateOrUpdateGame(ctx context.Context, gameKey string, fields map[string]interface{}) error {
	err := r.client.HSet(ctx, gameKey, fields).Err()
	if err != nil {
		return fmt.Errorf("failed to create or update game %s: %v", gameKey, err)
	}
	return nil
}

func (r *redisGamesManager) GameExists(ctx context.Context, gameID string) (bool, error) {
	// Use the game ID to construct the Redis key
	gameKey := fmt.Sprintf("game:%s", gameID)

	// Check if the key exists in Redis
	exists, err := r.client.Exists(ctx, gameKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check game existence: %v", err)
	}

	// Redis EXISTS returns 1 if the key exists, 0 otherwise
	return exists > 0, nil
}

func (r *redisGamesManager) AddUpcomingGame(ctx context.Context, zsetKey, gameID string, score int64) error {
	err := r.client.ZAdd(ctx, zsetKey, redis.Z{
		Score:  float64(score),
		Member: gameID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add game to  home games: %v", err)
	}
	return nil
}

func (r *redisGamesManager) GetUpcomingGames(ctx context.Context, teamID string, count int64) ([]string, error) {
	zsetKey := fmt.Sprintf("team:%s:upcoming_home_games", teamID)
	now := time.Now().Unix()
	gameIDs, err := r.client.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%d", now),
		Max:    "+inf",
		Offset: 0,
		Count:  count,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get  games: %v", err)
	}
	return gameIDs, nil
}

func (r *redisGamesManager) GetGame(ctx context.Context, gameID string) (map[string]string, error) {
	gameKey := fmt.Sprintf("game:%s", gameID)
	gameData, err := r.client.HGetAll(ctx, gameKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get game data: %v", err)
	}
	return gameData, nil
}

func (r *redisGamesManager) RemovePastGames(ctx context.Context, teamID string) error {
	zsetKey := fmt.Sprintf("team:%s:_home_games", teamID)
	now := time.Now().Unix()
	// Remove games with scores less than current time
	err := r.client.ZRemRangeByScore(ctx, zsetKey, "0", fmt.Sprintf("(%d", now)).Err()
	if err != nil {
		return fmt.Errorf("failed to remove past games: %v", err)
	}
	return nil
}

// func (r *redisGamesManager) RemovePastGames(ctx context.Context, teamID string) error {
// 	zsetKey := fmt.Sprintf("team:%s:_home_games", teamID)
// 	now := time.Now().Unix()
// 	// Remove games with scores less than current time
// 	err := r.client.ZRemRangeByScore(ctx, zsetKey, "0", fmt.Sprintf("(%d", now)).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to remove past games: %v", err)
// 	}
// 	return nil
// }

// func (r *redisGamesManager) GetGame(ctx context.Context, gameID string) (map[string]string, error) {
//     gameKey := fmt.Sprintf("game:%s", gameID)
//     gameData, err := r.client.HGetAll(ctx, gameKey).Result()
//     if err != nil {
//         return nil, fmt.Errorf("failed to get game data: %v", err)
//     }
//     return gameData, nil
// }

// // AddGame implements GamesManager.
// func (r *redisGamesManager) AddGame(ctx context.Context, teamID string, gameID string, startTime time.Time) error {
// 	zsetKey := fmt.Sprintf("team:%s:_home_games", teamID)
// 	score := float64(startTime.Unix())

// 	err := r.client.ZAdd(ctx, zsetKey, redis.Z{
// 		Score:  score,
// 		Member: gameID,
// 	}).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to add game to ZSET: %v", err)
// 	}

// 	return nil
// }

// // GetGames implements GamesManager.
// func (r *redisGamesManager) GetGames(ctx context.Context, teamID string, count int) ([]string, error) {
// 	zsetKey := fmt.Sprintf("team:%s:_home_games", teamID)
// 	now := time.Now().Unix()

// 	gameIDs, err := r.client.ZRangeByScore(ctx, zsetKey, &redis.ZRangeBy{
// 		Min:    fmt.Sprintf("%d", now),
// 		Max:    "+inf",
// 		Offset: 0,
// 		Count:  int64(count),
// 	}).Result()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get  games: %v", err)
// 	}

// 	return gameIDs, nil
// }

// // DeleteGames implements GamesManager.
// // may want to also delete the actual game data for each deleted game down the line as well
// func (r *redisGamesManager) DeleteGames(ctx context.Context, teamID string) error {
// 	zsetKey := fmt.Sprintf("team:%s:_home_games", teamID)
// 	now := time.Now().Unix()

// 	// Remove games with scores less than now
// 	err := r.client.ZRemRangeByScore(ctx, zsetKey, "0", fmt.Sprintf("(%d", now)).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to delete past games for team %s: %v", teamID, err)
// 	}

// 	return nil
// }

// // StoreGame implements GamesManager.
// func (r *redisGamesManager) StoreGame(ctx context.Context, gameID string, game Game) error {
// 	hashKey := fmt.Sprintf("game:%s", gameID)
// 	gameData, err := json.Marshal(game)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal game data: %v", err)
// 	}

// 	err = r.client.Set(ctx, hashKey, gameData, 0).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to store gameData: %v", err)
// 	}
// 	return nil
// }

// // GetGame implements GamesManager.
// func (r *redisGamesManager) GetGame(ctx context.Context, gameID string) (Game, error) {
// 	hashKey := fmt.Sprintf("game:%s", gameID)

// 	var game Game
// 	gameData, err := r.client.Get(ctx, hashKey).Result()
// 	if err != nil {
// 		if err == redis.Nil {
// 			return game, fmt.Errorf("game %s does not exist", gameID)
// 		}
// 		return game, fmt.Errorf("failed to get game data: %v", err)
// 	}

// 	err = json.Unmarshal([]byte(gameData), &game)
// 	if err != nil {
// 		return game, fmt.Errorf("failed to unmarshal game data: %v", err)
// 	}

// 	return game, nil
// }

// // UpdateGame implements GamesManager.
// func (r *redisGamesManager) UpdateGame(ctx context.Context, gameID string, fields map[string]interface{}) error {
// 	// Get the current game data
// 	game, err := r.GetGame(ctx, gameID)
// 	if err != nil {
// 		return err
// 	}

// 	// Update the fields
// 	for key, value := range fields {
// 		switch key {
// 		case "game_id":
// 			game.GameID = value.(string)
// 		case "home_team_odds":
// 			game.HomeTeamOdds = value.(string)
// 		case "venue":
// 			game.Venue = value.(string)
// 		case "time":
// 			game.Time = value.(string)
// 		case "lowest_ticket_price":
// 			game.LowestTicketPrice = value.(string)
// 		case "injured_players":
// 			// Ensure value is a slice of injured players
// 			if players, ok := value.([]struct {
// 				Team       string `json:"team"`
// 				PlayerName string `json:"player_name"`
// 				Status     string `json:"status"`
// 			}); ok {
// 				game.InjuredPlayers = players
// 			} else {
// 				return fmt.Errorf("invalid type for injured_players")
// 			}
// 		default:
// 			return fmt.Errorf("unknown field %s", key)
// 		}
// 	}

// 	// Store the updated game
// 	return r.StoreGame(ctx, gameID, game)
// }

// // DeleteGame implements GamesManager.
// func (r *redisGamesManager) DeleteGame(ctx context.Context, gameID string) error {
// 	hashKey := fmt.Sprintf("game:%s", gameID)
// 	err := r.client.Del(ctx, hashKey).Err()
// 	if err != nil {
// 		return fmt.Errorf("failed to delete game %s: %v", gameID, err)
// 	}
// 	return nil
// }

// NewGamesManager initializes a new Redis client and returns an GamesManager.
