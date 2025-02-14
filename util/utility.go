package util

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
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
