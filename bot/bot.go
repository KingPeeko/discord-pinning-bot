package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var BotToken string
var GuildID string

func Run(dbConn *pgx.Conn) {
	conn = dbConn

	fmt.Println("Running with token: ", BotToken)
	// Create new Discord Session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal(err)
	}

	// Handler for all commands
	discord.AddHandler(
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		},
	)

	// Open session
	err = discord.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Register commands
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer discord.Close()

	// Run until code is terminated
	fmt.Println("Bot is running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
