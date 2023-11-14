package main

import (
	"log"

	"github.com/mykyta-kravchenko98/telegram-parser/internal/config"
	"github.com/mykyta-kravchenko98/telegram-parser/internal/model"
	"github.com/mykyta-kravchenko98/telegram-parser/internal/service"
)

func main() {

	var conf *config.Config
	var confErr error
	//Load configuration
	conf, confErr = config.LoadConfigYAML()

	checkError(confErr)

	googleSheetService, err := service.NewGoogleSheetsService(conf.Google)
	checkError(err)

	bot, err := model.NewTelegramBot(conf.Telegram.Token, googleSheetService, []string{}, conf.Telegram.AdminId)
	checkError(err)

	bot.Listen()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
