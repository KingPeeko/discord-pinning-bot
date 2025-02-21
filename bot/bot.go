package bot

import (
	"discord-pinning-bot/bot/commands"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var BotToken string
var GuildID string

func Run(conn *pgx.Conn) {

	fmt.Println("Running with token: ", BotToken)
	// Create new Discord Session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal(err)
	}

	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentMessageContent

	// Handler for all commands
	commandHandlers := GetCommandHandlers(conn)
	discord.AddHandler(
		func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		},
	)
	discord.AddHandler(
		func(s *discordgo.Session, g *discordgo.GuildCreate) {
			fmt.Println("GuildCreate event !")
			commands.JoinHandler(conn, s, g)
		},
	)

	// Open session
	err = discord.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Register commands
	registeredCommands := make([]*discordgo.ApplicationCommand, len(AllCommands))
	for i, v := range AllCommands {
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
