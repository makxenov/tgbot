package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"

	// "net/http"
	"sync"
)

var mutex sync.Mutex

func _check(err error) {
	if err != nil {
		panic(err)
	}
}

var bot *tgbotapi.BotAPI
var gerr error
var c chan int64
var tm *TaskManager

func commandDispatcher() {
	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)
	_check(err)

	data_folder := os.Getenv("DATA_FOLDER") + "/"
	log_folder := data_folder + "logs/"
	os.MkdirAll(log_folder, os.ModePerm)
	task_folder := data_folder + "tasks/"
	os.MkdirAll(task_folder, os.ModePerm)
	logger := Logger{log_folder}
	tm = &TaskManager{
		taskDir:    task_folder,
		daily:      make(map[int]Task),
		dailyId:    0,
		dailyRec:   make(map[int]Task),
		dailyRecId: 0,
		weekly:     make(map[int]Task),
		weeklyId:   0,
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.From.UserName != "makxenov" {
			fmt.Println("Unexpected user: " + update.Message.From.UserName)
			continue
		}

		// Make sure that message in text
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

			chat_folder := data_folder + strconv.FormatInt(update.Message.Chat.ID, 10)
			splitted := strings.Split(update.Message.Text, " ")
			command := splitted[0]

			switch command {
			case "/start":

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi, i'm a Voice of mind bot!")
				bot.Send(msg)
				msg2 := tgbotapi.NewMessage(update.Message.Chat.ID, "New user: @"+update.Message.From.UserName)
				bot.Send(msg2)
				c <- update.Message.Chat.ID

				fmt.Println("Start chat with id:" + strconv.FormatInt(update.Message.Chat.ID, 10) + ". User: @" + update.Message.From.UserName)

			case "/log":
				if len(splitted) < 2 {
					err_msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Log category is not specified")
					bot.Send(err_msg)
					continue
				}
				file := splitted[1]
				comment := ""
				if len(splitted) > 2 {
					comment = strings.Join(splitted[2:], " ")
				}
				logger.log(file, comment)
				lines, err := logger.query(file, 7)
				_check(err)
				stat := tgbotapi.NewMessage(update.Message.Chat.ID, strings.Join(lines, "\n"))
				bot.Send(stat)

			case "/logcat":
				if len(splitted) < 2 {
					err_msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Log category is not specified")
					bot.Send(err_msg)
					continue
				}
				file := splitted[1]
				lines, err := logger.query(file, 7)
				if err != nil {
					err_msg := tgbotapi.NewMessage(update.Message.Chat.ID, "File not found")
					bot.Send(err_msg)
				}
				stat := tgbotapi.NewMessage(update.Message.Chat.ID, strings.Join(lines, "\n"))
				bot.Send(stat)

			case "/daily":
				if len(splitted) < 2 {
					err_msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Task is not specified")
					bot.Send(err_msg)
					continue
				}
				task_descr := strings.Join(splitted[1:], " ")
				tm.load()
				tm.addDaily(task_descr)
				summ_msg := tgbotapi.NewMessage(update.Message.Chat.ID, tm.summ())
				bot.Send(summ_msg)
				tm.dump()

			case "/tasks":
				tm.load()
				summ_msg := tgbotapi.NewMessage(update.Message.Chat.ID, tm.summ())
				bot.Send(summ_msg)

			case "/complete":
				if len(splitted) != 2 {
					err_msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Id is not specified")
					bot.Send(err_msg)
					continue
				}
				id := splitted[1]
				tm.load()
				if !tm.completeDaily(parseInt(id)) {
					err_msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown task ID")
					bot.Send(err_msg)
					continue
				}
				summ_msg := tgbotapi.NewMessage(update.Message.Chat.ID, tm.summ())
				bot.Send(summ_msg)
				tm.dump()

			case "/stop":
				err := os.RemoveAll(chat_folder)
				_check(err)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You successfully stop following all advertisements")
				bot.Send(msg)

			default:
				fmt.Println("Other command: " + command + " orig text: " + update.Message.Text)
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send command")
			bot.Send(msg)

		}
	}
}

func noonTask(id int64) {
	tm.load()
	msg := tgbotapi.NewMessage(id, tm.summ())
	bot.Send(msg)
}
func initNoon() {
	t := curTime()
	id := <-c
	fmt.Println("Id received")
	loc, err := time.LoadLocation("Asia/Nicosia")
	_check(err)
	n := time.Date(t.Year(), t.Month(), t.Day(), 20, 0, 0, 0, loc)
	d := n.Sub(t)
	if d < 0 {
		n = n.Add(24 * time.Hour)
		d = n.Sub(t)
	}
	for {
		time.Sleep(d)
		d = 24 * time.Hour
		noonTask(id)
	}
}

func main() {
	sayHello()
	bot, gerr = tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	_check(gerr)
	c = make(chan int64)
	go commandDispatcher()
	go initNoon()

	checking_interval := 300

	if os.Getenv("CHECKING_INTERVAL") != "" {
		parsed_int, err := strconv.Atoi(os.Getenv("CHECKING_INTERVAL"))
		_check(err)
		checking_interval = parsed_int
	}

	for {
		//
		time.Sleep(time.Second * time.Duration(checking_interval))
	}
}
