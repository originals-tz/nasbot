package main

import (
	"fmt"
	"io/ioutil"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	TOKEN string
}

const (
	cmdNone = iota
	cmdState
	cmdSearch
	cmdDownload
)

type BotMessage struct {
	session int
	param   string
}

type TelegramBot struct {
	token    string
	bot      *tgbotapi.BotAPI
	bot_chan chan BotMessage
}

func (tg TelegramBot) send(msg tgbotapi.MessageConfig) {
	if _, err := tg.bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func getMessageSession(msg tgbotapi.Message) (bool, int, string) {
	var session int
	var reply string
	switch msg.Command() {
	case "state":
		return true, cmdState, "get state..."
	case "search":
		session = cmdSearch
		reply = "search from 36dm..."
	case "down":
		session = cmdDownload
		reply = "wakeup qBittorrent to download..."
	default:
		return false, cmdNone, "no such command"
	}

	if msg.CommandArguments() == "" {
		return false, session, "error, no params"
	}
	return true, session, reply
}

func (tg TelegramBot) run() {
	var err error
	tg.bot, err = tgbotapi.NewBotAPI(tg.token)
	if err != nil {
		log.Panic(err)
	}
	tg.bot.Debug = true
	log.Printf("Authorized on account %s", tg.bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tg.bot.GetUpdatesChan(u)
	session := cmdNone
	var bot_msg BotMessage
	for update := range updates {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message == nil {
			continue
		}
		if !update.Message.IsCommand() {
			msg.Text = "sorry, use /cmd param"
			tg.send(msg)
			continue
		}

		ok, tmp_seesion, reply := getMessageSession(*update.Message)
		msg.Text = reply
		tg.send(msg)
		if ok {
			session = tmp_seesion
			bot_msg.session = session
			tg.bot_chan <- bot_msg
		}
	}
}

func telegramBot(bot_chan chan BotMessage) {
	var config Config
	File, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(File, &config)
	if err != nil {
		log.Panic(err)
	}

	var tg TelegramBot
	tg.bot_chan = bot_chan
	tg.token = config.TOKEN
	tg.run()
}

func main() {
	bot_chan := make(chan BotMessage)
	go telegramBot(bot_chan)
	for true {
		select {
		case msg := <-bot_chan:
			fmt.Print(msg.session, ":", msg.param)
		}
	}
}
