// Cryptocurrency Guide Bot (Crypto Assistant)
// Essence: The bot provides up-to-date information on the rates of top
// cryptocurrencies (BTC, ETH, SOL), sets alerts on price thresholds
// The mine package provides the basic functionality of the bot
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	// Importing a library to work with Telegram Bot API
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Structure for parsing the response from CoinGecko API
// JSON from the API will be automatically converted to this structure
type CoinGeckoResponse struct {
	Bitcoin struct {
		Usd float64 `json:"usd"`
	} `json:"bitcoin"`
}

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

		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		// Create a message for the reply
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		// Process the command
		switch update.Message.Command() {
		case "start":
			msg.Text = "Привет! Я твой крипто-бот 🤖\nЯ могу показать актуальные цены на криптовалюты.\nНапиши /price BTC"
		case "price", "p":
			// Get command arguments (e.g. "BTC" from "/price BTC")
			args := update.Message.CommandArguments()
			coin := strings.ToUpper(args)
			if coin == "" {
				coin = "BTC" // Default value
			}
			// Get price from API
			price, err := getCryptoPrice(coin)
			if err != nil {
				log.Printf("Ошибка получения цены: %v", err)
				msg.Text = "Извини, не могу получить данные 😕 Попробуй позже."
			} else {
				msg.Text = fmt.Sprintf("💰 %s: $%.2f", coin, price)
			}
		default:
			msg.Text = "Я не знаю такой команды. Попробуй /start"
		}
		// Sending a message
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

// Function to get the price of a cryptocurrency
func getCryptoPrice(coinSymbol string) (float64, error) {
	if coinSymbol != "BTC" {
		return 0, fmt.Errorf("пока поддерживается только BTC")
	}
	// Forming a URL request to the API
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"
	// Perform an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	// Parse JSON into our structure
	var data CoinGeckoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}
	// Return the price
	return data.Bitcoin.Usd, nil
}
