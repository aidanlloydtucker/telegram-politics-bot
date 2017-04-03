package main

import (
	"os"
	"time"

	"github.com/urfave/cli"
	"os/signal"
	"syscall"
	"log"
	"errors"
	"strconv"
)

func main() {
	app := cli.NewApp()

	app.Name = "Politics Bot"
	app.Usage = "Telegram bot"

	app.Authors = []cli.Author{
		{
			Name: "Aidan Lloyd-Tucker",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "token, t",
			Usage: "The telegram bot api token",
		},
		cli.StringFlag{
			Name:  "ip",
			Usage: "The IP for the webhook port",
		},
		cli.StringFlag{
			Name:  "webhook_port",
			Usage: "The telegram bot api webhook port",
			Value: "8443",
		},
		cli.StringFlag{
			Name:  "webhook_cert",
			Usage: "The telegram bot api webhook cert",
			Value: "./ignored/cert.pem",
		},
		cli.StringFlag{
			Name:  "webhook_key",
			Usage: "The telegram bot api webhook key",
			Value: "./ignored/key.key",
		},
		cli.BoolFlag{
			Name:  "enable_webhook, w",
			Usage: "Enables webhook if true",
		},
		cli.BoolFlag{
			Name:  "prod",
			Usage: "Sets bot to production mode",
		},
		cli.StringFlag{
			Name:  "congress-key",
			Usage: "ProPublica API key",
		},
		cli.IntFlag{
			Name:  "session",
			Usage: "Session of Congress",
			Value: 115,
		},
		cli.Int64SliceFlag{
			Name: "chats",
			Usage: "Chats to send updates to (v1)",
		},
	}

	app.Action = runApp
	app.Run(os.Args)
}

func runApp(c *cli.Context) error {
	log.Println("Running app")

	// Start bot

	var webhookConf *WebhookConfig = nil

	if c.IsSet("ip") && c.Bool("enable_webhook") {
		webhookConf = &WebhookConfig{
			IP:       c.String("ip"),
			CertPath: c.String("webhook_cert"),
			KeyPath:  c.String("webhook_key"),
			Port:     c.String("webhook_port"),
		}
	}

	log.Println("Starting bot and website")

	if !c.IsSet("congress-key") {
		return errors.New("Missing ProPublica API key")
	}

	congressKey := c.String("congress-key")
	congressSessionInt := c.Int("session")
	congressSession := strconv.Itoa(congressSessionInt)

	updateChats := c.Int64Slice("chats")

	go startBot(c.String("token"), webhookConf, updateChats, runGovUpdatePolling(congressKey, congressSession))

	// Safe Exit

	var Done = make(chan bool, 1)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		if runningWebhook {
			mainBot.RemoveWebhook()
		}

		Done <- true
	}()
	<-Done

	log.Println("Safe Exit")
	return nil
}

func runGovUpdatePolling(congressAPI string, congressSession string) chan interface{} {
	govUpdates := make(chan interface{}, 100)

	go func() {
		for {
			newEOs, err := getNewExecutiveOrders()
			if err != nil {
				log.Println("Error:", err)
				govUpdates <- NewServiceMessageUpdate("Error while getting executive orders: " + err.Error())
			} else {
				for _, eo := range newEOs {
					govUpdates <- eo
				}
			}
			time.Sleep(time.Hour)
		}
	}()

	go func() {
		for {
			newSenateBills, err := getNewSenateBills(congressAPI, congressSession)
			if err != nil {
				log.Println("Error:", err)
				govUpdates <- NewServiceMessageUpdate("Error while getting senate bills: " + err.Error())
			} else {
				for _, bill := range newSenateBills {
					govUpdates <- bill
				}
			}

			newHouseBills, err := getNewHouseBills(congressAPI, congressSession)
			if err != nil {
				log.Println("Error:", err)
				govUpdates <- NewServiceMessageUpdate("Error while getting house bills: " + err.Error())
			} else {
				for _, bill := range newHouseBills {
					govUpdates <- bill
				}
			}
			time.Sleep(time.Minute * 20)
		}
	}()

	return govUpdates
}

type ServiceMessageUpdate struct {
	Message string
	Time time.Time
}

func NewServiceMessageUpdate(msg string) ServiceMessageUpdate {
	return ServiceMessageUpdate{msg, time.Now()}
}