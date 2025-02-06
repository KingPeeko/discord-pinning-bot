package main

import (
	"context"
	"discord-pinning-bot/bot"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load in .env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Lookup variables, exit if not there
	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatal("Must set Discord token as env variable: BOT_TOKEN")
	}
	databaseString, ok := os.LookupEnv("DATABASE_STRING")
	if !ok {
		log.Fatal("Must set Database Connection String as env variable: DATABASE_STRING")
	}
	testGuildID, ok := os.LookupEnv("TEST_GUILD_ID")
	if !ok {
		testGuildID = ""
	}

	// Connect to database
	conn, err := pgx.Connect(context.Background(), databaseString)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	defer conn.Close(context.Background())

	bot.BotToken = botToken
	bot.GuildID = testGuildID
	bot.Run(conn)
}
