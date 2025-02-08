package commands

import (
	"context"
	"discord-pinning-bot/util"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var RandomPinCommand = &discordgo.ApplicationCommand{
	Name:        "randompin",
	Description: "Shows you a random pin!",
}

func selectRandomPin(conn *pgx.Conn, guildID string) ([3]string, error) {
	var link, description, pinner string

	err := conn.QueryRow(
		context.Background(),
		"SELECT link, description, pinner, date FROM pins WHERE guild = $1",
		guildID,
	).Scan(&link, &description, &pinner)
	if err != nil {
		return [3]string{"", "", ""}, err
	}

	return [3]string{link, description, pinner}, err
}

func RandomPinHandler(conn *pgx.Conn) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		// Get guildID
		guildID := i.GuildID

		// Get a random pin
		pin, err := selectRandomPin(conn, guildID)
		if err != nil {
			if err == pgx.ErrNoRows {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "There are no pins",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			log.Println("Error in scanning for one random pin: ", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Internal Error",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		link := pin[0]
		description := pin[1]
		pinner := pin[2]

		matches := messageLinkRegexAllLinks.FindStringSubmatch(link)
		if matches == nil {
			log.Println("Invalid pin format:", pin)
			return
		}

		channelID := matches[1]
		messageID := matches[2]

		msg, err := s.ChannelMessage(channelID, messageID)
		if err != nil {
			log.Println("Failed to fetch message:", err)
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Failed to retrieve pinned message: %s", pin),
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Random Pinned Message!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		embed := util.CreateEmbed(msg, link, description, pinner)
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{embed},
		})
		if err != nil {
			log.Println("Sending embeds failed due to error: ", err)
		}
	}
}
