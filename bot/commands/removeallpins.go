package commands

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var RemoveAllPinsCommand = &discordgo.ApplicationCommand{
	Name:        "removeallpins",
	Description: "Remove all pins saved on the server, only available to admin! Be careful :)",
}

func RemoveAllPinsHandler(conn *pgx.Conn) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		_, err := conn.Exec(context.Background(), "DELETE FROM pins WHERE guild = $1", i.GuildID)
		if err != nil {
			log.Println("Error while trying to delete all messages: ", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Could not delete all pins",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "All pinned messages have been deleted from the database",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	}
}
