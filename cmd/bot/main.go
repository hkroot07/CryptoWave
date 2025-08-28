// Cryptocurrency Guide Bot (Crypto Assistant)
// Essence: The bot provides up-to-date information on the rates of top
// cryptocurrencies (BTC, ETH, SOL), sets alerts on price thresholds
// The mine package provides the basic functionality of the bot
package main

import (
	"log"
	"os"

	// Importing a library to work with Telegram Bot API
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Function main
// Entry point to the program
func main() {
	token := os.Getenv("TELEGRAM_APITOKEN")
	if token == "" {
		log.Panic("Переменная окружения TELEGRAM_APITOKEN не задана")
	}

	//Create a bot instance using the token
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// Turn on debug mode
	bot.Debug = true

	// We display information that the bot was launched on behalf of its username
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// We set up a channel from which we will receive updates (messages from users)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// GetUpdatesChan returns the channel from which we will read incoming messages.
	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			continue
		}

		// We log who wrote what
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// We are preparing a response message.
		// NewMessage() creates a new message.
		// The first argument is the chat ID where to send the message.
		// The second argument is the message text.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я твой первый бот на Go! Ты написал: "+update.Message.Text)

		// Sending a message
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
