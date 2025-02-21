package util

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

func HasAdminPermissions(i *discordgo.InteractionCreate) bool {
	return i.Member.Permissions&discordgo.PermissionAdministrator != 0
}

func NewInteractionRespond(s *discordgo.Session, i *discordgo.Interaction, content string, invisible bool) {
	if invisible {
		s.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
	} else {
		s.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	}
}

func CreateEmbed(s *discordgo.Session, msg *discordgo.Message, link string, description string, pinner string) (*discordgo.MessageEmbed, error) {
	// Get username of pinner with their ID
	pinnerMember, err := s.User(pinner)
	if err != nil {
		log.Println("Could not find find user in guild, using ID: ", pinner)
		return nil, err
	}
	pinnerUsername := pinnerMember.Username

	embed := &discordgo.MessageEmbed{
		Title:       description,
		Description: fmt.Sprintf("```%s```", msg.Content),
		URL:         link,
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Pinned by %s", pinnerUsername),
		},
		Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z"),
		/*Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: msg.Author.AvatarURL(""),
		},
		*/
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

	return embed, nil
}

func CheckAndCreatePinChannel(conn *pgx.Conn, s *discordgo.Session, g *discordgo.Guild) (string, error) {
	var pinChannelID string

	// Check if there is a row for pin channel id already
	err := conn.QueryRow(context.Background(), "SELECT pin_channel FROM guilds WHERE guild = $1", g.ID).Scan(&pinChannelID)
	if err == pgx.ErrNoRows {
		// if no row, make new channel and add guild to database
		channel, err := s.GuildChannelCreate(g.ID, "pinned-messages", discordgo.ChannelTypeGuildText)
		if err != nil {
			log.Println("Could not create pin channel: ", err)
			return "", err
		}

		_, err = conn.Exec(context.Background(), "INSERT INTO guilds (guild, pin_channel, pin_count) VALUES ($1, $2, $3)", g.ID, channel.ID, 0)
		if err != nil {
			log.Println("Error inserting new pin channel into DB:", err)
			return "", err
		}

		time.Sleep(2000 * time.Millisecond)

		return channel.ID, nil
	}

	if err != nil {
		log.Println("There was a database error: ", err)
		return "", err
	}

	channel, err := s.Channel(pinChannelID)
	if err == nil && channel.Type == discordgo.ChannelTypeGuildText {
		return pinChannelID, nil
	}

	newChannel, err := s.GuildChannelCreate(g.ID, "pinned-messages", discordgo.ChannelTypeGuildText)
	if err != nil {
		log.Println("Could not create pin channel: ", err)
		return "", err
	}

	time.Sleep(2000 * time.Millisecond)

	_, err = conn.Exec(context.Background(), "UPDATE guilds SET pin_channel = $1 WHERE guild = $2", newChannel.ID, g.ID)
	if err != nil {
		log.Println("Error updating pin channel in DB:", err)
		return "", err
	}
	return pinChannelID, nil
}
