// internal/bot/notifier.go
package bot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CheckAndNotify recently all alerts and sending notifications when organizing
func CheckAndNotify(db *sql.DB, botAPI *tgbotapi.BotAPI) {
	// Get ALL notifications from the database
	alerts, err := GetAllAlerts(db)
	if err != nil {
		log.Printf("Ошибка получения оповещений: %v", err)
		return
	}
	// For each notification we check the current price
	for _, alert := range alerts {
		currentPrice, err := GetCryptoPrice(alert.Coin)
		if err != nil {
			log.Printf("Ошибка получения цены для %s: %v", alert.Coin, err)
			continue
		}
		// Check the trigger conditions
		shouldNotify := false
		var message string
		if alert.IsAbove && currentPrice >= alert.Threshold {
			shouldNotify = true
			message = fmt.Sprintf("🚀 Цена %s достигла $%.2f (текущая: $%.2f)",
				alert.Coin, alert.Threshold, currentPrice)
		} else if !alert.IsAbove && currentPrice <= alert.Threshold {
			shouldNotify = true
			message = fmt.Sprintf("🔻 Цена %s упала до $%.2f (текущая: $%.2f)",
				alert.Coin, alert.Threshold, currentPrice)
		}
		// If the condition is met, we send a message
		if shouldNotify {
			msg := tgbotapi.NewMessage(alert.ChatID, message)
			if _, err := botAPI.Send(msg); err != nil {
				log.Printf("Ошибка отправки оповещения: %v", err)
			} else {
				log.Printf("Отправлено оповещение для %s chat_id %d", alert.Coin, alert.ChatID)
				// REMOVE the notification after sending (to avoid spam)
				err = deleteAlert(db, alert.ChatID, alert.Coin)
				if err != nil {
					log.Printf("Ошибка удаления оповещения: %v", err)
				}
			}
		}
	}
}

// deleteAlert deletes an alert from the database
func deleteAlert(db *sql.DB, chatID int64, coin string) error {
	query := `DELETE FROM user_alerts WHERE chat_id = ? AND coin = ?`
	_, err := db.Exec(query, chatID, coin)
	if err != nil {
		log.Printf("Ошибка удаления оповещения: %v", err)
		return err
	}
	log.Printf("Оповещение удалено: chat_id=%d, coin=%s", chatID, coin)
	return nil
}
