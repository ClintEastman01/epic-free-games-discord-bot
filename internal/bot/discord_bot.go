package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"free-games-scrape/internal/config"
	"free-games-scrape/internal/database"
	"free-games-scrape/internal/models"
	"free-games-scrape/internal/service"
)

// DiscordBot handles Discord interactions
type DiscordBot struct {
	session     *discordgo.Session
	config      *config.DiscordConfig
	channelID   string
	gameService *service.GameService
	database    *database.Database
}

// NewDiscordBot creates a new Discord bot instance
func NewDiscordBot(cfg *config.DiscordConfig, gameService *service.GameService, db *database.Database) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &DiscordBot{
		session:     session,
		config:      cfg,
		channelID:   cfg.ChannelID,
		gameService: gameService,
		database:    db,
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
	
	// Register slash commands
	err = b.registerSlashCommands()
	if err != nil {
		log.Printf("Error registering slash commands: %v", err)
		// Don't fail startup, just log the error
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
		b.sendWelcomeMessage(s, g)
	})

	// Add message handler for commands
	b.session.AddHandler(b.messageHandler)
	
	// Add slash command handler
	b.session.AddHandler(b.interactionHandler)
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

// SendGameUpdates sends game updates to all configured Discord channels
func (b *DiscordBot) SendGameUpdates(gameCollection *models.GameCollection) error {
	// Get all active server configurations
	serverConfigs, err := b.database.GetAllActiveServerConfigs()
	if err != nil {
		return fmt.Errorf("error getting server configs: %w", err)
	}

	// If no server configs and we have a legacy channel, use that
	if len(serverConfigs) == 0 && b.channelID != "" {
		if err := b.sendFreeNowGames(gameCollection.FreeNow, b.channelID); err != nil {
			return fmt.Errorf("error sending Free Now games to legacy channel: %w", err)
		}
		if err := b.sendComingSoonGames(gameCollection.ComingSoon, b.channelID); err != nil {
			return fmt.Errorf("error sending Coming Soon games to legacy channel: %w", err)
		}
		return nil
	}

	// Send to all configured channels
	for _, config := range serverConfigs {
		if err := b.sendFreeNowGames(gameCollection.FreeNow, config.ChannelID); err != nil {
			log.Printf("Error sending Free Now games to channel %s: %v", config.ChannelID, err)
			continue
		}
		if err := b.sendComingSoonGames(gameCollection.ComingSoon, config.ChannelID); err != nil {
			log.Printf("Error sending Coming Soon games to channel %s: %v", config.ChannelID, err)
			continue
		}
	}

	return nil
}

// sendFreeNowGames sends "Free Now" games to Discord with images displayed
func (b *DiscordBot) sendFreeNowGames(games []models.Game, channelID string) error {
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

		_, err := b.session.ChannelMessageSendEmbed(channelID, embed)
		if err != nil {
			return fmt.Errorf("error sending Free Now message for %s: %w", game.Title, err)
		}
	}

	log.Printf("Sent %d Free Now games to Discord with images", len(games))
	return nil
}

// sendComingSoonGames sends "Coming Soon" games to Discord with images displayed
func (b *DiscordBot) sendComingSoonGames(games []models.Game, channelID string) error {
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

		_, err := b.session.ChannelMessageSendEmbed(channelID, embed)
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

// registerSlashCommands registers all slash commands with Discord
func (b *DiscordBot) registerSlashCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "setup",
			Description: "Configure which channel to send free game notifications to",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to send notifications to",
					Required:    true,
					ChannelTypes: []discordgo.ChannelType{
						discordgo.ChannelTypeGuildText,
					},
				},
			},
		},
		{
			Name:        "games",
			Description: "Show current free games",
		},
		{
			Name:        "refresh",
			Description: "Manually check for new games",
		},
		{
			Name:        "status",
			Description: "Show bot status and configuration",
		},
		{
			Name:        "help",
			Description: "Show all available commands",
		},
	}

	for _, command := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", command)
		if err != nil {
			return fmt.Errorf("error creating command %s: %w", command.Name, err)
		}
	}

	log.Printf("Successfully registered %d slash commands", len(commands))
	return nil
}

// interactionHandler handles slash command interactions
func (b *DiscordBot) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "setup":
		b.handleSetupCommand(s, i)
	case "games":
		b.handleGamesSlashCommand(s, i)
	case "refresh":
		b.handleRefreshSlashCommand(s, i)
	case "status":
		b.handleStatusCommand(s, i)
	case "help":
		b.handleHelpSlashCommand(s, i)
	}
}

// handleSetupCommand handles the /setup slash command
func (b *DiscordBot) handleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if user has manage channels permission
	permissions, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil {
		b.respondToInteraction(s, i, "Error checking permissions.", true)
		return
	}

	if permissions&discordgo.PermissionManageChannels == 0 {
		b.respondToInteraction(s, i, "You need 'Manage Channels' permission to use this command.", true)
		return
	}

	// Get the channel from the command options
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		b.respondToInteraction(s, i, "Please specify a channel.", true)
		return
	}

	channelID := options[0].ChannelValue(s).ID
	guildID := i.GuildID

	// Save the server configuration
	err = b.database.SaveServerConfig(guildID, channelID)
	if err != nil {
		log.Printf("Error saving server config: %v", err)
		b.respondToInteraction(s, i, "Failed to save configuration. Please try again.", true)
		return
	}

	channelMention := fmt.Sprintf("<#%s>", channelID)
	response := fmt.Sprintf("Successfully configured! I'll send free game notifications to %s", channelMention)
	b.respondToInteraction(s, i, response, false)
	
	log.Printf("Server %s configured to use channel %s", guildID, channelID)
}

