package main

import (
	"log"

	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"fmt"
)

var mainBot *tgbotapi.BotAPI
var runningWebhook bool

type WebhookConfig struct {
	IP       string
	Port     string
	KeyPath  string
	CertPath string
}

type WhoUser struct {
	*tgbotapi.User
	Choice int
}

type WhoMessage struct {
	Users []WhoUser
	Question string
	Choices []string
}

func startBot(token string, webhookConf *WebhookConfig, updateChats []int64, govUpdates chan interface{}) {
	log.Println("Starting Bot")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalln(err)
	}

	mainBot = bot

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	var updates <-chan tgbotapi.Update
	var webhookErr error

	if webhookConf != nil {
		_, webhookErr = bot.SetWebhook(tgbotapi.NewWebhookWithCert(webhookConf.IP+":"+webhookConf.Port+"/"+bot.Token, webhookConf.CertPath))
		if webhookErr != nil {
			log.Println("Webhook Error:", webhookErr, "Switching to poll")
		} else {
			runningWebhook = true
			updates = bot.ListenForWebhook("/" + bot.Token)
			go http.ListenAndServeTLS(webhookConf.IP+":"+webhookConf.Port, webhookConf.CertPath, webhookConf.KeyPath, nil)
			log.Println("Running on Webhook")
		}
	}

	if webhookErr != nil || webhookConf == nil {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, err = bot.GetUpdatesChan(u)
		if err != nil {
			log.Fatalln("Error found on getting poll updates:", err, "HALTING")
		}
		log.Println("Running on Poll")
	}

	// Send online status
	for _, chatID := range updateChats {
		msg := tgbotapi.NewMessage(chatID, "Bot connected to this chat")
		bot.Send(msg)
	}

	// Government Updates
	go func() {
		for {
			govUpdate := <- govUpdates
			switch govUpdate.(type) {
			case ExecutiveOrder:
				govTyp := govUpdate.(ExecutiveOrder)
				updateStr := parseExecutiveOrders(govTyp)
				for _, chatID := range updateChats {
					msg := tgbotapi.NewMessage(chatID, updateStr)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
			case Bill:
				govTyp := govUpdate.(Bill)
				updateStr := parseBills(govTyp)
				for _, chatID := range updateChats {
					msg := tgbotapi.NewMessage(chatID, updateStr)
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}
				return
			case ServiceMessageUpdate:
				govTyp := govUpdate.(ServiceMessageUpdate)
				for _, chatID := range updateChats {
					msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("SERVICE MESSAGE\n\n%s\nREPORTED: %s", govTyp.Message, govTyp.Time.String()))
					bot.Send(msg)
				}
			}
		}
	}()


	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.String(), update.Message.Text)
			if update.Message.Text == "/info" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%d", update.Message.Chat.ID))
				bot.Send(msg)
			}
		}
	}
}

func parseExecutiveOrders(eo ExecutiveOrder) string {
	return fmt.Sprintf("<b>Executive Order %d was Just Signed</b>\n" +
		"<i>%s</i>\n" +
		"Signed On: %s\n" +
		"<a href=\"%s\">Click for More</a>", eo.ExecutiveOrderNumber, eo.Title, eo.SigningDate, eo.HTMLUrl)
}

func parseBills(bill Bill) string {
	return fmt.Sprintf("<b>New Update to Bill %s</b>\n" +
		"<i>%s</i>\n" +
		"Last Major Action: %s\n" +
		"Date: %s\n" +
		"<a href=\"%s\">Click for More</a>", bill.Number, bill.Title, bill.LatestMajorAction, bill.LatestMajorActionDate, bill.GovtrackURL)
}