package bot

import (
	"github.com/alfredosa/GoDiscordBot/config"

	"github.com/bwmarrin/discordgo"
)

var botId string
var botSession *discordgo.Session

func startBot() {
	botSession, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		return
	}

}
