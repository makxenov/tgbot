package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

func _check(err error) {
	if err != nil {
		panic(err)
	}
}

var bot *BotApi
var c chan int64
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
		weekly:     make(map[int]Task),
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

func noonTask() {
	tm.load()
	bot.send(tm.summ())
}
func initNoon() {
	t := curTime()
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
		noonTask()
	}
}

func main() {
	sayHello()
	bot = &BotApi{
		isDebug: false,
		bot:     nil,
		chatID:  0,
	}
	if len(os.Args) == 2 && os.Args[1] == "debug" {
		bot.isDebug = true
	}
	bot.init()
	go commandDispatcher()
	go initNoon()

	checking_interval := 300

	if os.Getenv("CHECKING_INTERVAL") != "" {
		parsed_int, err := strconv.Atoi(os.Getenv("CHECKING_INTERVAL"))
		_check(err)
		checking_interval = parsed_int
	}
	if bot.isDebug {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			bot.adapterChan <- scanner.Text()
		}

		if scanner.Err() != nil {
			fmt.Println(scanner.Err().Error())
		}
		return
	}

	for {
		//
		time.Sleep(time.Second * time.Duration(checking_interval))
	}
}
