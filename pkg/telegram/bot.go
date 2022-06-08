package telegram

import (
	cfg "github.com/StepanShevelev/tg-bot-article/pkg/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

func BotInit(config *cfg.Config) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	//bot.Debug = true
	log.WithFields(log.Fields{
		"BotUserName": bot.Self.UserName,
	}).Info("Authorized in account")

	bot.RemoveWebhook()
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(
		config.WebhookURL + "/" + bot.Token))
	if err != nil {
		log.Fatal(err)
	}
	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}
	if info.LastErrorDate != 0 {
		log.WithFields(log.Fields{
			"LastErrorMessege": info.LastErrorMessage,
		}).Warn("Telgram callback failed")
	}
	updates := bot.ListenForWebhook("/" + bot.Token)
	return bot, updates
}
