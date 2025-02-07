package bot

import (
	"discord-pinning-bot/bot/commands"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

// Global slice of commands
var AllCommands = []*discordgo.ApplicationCommand{
	commands.PinCommand,
	commands.AllPinsCommand,
	commands.RemovePinCommand,
}

// Command handler mapping
func GetCommandHandlers(conn *pgx.Conn) map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"pin":       commands.PinHandler(conn),
		"allpins":   commands.AllPinsHandler(conn),
		"removepin": commands.RemovePinHandler(conn),
	}
}
