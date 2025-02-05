package bot

import "github.com/bwmarrin/discordgo"

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "say-hello",
			Description: "Makes the bot respond with 'hello!'",
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
	}
)
