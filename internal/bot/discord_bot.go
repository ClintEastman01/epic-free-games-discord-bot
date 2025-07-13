package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"free-games-scrape/internal/config"
	"free-games-scrape/internal/models"
)

// DiscordBot handles Discord interactions
type DiscordBot struct {
	session   *discordgo.Session
	config    *config.DiscordConfig
	channelID string
}

// NewDiscordBot creates a new Discord bot instance
func NewDiscordBot(cfg *config.DiscordConfig) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	bot := &DiscordBot{
		session:   session,
		config:    cfg,
		channelID: cfg.ChannelID,
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

// sendFreeNowGames sends "Free Now" games to Discord
func (b *DiscordBot) sendFreeNowGames(games []models.Game) error {
	if len(games) == 0 {
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎮 Free Games Available Now!",
		Description: "These games are currently free on Epic Games Store:",
		Color:       0x00ff00, // Green color
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store • Free Games Bot",
		},
	}

	for i, game := range games {
		fieldValue := fmt.Sprintf("**Status:** %s\n", game.Status)
		if game.FreeTo != "" {
			fieldValue += fmt.Sprintf("**Free Until:** %s\n", game.FreeTo)
		}
		if game.ImageURL != "" {
			fieldValue += fmt.Sprintf("[View Game Image](%s)", game.ImageURL)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", i+1, game.Title),
			Value:  fieldValue,
			Inline: false,
		})
	}

	_, err := b.session.ChannelMessageSendEmbed(b.channelID, embed)
	if err != nil {
		return fmt.Errorf("error sending Free Now message: %w", err)
	}

	log.Printf("Sent %d Free Now games to Discord", len(games))
	return nil
}

// sendComingSoonGames sends "Coming Soon" games to Discord
func (b *DiscordBot) sendComingSoonGames(games []models.Game) error {
	if len(games) == 0 {
		return nil
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📅 Free Games Coming Soon!",
		Description: "These games will be free soon on Epic Games Store:",
		Color:       0x0099ff, // Blue color
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store • Free Games Bot",
		},
	}

	for i, game := range games {
		fieldValue := fmt.Sprintf("**Status:** %s\n", game.Status)
		if game.FreeFrom != "" && game.FreeTo != "" {
			fieldValue += fmt.Sprintf("**Free Period:** %s - %s\n", game.FreeFrom, game.FreeTo)
		}
		if game.ImageURL != "" {
			fieldValue += fmt.Sprintf("[View Game Image](%s)", game.ImageURL)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", i+1, game.Title),
			Value:  fieldValue,
			Inline: false,
		})
	}

	_, err := b.session.ChannelMessageSendEmbed(b.channelID, embed)
	if err != nil {
		return fmt.Errorf("error sending Coming Soon message: %w", err)
	}

	log.Printf("Sent %d Coming Soon games to Discord", len(games))
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
		Title:       "⚠️ Bot Error",
		Description: errorMsg,
		Color:       0xff0000, // Red color
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Epic Games Store • Free Games Bot",
		},
	}

	_, err := b.session.ChannelMessageSendEmbed(b.channelID, embed)
	if err != nil {
		return fmt.Errorf("error sending error message: %w", err)
	}
	return nil
}