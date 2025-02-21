package commands

import (
	"context"
	"discord-pinning-bot/util"
	"fmt"
	"log"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var PinCommand = &discordgo.ApplicationCommand{
	Name:        "pin",
	Description: "Pin a message by following this command with a message link",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "message-link",
			Description: "A link to the message you want to pin.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "description",
			Description: "Give that message a description!",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	},
}

// Regex to check if link is a discord message link

// Helper function to insert pins into the database
func InsertPin(conn *pgx.Conn, link string, guild string, description string, pinner string, date string) (int64, error) {
	commandTag, err := conn.Exec(context.Background(),
		"INSERT INTO pins (link, guild, description, pinner, date) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (link) DO NOTHING",
		link, guild, description, pinner, date)
	return commandTag.RowsAffected(), err
}

// Command handler
func PinHandler(conn *pgx.Conn) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		messageLinkRegex = regexp.MustCompile(`^https?:\/\/discord(?:app)?\.com\/channels\/\d+\/(\d+)\/(\d+)$`)

		var messageLink, description string

		for _, option := range i.ApplicationCommandData().Options {
			if option.Name == "message-link" {
				messageLink = option.StringValue()
			} else if option.Name == "description" {
				description = option.StringValue()
			}
		}

		// Check if message is a valid discord message link
		if !messageLinkRegex.MatchString(messageLink) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Invalid message link. Please provide a valid Discord message link.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Get the GuildID(Server ID)
		guildID := i.GuildID

		// Get pinner
		pinner := i.Member.User.ID

		// Get date of message
		matches := messageLinkRegex.FindStringSubmatch(messageLink)
		channelID := matches[1]
		messageID := matches[2]

		// Fetch the message from Discord
		msg, err := s.ChannelMessage(channelID, messageID)
		if err != nil {
			log.Println("Failed to fetch message:", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Invalid message link. Please provide a valid Discord message link.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		date := msg.Timestamp.Format("2006-01-02 15:04:05")

		// Insert entry into database
		affectedRows, err := InsertPin(conn, messageLink, guildID, description, pinner, date)
		if err != nil {
			log.Println("Database insert error:", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to pin message due to database error.",
				},
			})
			return
		}

		// Check if message is already pinned by seeing if affected rows are 0
		if affectedRows == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🔹 This message has already been pinned!",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Respond if success
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("✅ Message pinned successfully! -> [%s](%s)", description, messageLink),
			},
		})

		err = sendPinToPinChannel(s, guildID, conn, msg, messageLink, description, pinner)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: ("❌ Failed to send pinned message to pin channel, but message was pinned"),
				Flags:   discordgo.MessageFlagsEphemeral,
			})
		}
	}
}

func sendPinToPinChannel(s *discordgo.Session, guildID string, conn *pgx.Conn, msg *discordgo.Message, messageLink string, description string, pinner string) error {
	// Get guild from guildID
	guild, err := s.Guild(guildID)
	if err != nil {
		log.Println("Could not retrieve guild from guildID", err)
		return err
	}

	// Get pin channel
	guildPinChannel, err := util.CheckAndCreatePinChannel(conn, s, guild)
	if err != nil {
		log.Println("Could not retrieve pin channel from guild", err)
		return err
	}

	// Create embed for pinned message
	embed, err := util.CreateEmbed(s, msg, messageLink, description, pinner)
	if err != nil {
		log.Println("Could not create embed from message", err)
		return err
	}
	s.ChannelMessageSendEmbed(guildPinChannel, embed)
	return nil
}
