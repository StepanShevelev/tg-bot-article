package main

import (
	mydb "github.com/StepanShevelev/tg-bot-article/db"
	cfg "github.com/StepanShevelev/tg-bot-article/pkg/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("get an article", "CALLBACKDATA"),
	),
)

func main() {

	mydb.ConnectToDb()

	config := cfg.New()
	if err := config.Load("./configs", "config", "yml"); err != nil {
		mydb.UppendErrorWithPath(err)
		logrus.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		mydb.UppendErrorWithPath(err)
		logrus.Fatal(err)
	}

	//bot.Debug = true
	logrus.WithFields(logrus.Fields{
		"BotUserName": bot.Self.UserName,
	}).Info("Authorized in account")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery.Data == "CALLBACKDATA" {
			logrus.Info("Нажата клавиша, = get an article !")
		}
		if update.Message == nil { // ignore non-Message updates
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		switch update.Message.Text {
		case "open":
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}

		_, err := bot.Send(msg)
		if err != nil {
			mydb.UppendErrorWithPath(err)
			logrus.Fatal(err)
		}
	}
}
