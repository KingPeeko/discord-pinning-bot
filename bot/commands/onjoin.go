package commands

import (
	"context"
	"log"
	"time"

	"discord-pinning-bot/util"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

func JoinHandler(conn *pgx.Conn, s *discordgo.Session, g *discordgo.GuildCreate) {
	var guild string

	// Check if there is a row for guild already
	err := conn.QueryRow(context.Background(), "SELECT guild FROM guilds WHERE guild = $1", g.ID).Scan(&guild)
	if err != pgx.ErrNoRows {
		return
	}

	time.Sleep(3000 * time.Millisecond)
	_, err = util.CheckAndCreatePinChannel(conn, s, g.Guild)
	if err != nil {
		sendDMToInviter(s, g.Guild.ID, "❌ Unfortunately there seems to have been an error while adding the bot to your server.\n\nInvite the bot again and be sure to give it necessary authorization for proper function!")
	}

	sendDMToInviter(s, g.Guild.ID, "Thank you for inviting me to your server!\n\nI have created a channel for pins called 'pinned-messages' where new pins will appear. \nYou can pin messages by pinning them on the server, or by using /pin\n\nEnjoy!")
}

func sendDMToInviter(s *discordgo.Session, guildID string, message string) {
	// Get latest log on any bot invited
	auditLogs, err := s.GuildAuditLog(guildID, "", "", 28, 1)
	if err != nil || len(auditLogs.AuditLogEntries) == 0 {
		log.Println("Could not retrieve audit log or no entries found:", err)
		return
	}

	// Get the user who invited the bot
	inviterID := auditLogs.AuditLogEntries[0].UserID

	// Send message
	dmChannel, err := s.UserChannelCreate(inviterID)
	if err != nil {
		log.Println("Could not create DM channel:", err)
		return
	}
	_, err = s.ChannelMessageSend(dmChannel.ID, message)
	if err != nil {
		log.Println("Could not send DM:", err)
	}
}
