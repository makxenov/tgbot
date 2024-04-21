package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

var mutex sync.Mutex

func _check(err error) {
	if err != nil {
		panic(err)
	}
}

var bot *BotApi
var tm *TaskManager

func commandDispatcher() {
	updates := bot.getUpdatesChan()

	data_folder := os.Getenv("DATA_FOLDER") + "/"
	log_folder := data_folder + "logs/"
	os.MkdirAll(log_folder, os.ModePerm)
	task_folder := data_folder + "tasks/"
	os.MkdirAll(task_folder, os.ModePerm)
	logger := Logger{log_folder}
	tm = &TaskManager{
		taskDir:   task_folder,
		daily:     make(map[int]Task),
		dailyId:   0,
		backlog:   make(map[int]Task),
		backlogId: 0,
	}

	for message := range updates {
		splitted := strings.Split(message, " ")
		command := splitted[0]

		switch command {
		case "/start":
			bot.send("Hi, i'm a Voice of mind bot!")

		case "/log":
			if len(splitted) < 2 {
				bot.send("Log category is not specified")
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
			stat := strings.Join(lines, "\n")
			bot.send(stat)

		case "/logcat":
			if len(splitted) < 2 {
				bot.send("Log category is not specified")
				continue
			}
			file := splitted[1]
			lines, err := logger.query(file, 7)
			if err != nil {
				bot.send("File not found")
			}
			stat := strings.Join(lines, "\n")
			bot.send(stat)

		case "/backlog":
			tm.load()
			if len(splitted) < 2 {
				bot.send(tm.backlogSumm())
				continue
			}
			task_descr := strings.Join(splitted[1:], " ")
			tm.addBacklog(task_descr)
			summ_msg := tm.backlogSumm()
			bot.send(summ_msg)
			tm.dump()

		case "/tasks":
			tm.load()
			summ_msg := tm.dailySumm()
			bot.send(summ_msg)

		case "/complete", "/giveup":
			if len(splitted) != 2 {
				bot.send("Id is not specified")
				continue
			}
			id := splitted[1]
			tm.load()
			if !tm.completeDaily(parseInt(id)) {
				bot.send("Unknown task ID")
				continue
			}
			summ_msg := tm.dailySumm()
			bot.send(summ_msg)
			tm.dump()

		case "/take":
			if len(splitted) != 2 {
				bot.send("Id is not specified")
				continue
			}
			id := splitted[1]
			tm.load()
			if !tm.takeToDaily(parseInt(id), false) {
				bot.send("Unknown task ID")
				continue
			}
			summ_msg := tm.dailySumm()
			bot.send(summ_msg)
			tm.dump()

		case "/pick":
			if len(splitted) != 2 {
				bot.send("Id is not specified")
				continue
			}
			id := splitted[1]
			tm.load()
			if !tm.takeToDaily(parseInt(id), true) {
				bot.send("Unknown task ID")
				continue
			}
			summ_msg := tm.dailySumm()
			bot.send(summ_msg)
			tm.dump()

		case "/stop":
			bot.send("Sorry, stop is not implemented, keep going!")

		default:
			fmt.Println("Other command: " + command + " orig text: " + message)
		}
	}
}

func main() {
	bot = &BotApi{
		isDebug: false,
		bot:     nil,
		chatID:  0,
	}
	if len(os.Args) == 2 && os.Args[1] == "debug" {
		bot.isDebug = true
	}
	bot.init()
	initCallbacks()

	if bot.isDebug {
		go func() {
			defer close(bot.adapterChan)
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				bot.adapterChan <- scanner.Text()
			}

			if scanner.Err() != nil {
				fmt.Println(scanner.Err().Error())
			}
		}()
	}

	commandDispatcher()
}
