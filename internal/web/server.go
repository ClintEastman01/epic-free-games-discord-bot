package web

import (
	"fmt"
	"free-games-scrape/internal/database"
	"free-games-scrape/internal/service"
	"html/template"
	"log"
	"net/http"
	"time"
)

// WebServer handles HTTP requests for documentation
type WebServer struct {
	port        string
	gameService *service.GameService
	db          *database.Database
	templates   *template.Template
}

// NewWebServer creates a new web server instance
func NewWebServer(port string, gameService *service.GameService, db *database.Database) *WebServer {
	return &WebServer{
		port:        port,
		gameService: gameService,
		db:          db,
	}
}

// Start starts the web server
func (ws *WebServer) Start() error {
	// Load templates
	if err := ws.loadTemplates(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Setup routes
	ws.setupRoutes()

	log.Printf("Starting web server on port %s", ws.port)
	log.Printf("Documentation available at: http://localhost%s/help", ws.port)
	log.Printf("Bot invite page available at: http://localhost%s/invite", ws.port)

	return http.ListenAndServe(ws.port, nil)
}

// loadTemplates loads HTML templates
func (ws *WebServer) loadTemplates() error {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("January 2, 2006 at 3:04 PM")
		},
	}).ParseGlob("web/templates/*.html")
	if err != nil {
		// If templates don't exist, create them inline
		ws.createInlineTemplates()
		return nil
	}

	ws.templates = tmpl
	return nil
}

// setupRoutes configures HTTP routes
func (ws *WebServer) setupRoutes() {
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Documentation endpoints
	http.HandleFunc("/", ws.handleHome)
	http.HandleFunc("/help", ws.handleHelp)
	http.HandleFunc("/invite", ws.handleInvite)
	http.HandleFunc("/api/status", ws.handleAPIStatus)
	http.HandleFunc("/api/games", ws.handleAPIGames)
}

// Page data structures
type PageData struct {
	Title       string
	Description string
	ServerCount int
	GameCount   int
	LastUpdate  time.Time
	Games       interface{}
}

type StatusData struct {
	Status      string    `json:"status"`
	ServerCount int       `json:"server_count"`
	GameCount   int       `json:"game_count"`
	LastUpdate  time.Time `json:"last_update"`
	Uptime      string    `json:"uptime"`
}

// Route handlers
func (ws *WebServer) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/help", http.StatusMovedPermanently)
}

func (ws *WebServer) handleHelp(w http.ResponseWriter, r *http.Request) {
	data := ws.getPageData("Free Games Bot - Complete Documentation")
	ws.renderTemplate(w, "documentation", data)
}

