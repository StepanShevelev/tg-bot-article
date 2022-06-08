package main

import (
	mydb "github.com/StepanShevelev/tg-bot-article/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

//var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
//	tgbotapi.NewInlineKeyboardRow(
//		tgbotapi.NewInlineKeyboardButtonData("Показать статьи", showArticles()),
//	))

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Показать статьи"),
	),
)

func showArticles() string {

	postMap, err := mydb.GetPosts()
	if err != nil {
		mydb.UppendErrorWithPath(err)
		logrus.Info("Could not find post", err)
		return ""
	}
	msg := "Доступные статьи:"
	for i, _ := range postMap {
		msg = msg + "\n"
		msg += postMap[i] + "\n"
	}

	return msg
}

func main() {

	//post, err := mydb.GetPostByTitle()

	mydb.ConnectToDb()

	bot, err := tgbotapi.NewBotAPI("5085408878:AAGMbrpXnjoJtZRYVYkKCvxeWYLfTztofHI")
	if err != nil {
		logrus.Info(err)
	}

	bot.Debug = true
	bot.RemoveWebhook()
	logrus.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		// Check if we've gotten a message update.
		if update.Message != nil {

			// Construct a new message from the given chat ID and containing
			// the text that we received.
			file := mydb.CreateHTML(update.Message.Text)
			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//msg.ReplyMarkup = numericKeyboard
			msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, file)
			//switch update.Message.Text {
			//case "Показать статьи":
			//	msg.Text = showArticles()
			//case "Обзор Vampire: The Masquerade — Swansong. Отцы и дети и вампиры":
			//	bot.Send(doc)
			//	return
			//}

			//Send the message.
			if _, err = bot.Send(msg); err != nil {
				//mydb.UppendErrorWithPath(err)
				logrus.Fatal(err)
			}
		}

	}
	//mydb.CreateHTML()
}
