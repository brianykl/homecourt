package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"homecourt-api/games"
	"net/http"
)

type GetRequest struct {
	Team string
}

var Manager games.GamesManager

type GetResponse struct {
	Games []map[string]string `json:"games"`
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	var req GetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	team := req.Team
	upcomingGamesKeys, err := Manager.GetUpcomingGames(context.Background(), team, 5)
	if err != nil {
		http.Error(w, "failed to fetch upcoming games", http.StatusInternalServerError)
		return
	}

	var games []map[string]string

	for _, gameID := range upcomingGamesKeys {
		gameData, err := Manager.GetGame(context.Background(), gameID)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch game data for key: %s", gameID), http.StatusInternalServerError)
			return
		}
		games = append(games, gameData)
	}

	// Construct response
	response := GetResponse{
		Games: games,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
