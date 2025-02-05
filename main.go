package main

import (
	"discord-pinning-bot/bot"
	"log"
	"os"
)

func main() {
	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatal("Must set Discord token as env variabler: BOT_TOKEN")
	}

	bot.BotToken = botToken
	bot.Run()
}
