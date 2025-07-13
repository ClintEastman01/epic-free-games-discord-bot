package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"free-games-scrape/internal/models"
)

// Database handles SQLite operations
type Database struct {
	db *sql.DB
}

// New creates a new database connection and initializes tables
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// createTables creates the necessary database tables
func (d *Database) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		image_url TEXT,
		status TEXT NOT NULL,
		free_from TEXT,
		free_to TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
	CREATE INDEX IF NOT EXISTS idx_games_title ON games(title);
	CREATE INDEX IF NOT EXISTS idx_games_last_seen ON games(last_seen);
	`

	_, err := d.db.Exec(query)
	return err
}

// SaveGames saves or updates games in the database
func (d *Database) SaveGames(games []models.Game) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO games (title, image_url, status, free_from, free_to, updated_at, last_seen)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, game := range games {
		_, err := stmt.Exec(game.Title, game.ImageURL, game.Status, game.FreeFrom, game.FreeTo)
		if err != nil {
			return fmt.Errorf("failed to save game %s: %w", game.Title, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Saved %d games to database", len(games))
	return nil
}

// GetActiveGames returns all currently active games
func (d *Database) GetActiveGames() ([]models.Game, error) {
	query := `
		SELECT title, image_url, status, free_from, free_to
		FROM games
		WHERE status IN ('Free Now', 'Coming Soon')
		AND last_seen > datetime('now', '-7 days')
		ORDER BY 
			CASE 
				WHEN status = 'Free Now' THEN 1 
				WHEN status = 'Coming Soon' THEN 2 
				ELSE 3 
			END,
			title
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active games: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(&game.Title, &game.ImageURL, &game.Status, &game.FreeFrom, &game.FreeTo)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}

	return games, nil
}

// GetNewGames returns games that are new since the last check
func (d *Database) GetNewGames(since time.Time) ([]models.Game, error) {
	query := `
		SELECT title, image_url, status, free_from, free_to
		FROM games
		WHERE created_at > ?
		AND status IN ('Free Now', 'Coming Soon')
		ORDER BY 
			CASE 
				WHEN status = 'Free Now' THEN 1 
				WHEN status = 'Coming Soon' THEN 2 
				ELSE 3 
			END,
			title
	`

	rows, err := d.db.Query(query, since.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("failed to query new games: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var game models.Game
		err := rows.Scan(&game.Title, &game.ImageURL, &game.Status, &game.FreeFrom, &game.FreeTo)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}

	return games, nil
}

// CleanupOldGames removes games that haven't been seen for more than 30 days
func (d *Database) CleanupOldGames() error {
	query := `DELETE FROM games WHERE last_seen < datetime('now', '-30 days')`
	
	result, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup old games: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d old games from database", rowsAffected)
	}

	return nil
}

// GetGameByTitle retrieves a specific game by title
func (d *Database) GetGameByTitle(title string) (*models.Game, error) {
	query := `
		SELECT title, image_url, status, free_from, free_to
		FROM games
		WHERE title = ?
		LIMIT 1
	`

	var game models.Game
	err := d.db.QueryRow(query, title).Scan(
		&game.Title, &game.ImageURL, &game.Status, &game.FreeFrom, &game.FreeTo,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game by title: %w", err)
	}

	return &game, nil
}