// Cryptocurrency Guide Bot (Crypto Assistant)
// Essence: The bot provides up-to-date information on the rates of top
// cryptocurrencies (BTC, ETH, SOL), sets alerts on price thresholds
// The mine package provides the basic functionality of the bot
package main

import (
	"crypto-bot/internal/bot"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
	// Initialize the database
	db := bot.InitDB()
	defer db.Close()

	token := os.Getenv("TELEGRAM_APITOKEN")
	if token == "" {
		log.Panic("Переменная окружения TELEGRAM_APITOKEN не задана")
	}

	//Create a bot instance using the token
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// Turn on debug mode
	botAPI.Debug = true

	// We display information that the bot was launched on behalf of its username
	log.Printf("Авторизован как %s", botAPI.Self.UserName)

	// We set up a channel from which we will receive updates (messages from users)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// GetUpdatesChan returns the channel from which we will read incoming messages.
	updates := botAPI.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		// Create a message for the reply
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		chatID := update.Message.Chat.ID // Save the chat ID

		// Process the command
		switch update.Message.Command() {
		case "start":
			msg.Text = "Привет! Я твой крипто-бот 🤖\nЯ могу показать актуальные цены на криптовалюты.\n\nКоманды:\n/price [тикер] - Узнать цену\n/alert [тикер] [цена] [above/below] - Установить оповещение\nНапример: /alert BTC 50000 above"
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
		case "alert":
			// Processing the command to set the alert
			// Format: /alert BTC 50000 above
			args := update.Message.CommandArguments()
			parts := strings.Split(args, " ")
			if len(parts) < 3 {
				msg.Text = "Неверный формат. Используй: /alert [тикер] [цена] [above/below]\nНапример: /alert BTC 50000 above"
				break
			}
			coin := strings.ToUpper(parts[0])
			threshold, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				msg.Text = "Не могу распознать цену. Убедись, что это число."
				break
			}
			isAbove := true
			if strings.ToLower(parts[2]) == "below" {
				isAbove = false
			} else if strings.ToLower(parts[2]) != "above" {
				msg.Text = "Последний аргумент должен быть 'above' или 'below'."
				break
			}
			// Create a notification structure and save it in the database
			alert := bot.UserAlert{
				ChatID:    chatID,
				Coin:      coin,
				Threshold: threshold,
				IsAbove:   isAbove,
			}
			err = bot.SaveAlert(db, alert)
			if err != nil {
				log.Printf("Ошибка сохранения оповещения: %v", err)
				msg.Text = "Произошла ошибка при сохранении. Попробуй ещё раз."
			} else {
				directionText := "выше"
				if !isAbove {
					directionText = "ниже"
				}
				msg.Text = fmt.Sprintf("✅ Оповещение установлено! Я сообщу, когда цена %s станет %s $ %.2f.", coin, directionText, threshold)
			}
		default:
			msg.Text = "Я не знаю такой команды. Попробуй /start"
		}

		// Sending a message
		if _, err := botAPI.Send(msg); err != nil {
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
