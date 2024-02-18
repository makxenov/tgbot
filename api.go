package main

import (
	"fmt"
	"os"
	"reflect"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

type BotApi struct {
	isDebug     bool
	bot         *tgbotapi.BotAPI
	chatID      int64
	adapterChan chan string
}

func redirect(api *BotApi, from tgbotapi.UpdatesChannel) {
	for {
		update := <-from
		if update.Message == nil {
			continue
		}

		if update.Message.From.UserName != "makxenov" {
			fmt.Println("Unexpected user: " + update.Message.From.UserName)
			continue
		}

		// Make sure that message is text
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {
			api.chatID = update.Message.Chat.ID
			api.adapterChan <- update.Message.Text
		} else {
			api.send("Send command")

		}
	}
}

func (api *BotApi) getUpdatesChan() chan string {
	api.adapterChan = make(chan string)
	if !api.isDebug {
		u := tgbotapi.NewUpdate(0)

		updates, err := api.bot.GetUpdatesChan(u)
		_check(err)
		go redirect(api, updates)
	}
	return api.adapterChan
}

func (api *BotApi) init() {
	if !api.isDebug {
		var err error
		api.bot, err = tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
		_check(err)
	}
}

func (api *BotApi) send(text string) {
	if !api.isDebug {
		msg := tgbotapi.NewMessage(api.chatID, text)
		_, e := api.bot.Send(msg)
		if e != nil {
			fmt.Println(e.Error())
		}
	} else {
		fmt.Println(text)
	}
}
