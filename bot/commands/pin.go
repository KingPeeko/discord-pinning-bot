package commands

import (
	"context"
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
var messageLinkRegex = regexp.MustCompile(`^https?:\/\/(discord(app)?\.com)\/channels\/\d+\/\d+\/\d+$`)

// Helper function to insert pins into the database
func InsertPin(conn *pgx.Conn, link string, guild string, description string) (int64, error) {
	commandTag, err := conn.Exec(context.Background(),
		"INSERT INTO pins (link, guild, description) VALUES ($1, $2, $3) ON CONFLICT (link) DO NOTHING",
		link, guild, description)
	return commandTag.RowsAffected(), err
}

// Command handler
func PinHandler(conn *pgx.Conn) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

		// Insert entry into database
		affectedRows, err := InsertPin(conn, messageLink, guildID, description)
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
				Content: fmt.Sprintf("✅ Message pinned successfully!\n📌 **Link:** %s\n📝 **Description:** %s", messageLink, description),
			},
		})
	}
}
