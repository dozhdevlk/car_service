// telegramBot.go
package main

import (
	"database/sql"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

func InitTelegramBot() {
	var err error
	bot, err = tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Ошибка инициализации Telegram-бота: %v", err)
	}

	bot.Debug = true
	log.Printf("Telegram bot authorized on account %s", bot.Self.UserName)
}

func StartTelegramBot(db *sql.DB) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		if update.Message.Command() == "start" {
			token := update.Message.CommandArguments()
			var userID int
			err := db.QueryRow("SELECT id FROM users WHERE telegram_token = $1", token).Scan(&userID)
			if err != nil {
				log.Println("Недействительный токен:", err)
				continue
			}

			_, err = db.Exec("UPDATE users SET telegram_chat_id = $1, telegram_token = NULL WHERE id = $2", update.Message.Chat.ID, userID)
			if err != nil {
				log.Println("Ошибка при сохранении chat_id:", err)
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы успешно подключили Telegram!")
			bot.Send(msg)
		}
	}
}

func SendTelegramNotification(db *sql.DB, userID int, message string) {
	var chatID int64
	err := db.QueryRow("SELECT telegram_chat_id FROM users WHERE id = $1", userID).Scan(&chatID)
	if err != nil || chatID == 0 {
		log.Println("Не удалось получить Telegram chat ID:", err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, message)
	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка при отправке Telegram-сообщения:", err)
	}
}
