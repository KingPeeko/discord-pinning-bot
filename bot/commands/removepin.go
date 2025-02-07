package commands

import (
	"context"
	"discord-pinning-bot/util"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var RemovePinCommand = &discordgo.ApplicationCommand{
	Name:        "removepin",
	Description: "Remove your own pin (Or any pin, if you are an admin)",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "message-link",
			Description: "A link to the message you want to remove the pin from.",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

func deletePin(conn *pgx.Conn, link string) error {
	_, err := conn.Exec(context.Background(),
		"DELETE FROM pins WHERE link = $1",
		link)
	return err
}

func RemovePinHandler(conn *pgx.Conn) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var messageLink string

		for _, option := range i.ApplicationCommandData().Options {
			if option.Name == "message-link" {
				messageLink = option.StringValue()
			}
		}

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

		var pinnerID string
		err := conn.QueryRow(
			context.Background(),
			"SELECT pinner FROM pins WHERE link = $1",
			messageLink,
		).Scan(&pinnerID)
		if err != nil {
			if err == pgx.ErrNoRows {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ This message is not pinned.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			log.Println("Error querying pin:", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Internal Error",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Check if user is the original pinner or an admin
		userID := i.Member.User.ID
		if pinnerID != userID && !util.HasAdminPermissions(i) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Only the original pinner and admins may remove pins.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		if err := deletePin(conn, messageLink); err != nil {
			log.Println("Error deleting pin:", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Failed to delete pin due to an internal error.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("✅ Successfully deleted pin: %s", messageLink),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
