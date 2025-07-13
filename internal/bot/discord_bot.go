package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"free-games-scrape/internal/config"
	"free-games-scrape/internal/models"
	"free-games-scrape/internal/service"
)

// DiscordBot handles Discord interactions
type DiscordBot struct {
	session     *discordgo.Session
	config      *config.DiscordConfig
	channelID   string
	gameService *service.GameService
}

// NewDiscordBot creates a new Discord bot instance
func NewDiscordBot(cfg *config.DiscordConfig, gameService *service.GameService) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &DiscordBot{
		session:     session,
		config:      cfg,
		channelID:   cfg.ChannelID,
		gameService: gameService,
	}

	// Set up event handlers
	bot.setupEventHandlers()

	return bot, nil
}

// Start opens the Discord connection
func (b *DiscordBot) Start() error {
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord connection: %w", err)
	}
	log.Println("Discord bot is now running")
	return nil
}

// Stop closes the Discord connection
func (b *DiscordBot) Stop() error {
	log.Println("Shutting down Discord bot")
	return b.session.Close()
}

// setupEventHandlers configures Discord event handlers
func (b *DiscordBot) setupEventHandlers() {
	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Bot is ready! Logged in as: %v#%v", r.User.Username, r.User.Discriminator)
	})

	b.session.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		log.Printf("Joined guild: %s (ID: %s)", g.Name, g.ID)
	})

	// Add message handler for commands
	b.session.AddHandler(b.messageHandler)
}

// messageHandler handles incoming Discord messages
func (b *DiscordBot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Only respond in the configured channel
	if m.ChannelID != b.channelID {
		return
	}

	// Check for commands
	content := strings.TrimSpace(m.Content)
	if !strings.HasPrefix(content, "!") {
		return
	}

	command := strings.ToLower(strings.Fields(content)[0])
	
	switch command {
	case "!games", "!freegames":
		b.handleGamesCommand(s, m)
	case "!refresh", "!update":
		b.handleRefreshCommand(s, m)
	case "!help":
		b.handleHelpCommand(s, m)
	}
}

// handleGamesCommand shows current free games from database
func (b *DiscordBot) handleGamesCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	games, err := b.gameService.GetActiveGames()
	if err != nil {
		b.SendErrorMessage(fmt.Sprintf("Failed to get games: %v", err))
		return
	}

	if len(games.FreeNow) == 0 && len(games.ComingSoon) == 0 {
		b.SendSimpleMessage("No free games currently available in the database.")
		return
	}

	if err := b.SendGameUpdates(games); err != nil {
		b.SendErrorMessage(fmt.Sprintf("Failed to send game updates: %v", err))
	}
}

// handleRefreshCommand manually triggers a refresh
func (b *DiscordBot) handleRefreshCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	b.SendSimpleMessage("Refreshing games from Epic Games Store...")
	
	if err := b.gameService.RefreshGames(); err != nil {
		b.SendErrorMessage(fmt.Sprintf("Failed to refresh games: %v", err))
		return
	}

	games, err := b.gameService.GetActiveGames()
	if err != nil {
		b.SendErrorMessage(fmt.Sprintf("Failed to get updated games: %v", err))
		return
	}

	b.SendSimpleMessage("Games refreshed successfully!")
	
	if len(games.FreeNow) > 0 || len(games.ComingSoon) > 0 {
		if err := b.SendGameUpdates(games); err != nil {
			b.SendErrorMessage(fmt.Sprintf("Failed to send game updates: %v", err))
		}
	} else {
		b.SendSimpleMessage("No free games found after refresh.")
	}
}

// handleHelpCommand shows available commands
func (b *DiscordBot) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Free Games Bot Commands",
		Description: "Available commands for the Epic Games Free Games Bot:",
		Color:       0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "!games or !freegames",
				Value:  "Show current free games from the database",
				Inline: false,
			},
			{
				Name:   "!refresh or !update",
				Value:  "Manually refresh games from Epic Games Store",
				Inline: false,
			},
			{
				Name:   "!help",
				Value:  "Show this help message",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store - Free Games Bot",
		},
	}

	_, err := s.ChannelMessageSendEmbed(b.channelID, embed)
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

