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
		taskDir:    task_folder,
		daily:      make(map[int]Task),
		dailyId:    0,
		dailyRec:   make(map[int]Task),
		dailyRecId: 0,
		backlog:    make(map[int]Task),
		weeklyId:   0,
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

		case "/daily":
			if len(splitted) < 2 {
				bot.send("Task is not specified")
				continue
			}
			task_descr := strings.Join(splitted[1:], " ")
			tm.load()
			tm.addDaily(task_descr)
			summ_msg := tm.summ()
			bot.send(summ_msg)
			tm.dump()

		case "/tasks":
			tm.load()
			summ_msg := tm.summ()
			bot.send(summ_msg)

		case "/complete":
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
			summ_msg := tm.summ()
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
