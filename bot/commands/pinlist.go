package commands

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var AllPinsCommand = &discordgo.ApplicationCommand{
	Name:        "allpins",
	Description: "Lists all pins (list is only visible to you)",
}

// Regex to extract channel Id and message Id
var messageLinkRegexAllLinks = regexp.MustCompile(`https://(?:discord.com|discordapp.com)/channels/\d+/(\d+)/(\d+)`)

func GetAllPins(conn *pgx.Conn, guildID string) ([][2]string, error) {
	rows, err := conn.Query(
		context.Background(),
		"SELECT link, description FROM pins WHERE guild = $1",
		guildID,
	)
	if err != nil {
		log.Println("Error gettings pins from database:", err)
		return nil, err
	}
	defer rows.Close()

	var pins [][2]string

	for rows.Next() {
		var link, description string
		err := rows.Scan(&link, &description)
		if err != nil {
			log.Println("Row scanning error:", err)
			return nil, err
		}
		pins = append(pins, [2]string{link, description})
	}
	return pins, nil
}

func AllPinsHandler(conn *pgx.Conn) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guildID := i.GuildID

		// Acknowledge interaction to prevent timeout
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		// Get pins from database
		pins, err := GetAllPins(conn, guildID)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Failed to retrieve pins due to internal error.",
			})
			return
		}

		if len(pins) == 0 {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "No pins found for this server!",
			})
			return
		}

		// Iterate through stored pins and create embeds
		for _, pin := range pins {
			link := pin[0]
			description := pin[1]

			matches := messageLinkRegexAllLinks.FindStringSubmatch(link)
			if matches == nil {
				log.Println("Invalid pin format:", link)
				continue
			}

			channelID := matches[1]
			messageID := matches[2]

			// Fetch the message from Discord
			msg, err := s.ChannelMessage(channelID, messageID)
			if err != nil {
				log.Println("Failed to fetch message:", err)
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: fmt.Sprintf("Failed to retrieve pinned message: %s", link),
				})
				continue
			}

			// Create an embed with the message details
			embed := &discordgo.MessageEmbed{
				Title:       description,
				Description: msg.Content,
				URL:         link,
				Color:       0x5865F2,
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Pinned by %s", msg.Author.Username),
				},
				Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z"),
			}

			// Attach author's avatar
			if msg.Author != nil {
				embed.Author = &discordgo.MessageEmbedAuthor{
					Name:    msg.Author.Username,
					IconURL: msg.Author.AvatarURL(""),
				}
			}

			// Include image if message contains one
			if len(msg.Attachments) > 0 {
				embed.Image = &discordgo.MessageEmbedImage{
					URL: msg.Attachments[0].URL,
				}
			}

			// Send the embed
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{embed},
			})
		}
	}
}