// respondToInteraction sends a response to a slash command interaction
func (b *DiscordBot) respondToInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, content string, ephemeral bool) {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}

// handleGamesSlashCommand handles the /games slash command
func (b *DiscordBot) handleGamesSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer the response since getting games might take time
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring interaction response: %v", err)
		return
	}

	games, err := b.gameService.GetActiveGames()
	if err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to get games: %v", err))
		return
	}

	if len(games.FreeNow) == 0 && len(games.ComingSoon) == 0 {
		b.followUpInteraction(s, i, "No free games currently available in the database.")
		return
	}

	// Send games to the current channel
	if err := b.sendFreeNowGames(games.FreeNow, i.ChannelID); err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to send Free Now games: %v", err))
		return
	}
	
	if err := b.sendComingSoonGames(games.ComingSoon, i.ChannelID); err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to send Coming Soon games: %v", err))
		return
	}

	b.followUpInteraction(s, i, "Sent current free games!")
}

// handleRefreshSlashCommand handles the /refresh slash command
func (b *DiscordBot) handleRefreshSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Defer the response since refreshing might take time
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring interaction response: %v", err)
		return
	}

	if err := b.gameService.RefreshGames(); err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to refresh games: %v", err))
		return
	}

	games, err := b.gameService.GetActiveGames()
	if err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to get updated games: %v", err))
		return
	}

	if len(games.FreeNow) == 0 && len(games.ComingSoon) == 0 {
		b.followUpInteraction(s, i, "Games refreshed successfully! No free games found.")
		return
	}

	// Send updated games to the current channel
	if err := b.sendFreeNowGames(games.FreeNow, i.ChannelID); err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to send Free Now games: %v", err))
		return
	}
	
	if err := b.sendComingSoonGames(games.ComingSoon, i.ChannelID); err != nil {
		b.followUpInteraction(s, i, fmt.Sprintf("Failed to send Coming Soon games: %v", err))
		return
	}

	b.followUpInteraction(s, i, "Games refreshed successfully!")
}

// handleStatusCommand handles the /status slash command
func (b *DiscordBot) handleStatusCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	
	// Get server configuration
	serverConfig, err := b.database.GetServerConfig(guildID)
	if err != nil {
		b.respondToInteraction(s, i, "Error checking server configuration.", true)
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: "Bot Status",
		Color: 0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Bot Status",
				Value:  "Online and running",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store - Free Games Bot",
		},
	}

	if serverConfig != nil {
		channelMention := fmt.Sprintf("<#%s>", serverConfig.ChannelID)
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Notification Channel",
			Value:  channelMention,
			Inline: true,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Notification Channel",
			Value:  "Not configured (use /setup)",
			Inline: true,
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Printf("Error responding to status command: %v", err)
	}
}

// handleHelpSlashCommand handles the /help slash command
func (b *DiscordBot) handleHelpSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Free Games Bot Commands",
		Description: "Available slash commands for the Epic Games Free Games Bot:",
		Color:       0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "/setup <channel>",
				Value:  "Configure which channel to send notifications to",
				Inline: false,
			},
			{
				Name:   "/games",
				Value:  "Show current free games",
				Inline: false,
			},
			{
				Name:   "/refresh",
				Value:  "Manually check for new games",
				Inline: false,
			},
			{
				Name:   "/status",
				Value:  "Show bot status and configuration",
				Inline: false,
			},
			{
				Name:   "/help",
				Value:  "Show this help message",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store - Free Games Bot",
		},
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	if err != nil {
		log.Printf("Error responding to help command: %v", err)
	}
}

// followUpInteraction sends a follow-up message to a deferred interaction
func (b *DiscordBot) followUpInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	if err != nil {
		log.Printf("Error sending follow-up message: %v", err)
	}
}

// sendWelcomeMessage sends a welcome message when the bot joins a new guild
func (b *DiscordBot) sendWelcomeMessage(s *discordgo.Session, g *discordgo.GuildCreate) {
	// Find a suitable channel to send the welcome message
	// Try to find a general channel, system channel, or the first text channel we can send to
	var targetChannelID string
	
	// First, try the system channel if it exists
	if g.SystemChannelID != "" {
		targetChannelID = g.SystemChannelID
	} else {
		// Find the first text channel we have permission to send messages to
		for _, channel := range g.Channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				// Check if we can send messages to this channel
				permissions, err := s.UserChannelPermissions(s.State.User.ID, channel.ID)
				if err == nil && permissions&discordgo.PermissionSendMessages != 0 {
					targetChannelID = channel.ID
					break
				}
			}
		}
	}
	
	// If we couldn't find a suitable channel, log and return
	if targetChannelID == "" {
		log.Printf("Could not find a suitable channel to send welcome message in guild %s", g.Name)
		return
	}
	
	// Create the welcome message embed
	embed := &discordgo.MessageEmbed{
		Title:       "Thanks for adding Free Games Bot!",
		Description: "I'll help you stay updated on free games from Epic Games Store.",
		Color:       0x0099ff,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Getting Started",
				Value:  "Use `/setup` to configure which channel I should send notifications to.",
				Inline: false,
			},
			{
				Name:   "Available Commands",
				Value:  "`/games` - Show current free games\n`/refresh` - Manually check for new games\n`/status` - Show bot status\n`/help` - Show all commands",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store - Free Games Bot",
		},
	}
	
	// Send the welcome message
	_, err := s.ChannelMessageSendEmbed(targetChannelID, embed)
	if err != nil {
		log.Printf("Error sending welcome message to guild %s: %v", g.Name, err)
	} else {
		log.Printf("Sent welcome message to guild %s in channel %s", g.Name, targetChannelID)
	}
}