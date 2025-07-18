package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"free-games-scrape/internal/models"
)

// ServerConfig represents a Discord server configuration
type ServerConfig struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

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

	if err := database.createServerConfigTable(); err != nil {
		return nil, fmt.Errorf("failed to create server config table: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// createTables creates the necessary database tables
func (d *Database) createTables() error {
	// First check if the table exists
	var tableName string
	err := d.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='games'").Scan(&tableName)
	
	if err == nil {
		// Table exists, check if we need to migrate
		var hasUniqueConstraint bool
		err = d.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_games_title_free_to'").Scan(&hasUniqueConstraint)
		
		if err == nil && !hasUniqueConstraint {
			// Need to migrate the table structure
			log.Println("Migrating games table to support composite key...")
			
			// Create a new table with the desired structure
			_, err = d.db.Exec(`
				CREATE TABLE IF NOT EXISTS games_new (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					title TEXT NOT NULL,
					image_url TEXT,
					status TEXT NOT NULL,
					free_from TEXT,
					free_to TEXT,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(title, free_to)
				);
				
				-- Copy data from old table
				INSERT OR IGNORE INTO games_new 
					(id, title, image_url, status, free_from, free_to, created_at, updated_at, last_seen)
				SELECT 
					id, title, image_url, status, free_from, free_to, created_at, updated_at, last_seen
				FROM games;
				
				-- Drop old table
				DROP TABLE games;
				
				-- Rename new table
				ALTER TABLE games_new RENAME TO games;
				
				-- Recreate indexes
				CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
				CREATE INDEX IF NOT EXISTS idx_games_title ON games(title);
				CREATE INDEX IF NOT EXISTS idx_games_last_seen ON games(last_seen);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_games_title_free_to ON games(title, free_to);
			`)
			
			if err != nil {
				return fmt.Errorf("failed to migrate games table: %w", err)
			}
			
			log.Println("Successfully migrated games table")
			return nil
		}
	}
	
	// Create table if it doesn't exist or if there was an error checking
	query := `
	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		image_url TEXT,
		status TEXT NOT NULL,
		free_from TEXT,
		free_to TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(title, free_to)
	);

	CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
	CREATE INDEX IF NOT EXISTS idx_games_title ON games(title);
	CREATE INDEX IF NOT EXISTS idx_games_last_seen ON games(last_seen);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_games_title_free_to ON games(title, free_to);
	`

	_, err = d.db.Exec(query)
	return err
}

// SaveGames saves or updates games in the database
func (d *Database) SaveGames(games []models.Game) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, mark all games as not seen in this update
	_, err = tx.Exec(`UPDATE games SET last_seen = datetime('now', '-1 day') WHERE 1=1`)
	if err != nil {
		return fmt.Errorf("failed to mark games as not seen: %w", err)
	}

	// Now insert or update each game
	// We'll use title AND free_to as a composite key to handle cases where the same game becomes free again
	stmt, err := tx.Prepare(`
		INSERT INTO games (title, image_url, status, free_from, free_to, updated_at, last_seen)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(title, free_to) DO UPDATE SET
			image_url = excluded.image_url,
			status = excluded.status,
			free_from = excluded.free_from,
			updated_at = CURRENT_TIMESTAMP,
			last_seen = CURRENT_TIMESTAMP
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

// GetServerCount returns the total number of configured servers
func (d *Database) GetServerCount() (int, error) {
	query := `SELECT COUNT(*) FROM server_configs WHERE active = 1`
	
	var count int
	err := d.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get server count: %w", err)
	}
	
	return count, nil
}

// GetAllActiveServerConfigs returns all active server configurations
func (d *Database) GetAllActiveServerConfigs() ([]*ServerConfig, error) {
	query := `
		SELECT guild_id, channel_id, created_at, updated_at
		FROM server_configs 
		WHERE active = 1
		ORDER BY created_at
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query server configs: %w", err)
	}
	defer rows.Close()
	
	var configs []*ServerConfig
	for rows.Next() {
		var config ServerConfig
		err := rows.Scan(&config.GuildID, &config.ChannelID, &config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server config: %w", err)
		}
		configs = append(configs, &config)
	}
	
	return configs, nil
}

// GetServerConfig retrieves server configuration by guild ID
func (d *Database) GetServerConfig(guildID string) (*ServerConfig, error) {
	query := `
		SELECT guild_id, channel_id, created_at, updated_at
		FROM server_configs 
		WHERE guild_id = ? AND active = 1
		LIMIT 1
	`
	
	var config ServerConfig
	err := d.db.QueryRow(query, guildID).Scan(
		&config.GuildID, &config.ChannelID, &config.CreatedAt, &config.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server config: %w", err)
	}
	
	return &config, nil
}

// SaveServerConfig saves or updates server configuration
func (d *Database) SaveServerConfig(guildID, channelID string) error {
	query := `
		INSERT OR REPLACE INTO server_configs (guild_id, channel_id, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	
	_, err := d.db.Exec(query, guildID, channelID)
	if err != nil {
		return fmt.Errorf("failed to save server config: %w", err)
	}
	
	log.Printf("Saved server config for guild %s, channel %s", guildID, channelID)
	return nil
}

// DeactivateServerConfig deactivates a server configuration
func (d *Database) DeactivateServerConfig(guildID, channelID string) error {
	query := `UPDATE server_configs SET active = 0, updated_at = CURRENT_TIMESTAMP WHERE guild_id = ? AND channel_id = ?`
	_, err := d.db.Exec(query, guildID, channelID)
	if err != nil {
		return fmt.Errorf("failed to deactivate server config: %w", err)
	}
	
	log.Printf("Deactivated server config for guild %s, channel %s", guildID, channelID)
	return nil
}

// createServerConfigTable creates the server_configs table
func (d *Database) createServerConfigTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS server_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guild_id TEXT NOT NULL UNIQUE,
		channel_id TEXT NOT NULL,
		active INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_server_configs_guild_id ON server_configs(guild_id);
	CREATE INDEX IF NOT EXISTS idx_server_configs_active ON server_configs(active);
	`

	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create server_configs table: %w", err)
	}

	log.Println("Server configs table created/verified")
	return nil
}