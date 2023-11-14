package model

import (
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mykyta-kravchenko98/telegram-parser/internal/service"
)

// TelegramBot структура для работы с Telegram API
type TelegramBot struct {
	bot              *tgbotapi.BotAPI
	googleSheet      *service.GoogleSheetsService
	keywords         []string
	authorizedUserID int64
}

// NewTelegramBot создает новый экземпляр TelegramBot
func NewTelegramBot(token string, googleSheet *service.GoogleSheetsService, keywords []string, authorizedUserID int64) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true
	return &TelegramBot{bot: bot, googleSheet: googleSheet, keywords: keywords, authorizedUserID: authorizedUserID}, nil
}

func (tb *TelegramBot) SetKeywords(keywords []string) {
	tb.keywords = keywords
}

func (tb *TelegramBot) containsKeywords(text string) bool {
	for _, keyword := range tb.keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// ListenAndRespond начинает слушать обновления и отвечать на сообщения
func (tb *TelegramBot) Listen() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 5

	updates := tb.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Проверка на команду
		if update.Message.IsCommand() {
			// Обработка команды
			tb.handleCommand(update)
			continue
		}

		// Обработка обычных сообщений
		if tb.containsKeywords(update.Message.Text) {
			sheetName := time.Now().Format("2006-01-02")
			tb.googleSheet.CreateSheetIfNotExists(sheetName)

			// Добавляем строку
			err := tb.googleSheet.AppendRow(sheetName, []interface{}{update.Message.Text})
			if err != nil {
				log.Printf("Error appending to sheet: %v", err)
			}
		}
	}
}

func (tb *TelegramBot) handleCommand(update tgbotapi.Update) {
	if !tb.isAuthorizedUser(update.Message.From.ID) {
		return
	}
	switch update.Message.Command() {
	case "keywords_list":
		keywords := strings.Join(tb.keywords, ", ")
		if keywords == "" {
			keywords = "Список ключевых слов пуст."
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Текущие ключевые слова: "+keywords)
		tb.bot.Send(msg)

	case "keyword_add":
		args := strings.Split(update.Message.CommandArguments(), " ")
		for _, arg := range args {
			tb.keywords = append(tb.keywords, arg)
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ключевые слова добавлены.")
		tb.bot.Send(msg)

	case "keyword_remove":
		arg := update.Message.CommandArguments()
		for i, keyword := range tb.keywords {
			if keyword == arg {
				tb.keywords = append(tb.keywords[:i], tb.keywords[i+1:]...)
				break
			}
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ключевое слово удалено: "+arg)
		tb.bot.Send(msg)

	case "keyword_clean":
		tb.keywords = []string{}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Список ключевых слов очищен.")
		tb.bot.Send(msg)

	case "sheet_link":
		tb.sendSheetLink(update.Message.Chat.ID)

	case "help":
		tb.sendHelp(update.Message.Chat.ID)
	}
}

func (tb *TelegramBot) sendSheetLink(chatID int64) {
	url := tb.googleSheet.GetSheetLink()

	// Отправка сообщения
	msg := tgbotapi.NewMessage(chatID, "Ссылка на лист Google Sheets: "+url)
	tb.bot.Send(msg)
}

func (tb *TelegramBot) sendHelp(chatID int64) {
	helpText := "Список доступных команд:\n" +
		"/keywords_list - Показать текущие ключевые слова.\n" +
		"/keyword_add - Добавить ключевые слова.\n" +
		"/keyword_remove - Удалить ключевое слово.\n" +
		"/keyword_clean - Очистить список ключевых слов.\n" +
		"/sheet_link - Получить ссылку на текущий лист Google Sheets.\n" +
		"/help - Показать эту справку."

	msg := tgbotapi.NewMessage(chatID, helpText)
	tb.bot.Send(msg)
}

func (tb *TelegramBot) isAuthorizedUser(userID int64) bool {
	// Замените на логику проверки пользователя
	// Например, проверка на соответствие userID списку разрешенных ID
	return userID == tb.authorizedUserID
}