func (ws *WebServer) handleInvite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Note: Replace YOUR_BOT_CLIENT_ID with your actual bot's client ID
	clientID := "1393810058441392230"
	permissions := "2147485696"
	inviteURL := fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&permissions=%s&scope=bot%%20applications.commands", clientID, permissions)

	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invite Free Games Bot</title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); margin: 0; padding: 20px; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
        .container { background: white; border-radius: 12px; box-shadow: 0 8px 32px rgba(0,0,0,0.1); padding: 40px; max-width: 600px; text-align: center; }
        .logo { font-size: 4rem; margin-bottom: 20px; }
        h1 { color: #7289da; margin-bottom: 10px; }
        .subtitle { color: #72767d; margin-bottom: 30px; font-size: 1.1rem; }
        .invite-button { background: #7289da; color: white; padding: 15px 30px; border: none; border-radius: 8px; font-size: 1.1rem; font-weight: bold; text-decoration: none; display: inline-block; transition: background 0.3s ease; margin-bottom: 30px; }
        .invite-button:hover { background: #5b6eae; }
        .permissions { background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; text-align: left; }
        .permissions h3 { color: #7289da; margin-top: 0; }
        .permissions ul { margin: 0; padding-left: 20px; }
        .permissions li { margin-bottom: 5px; }
        .steps { text-align: left; }
        .steps h3 { color: #7289da; }
        .steps ol { padding-left: 20px; }
        .steps li { margin-bottom: 10px; line-height: 1.5; }
        .note { background: #fff3cd; border: 1px solid #ffeaa7; color: #856404; padding: 15px; border-radius: 8px; margin-top: 20px; }
        .warning { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; padding: 15px; border-radius: 8px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🎮</div>
        <h1>Invite Free Games Bot</h1>
        <p class="subtitle">Add the Free Games Bot to your Discord server and never miss a free game again!</p>
        
        <div class="warning">
            <strong>⚠️ Setup Required:</strong> You need to configure your bot's Client ID first. See the setup guide below.
        </div>
        
        <a href="%s" class="invite-button" target="_blank">📨 Invite Bot to Discord</a>
        
        <div class="permissions">
            <h3>🔐 Required Permissions</h3>
            <p>The bot needs these permissions to function properly:</p>
            <ul>
                <li>✅ Send Messages - To send game notifications</li>
                <li>✅ Use Slash Commands - For /setup, /games, etc.</li>
                <li>✅ Embed Links - For rich game information</li>
                <li>✅ Attach Files - For game images</li>
                <li>✅ Read Message History - For command processing</li>
                <li>✅ Add Reactions - For interactive features</li>
            </ul>
        </div>
        
        <div class="steps">
            <h3>🚀 Quick Setup</h3>
            <ol>
                <li><strong>Click the invite button above</strong> to add the bot to your server</li>
                <li><strong>Select your Discord server</strong> from the dropdown</li>
                <li><strong>Confirm the permissions</strong> and click "Authorize"</li>
                <li><strong>Run the setup command:</strong> <code>/setup #your-channel</code></li>
                <li><strong>Test it:</strong> <code>/games</code> to see current free games</li>
            </ol>
        </div>
        
        <div class="note">
            <strong>📝 Note:</strong> You need Administrator permissions in your Discord server to add bots and run the setup command.
        </div>
        
        <p style="margin-top: 30px; color: #72767d;">
            <a href="/help" style="color: #7289da; text-decoration: none;">📖 View Documentation</a> | 
            <a href="/api/status" style="color: #7289da; text-decoration: none;">📊 Bot Status</a>
        </p>
    </div>
</body>
</html>`, inviteURL)
}

func (ws *WebServer) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	serverCount, _ := ws.db.GetServerCount()
	games, _ := ws.gameService.GetActiveGames()
	gameCount := len(games.FreeNow) + len(games.ComingSoon)

	status := StatusData{
		Status:      "online",
		ServerCount: serverCount,
		GameCount:   gameCount,
		LastUpdate:  time.Now(),
		Uptime:      "24/7",
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, `{
		"status": "%s",
		"server_count": %d,
		"game_count": %d,
		"last_update": "%s",
		"uptime": "%s"
	}`, status.Status, status.ServerCount, status.GameCount,
		status.LastUpdate.Format(time.RFC3339), status.Uptime)
}

func (ws *WebServer) handleAPIGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	games, err := ws.gameService.GetActiveGames()
	if err != nil {
		http.Error(w, "Failed to get games", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `{
		"free_now": %d,
		"coming_soon": %d,
		"total": %d,
		"last_updated": "%s"
	}`, len(games.FreeNow), len(games.ComingSoon),
		len(games.FreeNow)+len(games.ComingSoon), time.Now().Format(time.RFC3339))
}

// Helper functions
func (ws *WebServer) getPageData(title string) PageData {
	serverCount, _ := ws.db.GetServerCount()
	games, _ := ws.gameService.GetActiveGames()
	gameCount := len(games.FreeNow) + len(games.ComingSoon)

	return PageData{
		Title:       title,
		Description: "Epic Games Store Free Games Discord Bot",
		ServerCount: serverCount,
		GameCount:   gameCount,
		LastUpdate:  time.Now(),
		Games:       games,
	}
}

func (ws *WebServer) renderTemplate(w http.ResponseWriter, tmplName string, data PageData) {
	if ws.templates != nil {
		err := ws.templates.ExecuteTemplate(w, tmplName+".html", data)
		if err != nil {
			log.Printf("Template error: %v", err)
			ws.renderInlineTemplate(w, tmplName, data)
		}
	} else {
		ws.renderInlineTemplate(w, tmplName, data)
	}
}

// createInlineTemplates creates templates when files don't exist
func (ws *WebServer) createInlineTemplates() {
	log.Println("Using inline templates")
}

// renderInlineTemplate renders templates inline when files are not available
func (ws *WebServer) renderInlineTemplate(w http.ResponseWriter, tmplName string, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the documentation template inline
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); margin: 0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 30px; padding: 30px; }
        .header-content { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; }
        .logo h1 { color: #7289da; margin: 0; }
        .stats { display: flex; gap: 20px; }
        .stat-item { text-align: center; }
        .stat-number { display: block; font-size: 2rem; font-weight: bold; color: #7289da; }
        .stat-label { color: #72767d; font-size: 0.9rem; }
        .content { background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); padding: 40px; }
        .feature-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin: 30px 0; }
        .feature-card { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; border-radius: 8px; text-align: center; }
        .feature-icon { font-size: 3rem; margin-bottom: 15px; }
        .command-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 15px; margin: 30px 0; }
        .command-item { background: #f8f9fa; padding: 20px; border-radius: 8px; border-left: 4px solid #7289da; }
        .badge { background: #f04747; color: white; padding: 2px 6px; border-radius: 4px; font-size: 0.8rem; }
        code { background: #2c2f33; color: white; padding: 4px 8px; border-radius: 4px; }
        .cta { text-align: center; margin-top: 40px; padding: 30px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; border-radius: 8px; }
        .invite-section { background: #e8f5e8; border: 2px solid #28a745; border-radius: 8px; padding: 20px; margin: 30px 0; text-align: center; }
        .invite-button { background: #28a745; color: white; padding: 12px 24px; border: none; border-radius: 6px; font-size: 1rem; font-weight: bold; text-decoration: none; display: inline-block; margin: 10px; }
        .invite-button:hover { background: #218838; }
    </style>
</head>
<body>
    <div class="container">
        <header class="header">
            <div class="header-content">
                <div class="logo">
                    <h1>🎮 Free Games Bot</h1>
                    <p>Epic Games Store Discord Bot - Complete Documentation</p>
                </div>
                <div class="stats">
                    <div class="stat-item">
                        <span class="stat-number">%d</span>
                        <span class="stat-label">Servers</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">%d</span>
                        <span class="stat-label">Active Games</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">Online</span>
                        <span class="stat-label">Status</span>
                    </div>
                </div>
            </div>
        </header>
        
        <div class="content">
            <div class="invite-section">
                <h2 style="color: #28a745; margin-top: 0;">📨 Invite Bot to Your Server</h2>
                <p>Ready to add the Free Games Bot to your Discord server?</p>
                <a href="/invite" class="invite-button">Get Invite Link</a>
                <a href="#setup" class="invite-button" style="background: #007bff;">Setup Guide</a>
            </div>
            
            <h2 style="color: #7289da; text-align: center; margin-bottom: 30px;">🎮 Free Games Bot Documentation</h2>
            
            <div class="feature-grid">
                <div class="feature-card">
                    <div class="feature-icon">🔄</div>
                    <h3>Automatic Monitoring</h3>
                    <p>Checks Epic Games Store every 6 hours for new free games</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">💾</div>
                    <h3>Smart Caching</h3>
                    <p>SQLite database prevents duplicate notifications</p>
                </div>
                <div class="feature-card">
                    <div class="feature-icon">🎯</div>
                    <h3>Multi-Server Support</h3>
                    <p>Configure different channels for each Discord server</p>
                </div>
            </div>
            
            <h3 id="setup" style="color: #7289da; margin-bottom: 20px;">🚀 Quick Setup</h3>
            <ol style="line-height: 1.8; margin-bottom: 30px;">
                <li><strong>Invite the bot</strong> using the invite link above</li>
                <li><strong>Run setup:</strong> <code>/setup #your-channel</code></li>
                <li><strong>Test it:</strong> <code>/games</code></li>
                <li><strong>Enjoy automatic notifications!</strong></li>
            </ol>
            
            <h3 style="color: #7289da; margin-bottom: 20px;">🎯 Available Commands</h3>
            <div class="command-grid">
                <div class="command-item">
                    <strong>/setup &lt;channel&gt;</strong> <span class="badge">ADMIN</span>
                    <p style="margin: 5px 0 0 0; color: #72767d;">Configure the bot for this server</p>
                </div>
                <div class="command-item">
                    <strong>/games</strong>
                    <p style="margin: 5px 0 0 0; color: #72767d;">Show current free games</p>
                </div>
                <div class="command-item">
                    <strong>/refresh</strong> <span class="badge">ADMIN</span>
                    <p style="margin: 5px 0 0 0; color: #72767d;">Manually refresh games</p>
                </div>
                <div class="command-item">
                    <strong>/status</strong>
                    <p style="margin: 5px 0 0 0; color: #72767d;">Show bot status and configuration</p>
                </div>
            </div>
            
            <h3 style="color: #7289da; margin-bottom: 20px;">✨ Key Features</h3>
            <ul style="line-height: 1.8; margin-bottom: 30px;">
                <li><strong>Automatic Monitoring:</strong> Checks Epic Games Store every 6 hours</li>
                <li><strong>Rich Embeds:</strong> Beautiful Discord messages with game images</li>
                <li><strong>Smart Notifications:</strong> Only notifies about new games, no duplicates</li>
                <li><strong>Multi-Server:</strong> Each server has its own configuration</li>
                <li><strong>Instant Commands:</strong> Fast responses using cached database data</li>
                <li><strong>Cross-Platform:</strong> Works on Windows, macOS, and Linux</li>
            </ul>
            
            <h3 style="color: #7289da; margin-bottom: 20px;">🔌 API Endpoints</h3>
            <ul style="line-height: 1.8; margin-bottom: 30px;">
                <li><strong>GET /help</strong> - This documentation page</li>
                <li><strong>GET /invite</strong> - Bot invitation page</li>
                <li><strong>GET /api/status</strong> - Bot status and statistics</li>
                <li><strong>GET /api/games</strong> - Current game information</li>
            </ul>
            
            <h3 style="color: #7289da; margin-bottom: 20px;">🔧 Troubleshooting</h3>
            <div style="background: #f8f9fa; padding: 20px; border-radius: 8px; border-left: 4px solid #faa61a;">
                <p><strong>Bot not responding?</strong></p>
                <ul style="margin: 10px 0;">
                    <li>Ensure you've run <code>/setup</code> first</li>
                    <li>Check bot permissions (Send Messages, Use Slash Commands, Embed Links)</li>
                    <li>Try commands in the configured channel</li>
                    <li>Wait up to 1 hour for slash commands to sync globally</li>
                </ul>
            </div>
            
            <div class="cta">
                <h3>🎮 Ready to Get Started?</h3>
                <p>Add the Free Games Bot to your Discord server and never miss a free game again!</p>
                <a href="/invite" class="invite-button">Get Invite Link</a>
                <p style="margin-top: 20px; font-size: 0.9rem; opacity: 0.8;">
                    Bot Status: <strong>Online</strong> | 
                    Servers: <strong>%d</strong> | 
                    Active Games: <strong>%d</strong>
                </p>
            </div>
        </div>
    </div>
</body>
</html>`, data.Title, data.ServerCount, data.GameCount, data.ServerCount, data.GameCount)
}

