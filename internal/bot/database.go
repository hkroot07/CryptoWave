// internal/bot/database.go
package bot

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// UserAlert represents a record of a user's subscription to an alert
type UserAlert struct {
	ChatID    int64   //Chat ID with user
	Coin      string  // Coin ticker (e.g. "BTC")
	Threshold float64 // Price threshold at which notification is required
	IsAbove   bool    // true - if you want to notify when the price is ABOVE the threshold, false - when it is BELOW
}

// InitDB создаёт и инициализирует базу данных, создаёт таблицу, если её нет
func InitDB() *sql.DB {
	// Open the database file. If it does not exist, it will be created.
	db, err := sql.Open("sqlite", "./bot_data.db")
	if err != nil {
		log.Panic("Ошибка открытия базы данных:", err)
	}
	// Create a table to store notifications
	query := `
	CREATE TABLE IF NOT EXISTS user_alerts (
		chat_id INTEGER NOT NULL,
		coin TEXT NOT NULL,
		threshold REAL NOT NULL,
		is_above INTEGER NOT NULL, -- В SQLite нет boolean, используем INTEGER (0/1)
		PRIMARY KEY (chat_id, coin)
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		log.Panic("Ошибка создания таблицы:", err)
	}

	log.Println("База данных успешно инициализирована")
	return db
}

// SaveAlert saves or updates the user's alert setting
func SaveAlert(db *sql.DB, alert UserAlert) error {
	// Query to insert or replace data (using UPSERT)
	query := `
	INSERT OR REPLACE INTO user_alerts (chat_id, coin, threshold, is_above)
	VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(query, alert.ChatID, alert.Coin, alert.Threshold, alert.IsAbove)
	return err
}

// GetUserAlerts returns all alerts for a specific user
func GetUserAlerts(db *sql.DB, chatID int64) ([]UserAlert, error) {
	query := `
	SELECT chat_id, coin, threshold, is_above FROM user_alerts
	WHERE chat_id = ?
	`
	rows, err := db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []UserAlert
	for rows.Next() {
		var a UserAlert
		// Читаем данные из строки и заполняем структуру UserAlert
		err := rows.Scan(&a.ChatID, &a.Coin, &a.Threshold, &a.IsAbove)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

// GetAllAlerts возвращает ВСЕ оповещения из базы. Нужно будет для фоновой проверки цен.
func GetAllAlerts(db *sql.DB) ([]UserAlert, error) {
	query := `SELECT chat_id, coin, threshold, is_above FROM user_alerts`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []UserAlert
	for rows.Next() {
		var a UserAlert
		err := rows.Scan(&a.ChatID, &a.Coin, &a.Threshold, &a.IsAbove)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}
