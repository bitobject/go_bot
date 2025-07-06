package main

import (
	"log"
	"os"
	"goooo/internal/config"
	"goooo/internal/database"
	"goooo/internal/bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	config.LoadEnv()
	db := database.Init()
	defer database.Close(db)

	botToken := os.Getenv("TELEGRAM_TOKEN")
	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Start(tgBot, db)
} 