package bot

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var conn *pgx.Conn

// Helper function to insert pins into database
func insertPin(conn *pgx.Conn, link string, guild string, description string) error {
	_, err := conn.Exec(context.Background(),
		"INSERT INTO pins (link, guild, description) VALUES ($1, $2, $3) ON CONFLICT (link) DO NOTHING",
		link, guild, description)
	return err
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "say-hello",
			Description: "Makes the bot respond with 'hello!'",
		},
		{
			Name:        "pin",
			Description: "Pin a message by following this command with a message link",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message-link",
					Description: "A link to the message you want to pin. You can get it with right-click, and then 'Copy Message Link'",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:        "description",
					Description: "Give that message a description!",
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"say-hello": func(session *discordgo.Session, i *discordgo.InteractionCreate) {
			session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "hello!",
				},
			})
		},
		"pin": func(session *discordgo.Session, i *discordgo.InteractionCreate) {
			var messageLink string
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "message-link" {
					messageLink = option.StringValue()
					break
				}
			}
			var description = ""
			for _, option := range i.ApplicationCommandData().Options {
				if option.Name == "description" {
					description = option.StringValue()
					break
				}
			}

			guildID := i.GuildID

			err := insertPin(conn, messageLink, guildID, description)
			if err != nil {
				log.Println("Database insert error: ", err)
				session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Failed to pin message due to database error.",
					},
				})
				return
			}

			session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Message %s Pinned!", messageLink),
				},
			})
		},
	}
)
