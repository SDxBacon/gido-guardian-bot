package main

import (
	"log"
	"os"

	"github.com/SDxBacon/gido-guardian-bot/bot"
	"github.com/joho/godotenv"
)

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// get the bot token from the environment
	token := os.Getenv("BOT_TOKEN")

	bot.Token = token
	bot.Run()
}
