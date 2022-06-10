package main

import (
	mydb "github.com/StepanShevelev/tg-bot-article/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

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
	for i := range postMap {
		msg = msg + "\n"
		msg += postMap[i] + "\n"
	}

	return msg
}

func main() {

	mydb.ConnectToDb()

	bot, err := tgbotapi.NewBotAPI(" ")
	if err != nil {
		logrus.Info(err)
	}

	var user mydb.User

	bot.Debug = true
	_, err = bot.RemoveWebhook()
	if err != nil {
		mydb.UppendErrorWithPath(err)
		logrus.Info("Webhook error", err)
	}
	logrus.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {

		logrus.WithFields(logrus.Fields{
			"UserName": update.Message.From.UserName,
			"Text":     update.Message.Text,
		}).Info("Message from User")

		if update.Message != nil {

			result := mydb.Database.Db.First(&user, "name = ?", update.Message.From.UserName)
			if result.Error != nil {

				logrus.Info("Error occurred while searching user")
				mydb.UppendErrorWithPath(result.Error)
				mydb.CreateUser(update.Message.From.UserName, "before button")
			}
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyMarkup = numericKeyboard
				if _, err = bot.Send(msg); err != nil {
					mydb.UppendErrorWithPath(err)
					logrus.Info("Error occurred while sending message", err)
				}
			}
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		switch update.Message.Text {

		case "Показать статьи":
			msg.Text = showArticles()

			user.Position = "after button"
			result := mydb.Database.Db.Save(&user)
			if result.Error != nil {
				logrus.Info("Error occurred while updating user position")
				mydb.UppendErrorWithPath(result.Error)
				return
			}

			if _, err = bot.Send(msg); err != nil {
				mydb.UppendErrorWithPath(err)
				logrus.Info("Error occurred while sending message", err)
			}

		case msg.Text:

			if user.Position == "after button" {
				post, err := mydb.GetPostByTitle(msg.Text)
				if err != nil {
					mydb.UppendErrorWithPath(err)
					logrus.Info("Error occurred while calling CreateHTML", err)
				}

				if msg.Text == post.Title {
					file, err := mydb.CreateHTML(update.Message.Text, update.Message.From.UserName)
					if err != nil {
						mydb.UppendErrorWithPath(err)
						logrus.Info("Error occurred while calling CreateHTML", err)
					}

					doc := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, file)
					if _, err = bot.Send(doc); err != nil {
						mydb.UppendErrorWithPath(err)
						logrus.Info("Error occurred while sending document", err)
					}
					user.Name = update.Message.From.UserName
					user.Position = "before button"
					result := mydb.Database.Db.Save(&user)
					if result.Error != nil {
						logrus.Info("Error occurred while updating user position")
						mydb.UppendErrorWithPath(result.Error)
						return
					}
				}
			} else {
				msg.Text = update.Message.Text
				if _, err = bot.Send(msg); err != nil {
					mydb.UppendErrorWithPath(err)
					logrus.Info("Error occurred while sending text", err)
				}
			}
		}
	}
}
