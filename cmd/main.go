package main

import (
	mydb "github.com/StepanShevelev/tg-bot-article/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Показать статьи"),
		tgbotapi.NewKeyboardButton("Создать статью"),
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

func CreatePost(update tgbotapi.Update, bot *tgbotapi.BotAPI, user mydb.User) uint {
	var post mydb.Post

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	post.Title = msg.Text
	replacer := strings.NewReplacer(" ", "", "/", "", ".", "", ",", "", "!", "", ":", "", "?", "")
	post.Name = replacer.Replace(msg.Text)
	post.WhoCreatedMe = update.Message.From.UserName
	logrus.Info(post.Title)
	result := mydb.Database.Db.Select("Title", "Name", "WhoCreatedMe").Create(&post)

	//Id = post.ID
	logrus.Info(post.ID)

	if result.Error != nil {
		logrus.Info("Error occurred while creating a post")
		mydb.UppendErrorWithPath(result.Error)
	}

	msg.Text = "вставьте сссылку на главную картинку в формате https://images.pexels.com/your-picture.jpeg"
	if _, err := bot.Send(msg); err != nil {
		mydb.UppendErrorWithPath(err)
		logrus.Info("Error occurred while sending text", err)
	}

	user.Position = "after title create"
	result = mydb.Database.Db.Save(&user)
	if result.Error != nil {
		logrus.Info("Error occurred while updating user position")
		mydb.UppendErrorWithPath(result.Error)

	}

	return post.ID
}

func main() {

	//config := cfg.New()
	//if err := config.Load("./configs", "config", "yml"); err != nil {
	//	logrus.Info(err)
	//}

	mydb.ConnectToDb()
	logrus.Info("Connected to db")

	var user mydb.User
	var post mydb.Post
	var Id uint

	bot, err := tgbotapi.NewBotAPI(" ")
	if err != nil {
		logrus.Info(err)
	}

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

			user.Position = "after button show"
			result := mydb.Database.Db.Save(&user)
			logrus.Info(user.Position)
			if result.Error != nil {
				logrus.Info("Error occurred while updating user position")
				mydb.UppendErrorWithPath(result.Error)
				return
			}

			if _, err = bot.Send(msg); err != nil {
				mydb.UppendErrorWithPath(err)
				logrus.Info("Error occurred while sending message", err)
			}

		case "Создать статью":

			msg.Text = "введите название статьи"
			if _, err = bot.Send(msg); err != nil {
				mydb.UppendErrorWithPath(err)
				logrus.Info("Error occurred while sending text", err)
			}

			user.Position = "after button create"
			result := mydb.Database.Db.Save(&user)
			if result.Error != nil {
				logrus.Info("Error occurred while updating user position")
				mydb.UppendErrorWithPath(result.Error)
				return
			}

		case msg.Text:

			if user.Position == "after button create" {
				Id = CreatePost(update, bot, user)
				continue

			}

			if user.Position == "after title create" {

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

				result := mydb.Database.Db.Find(&post, "id = ?", Id)
				logrus.Info(Id)
				if result.Error != nil {
					logrus.Info("Error occurred while searching post")
					mydb.UppendErrorWithPath(result.Error)
					return
				}

				_, err := url.ParseRequestURI(msg.Text)
				if err != nil {
					logrus.Info("Not picture URL")
					mydb.UppendErrorWithPath(err)
					msg.Text = "Проверь правильность ссылки"
					if _, err = bot.Send(msg); err != nil {
						mydb.UppendErrorWithPath(err)
						logrus.Info("Error occurred while sending text", err)
					}
					continue
				}

				image := mydb.Image{
					Name:   []byte(msg.Text),
					PostID: Id,
				}

				result = mydb.Database.Db.Create(&image)
				if result.Error != nil {
					logrus.Info("Error occurred while updating post text")
					mydb.UppendErrorWithPath(result.Error)
					return
				}

				msg.Text = "введите текст статьи"
				if _, err = bot.Send(msg); err != nil {
					mydb.UppendErrorWithPath(err)
					logrus.Info("Error occurred while sending text", err)
				}

				user.Position = "after picture create"
				result = mydb.Database.Db.Save(&user)
				if result.Error != nil {
					logrus.Info("Error occurred while updating user position")
					mydb.UppendErrorWithPath(result.Error)
					return

				}
				continue
			}

			if user.Position == "after picture create" {

				//result := mydb.Database.Db.Find(&post, "id = ?", Id)
				//logrus.Info(Id)
				//if result.Error != nil {
				//	logrus.Info("Error occurred while searching post")
				//	mydb.UppendErrorWithPath(result.Error)
				//	return
				//}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				//mydb.Database.Db.Model(&post).Where("id = ?", Id).UpdateColumn("text", msg.Text)
				//mydb.Database.Db.Model(&post).Update("text", msg.Text)
				result := mydb.Database.Db.Model(&post).Where("id = ?", Id).Update("Text", msg.Text)
				logrus.Info(post.Text)
				logrus.Info(Id)
				if result.Error != nil {
					logrus.Info("Error occurred while updating post text")
					mydb.UppendErrorWithPath(result.Error)
					return
				}
				msg.Text = "Вставьте все ссылки на картинки одним сообщением"
				if _, err = bot.Send(msg); err != nil {
					mydb.UppendErrorWithPath(err)
					logrus.Info("Error occurred while sending text", err)
				}

				user.Position = "after text create"
				result = mydb.Database.Db.Save(&user)
				if result.Error != nil {
					logrus.Info("Error occurred while updating user position")
					mydb.UppendErrorWithPath(result.Error)
					return
				}
				continue
			}

			if user.Position == "after text create" {

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

				result := mydb.Database.Db.Find(&post, "id = ?", Id)
				if result.Error != nil {
					logrus.Info("Error occurred while searching post")
					mydb.UppendErrorWithPath(result.Error)
					return
				}

				var imgMass []string
				var imgB [][]byte

				imgMass = strings.Split(update.Message.Text, "\n")

				for _, img := range imgMass {
					imgB = append(imgB, []byte(img))
				}

				for _, img := range imgB {
					var images = []mydb.Image{{Name: img, PostID: Id}}
					result := mydb.Database.Db.Create(&images)
					if result.Error != nil {
						logrus.Info("Error occurred while creating an image")
						mydb.UppendErrorWithPath(result.Error)
						return
					}
				}

				msg.Text = "вы создали статью"
				if _, err = bot.Send(msg); err != nil {
					mydb.UppendErrorWithPath(err)
					logrus.Info("Error occurred while sending text", err)
				}

				user.Position = "before button"
				result = mydb.Database.Db.Save(&user)
				if result.Error != nil {
					logrus.Info("Error occurred while updating user position")
					mydb.UppendErrorWithPath(result.Error)
					return

				}
				continue
			}

			if user.Position == "after button show" {
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

					user.Position = "before button"
					result := mydb.Database.Db.Save(&user)
					if result.Error != nil {
						logrus.Info("Error occurred while updating user position")
						mydb.UppendErrorWithPath(result.Error)
						return
					}
					continue
				}
			}
		}

		//if user.Position == "before button" {
		//	msg.Text = update.Message.Text
		//	if _, err = bot.Send(msg); err != nil {
		//		mydb.UppendErrorWithPath(err)
		//		logrus.Info("Error occurred while sending text", err)
		//	}
		//
		//}
	}

	//telegram.BotServe()

}
