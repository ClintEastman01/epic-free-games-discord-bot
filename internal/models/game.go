package models

import (
	"fmt"
	"time"
)

// Game represents a free game from Epic Games Store
type Game struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
	Status   string `json:"status"`
	FreeFrom string `json:"free_from"`
	FreeTo   string `json:"free_to"`
}

// GameStatus constants for game availability
const (
	StatusFreeNow    = "Free Now"
	StatusComingSoon = "Coming Soon"
)

// IsActive checks if a "Free Now" game is still active
func (g *Game) IsActive() bool {
	if g.Status != StatusFreeNow || g.FreeTo == "" {
		return false
	}

	currentYear := time.Now().Year()
	// Parse FreeTo date (e.g., "Jul 17" -> "Jul 17 2025")
	freeToDate, err := time.Parse("Jan 02 2006", g.FreeTo+" "+fmt.Sprintf("%d", currentYear))
	if err != nil {
		return false
	}
	
	// Add one day to account for end-of-day expiration
	freeToDate = freeToDate.Add(24 * time.Hour)
	return time.Now().Before(freeToDate)
}

// GameCollection represents a collection of games categorized by status
type GameCollection struct {
	FreeNow    []Game
	ComingSoon []Game
}

// NewGameCollection creates a new GameCollection from a slice of games
func NewGameCollection(games []Game) *GameCollection {
	collection := &GameCollection{
		FreeNow:    make([]Game, 0),
		ComingSoon: make([]Game, 0),
	}

	for _, game := range games {
		switch game.Status {
		case StatusFreeNow:
			collection.FreeNow = append(collection.FreeNow, game)
		case StatusComingSoon:
			collection.ComingSoon = append(collection.ComingSoon, game)
		}
	}

	return collection
}

// HasActiveFreeGames checks if there are any active "Free Now" games
func (gc *GameCollection) HasActiveFreeGames() bool {
	for _, game := range gc.FreeNow {
		if game.IsActive() {
			return true
		}
	}
	return false
}