// SendGameUpdates sends game updates to the configured Discord channel
func (b *DiscordBot) SendGameUpdates(gameCollection *models.GameCollection) error {
	if err := b.sendFreeNowGames(gameCollection.FreeNow); err != nil {
		return fmt.Errorf("error sending Free Now games: %w", err)
	}

	if err := b.sendComingSoonGames(gameCollection.ComingSoon); err != nil {
		return fmt.Errorf("error sending Coming Soon games: %w", err)
	}

	return nil
}

// sendFreeNowGames sends "Free Now" games to Discord with images displayed
func (b *DiscordBot) sendFreeNowGames(games []models.Game) error {
	if len(games) == 0 {
		return nil
	}

	// Send each game as a separate embed to display images properly
	for i, game := range games {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Free Game Available Now! (%d/%d)", i+1, len(games)),
			Description: fmt.Sprintf("**%s** is currently free on Epic Games Store!", game.Title),
			Color:       0x00ff00, // Green color
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Epic Games Store - Free Games Bot",
			},
		}

		// Add game image as the main embed image (this displays the actual image)
		if game.ImageURL != "" {
			embed.Image = &discordgo.MessageEmbedImage{
				URL: game.ImageURL,
			}
		}

		// Add game details as fields
		if game.Status != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Status",
				Value:  game.Status,
				Inline: true,
			})
		}

		if game.FreeTo != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Free Until",
				Value:  game.FreeTo,
				Inline: true,
			})
		}

		_, err := b.session.ChannelMessageSendEmbed(b.channelID, embed)
		if err != nil {
			return fmt.Errorf("error sending Free Now message for %s: %w", game.Title, err)
		}
	}

	log.Printf("Sent %d Free Now games to Discord with images", len(games))
	return nil
}

// sendComingSoonGames sends "Coming Soon" games to Discord with images displayed
func (b *DiscordBot) sendComingSoonGames(games []models.Game) error {
	if len(games) == 0 {
		return nil
	}

	// Send each game as a separate embed to display images properly
	for i, game := range games {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Free Game Coming Soon! (%d/%d)", i+1, len(games)),
			Description: fmt.Sprintf("**%s** will be free soon on Epic Games Store!", game.Title),
			Color:       0x0099ff, // Blue color
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Epic Games Store - Free Games Bot",
			},
		}

		// Add game image as the main embed image (this displays the actual image)
		if game.ImageURL != "" {
			embed.Image = &discordgo.MessageEmbedImage{
				URL: game.ImageURL,
			}
		}

		// Add game details as fields
		if game.Status != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Status",
				Value:  game.Status,
				Inline: true,
			})
		}

		if game.FreeFrom != "" && game.FreeTo != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Free Period",
				Value:  fmt.Sprintf("%s - %s", game.FreeFrom, game.FreeTo),
				Inline: true,
			})
		} else if game.FreeFrom != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Available From",
				Value:  game.FreeFrom,
				Inline: true,
			})
		} else if game.FreeTo != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Available Until",
				Value:  game.FreeTo,
				Inline: true,
			})
		}

		_, err := b.session.ChannelMessageSendEmbed(b.channelID, embed)
		if err != nil {
			return fmt.Errorf("error sending Coming Soon message for %s: %w", game.Title, err)
		}
	}

	log.Printf("Sent %d Coming Soon games to Discord with images", len(games))
	return nil
}

// SendSimpleMessage sends a simple text message to the configured channel
func (b *DiscordBot) SendSimpleMessage(message string) error {
	_, err := b.session.ChannelMessageSend(b.channelID, message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	return nil
}

// SendErrorMessage sends an error message to the configured channel
func (b *DiscordBot) SendErrorMessage(errorMsg string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "Bot Error",
		Description: errorMsg,
		Color:       0xff0000, // Red color
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store - Free Games Bot",
		},
	}

	_, err := b.session.ChannelMessageSendEmbed(b.channelID, embed)
	if err != nil {
		return fmt.Errorf("error sending error message: %w", err)
	}
	return nil
}