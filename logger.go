package main

import (
	"fmt"
	"strings"
	"time"
)

func sayHello() {
	fmt.Println("Hello")
	if fits("31 Dec 2022 14:44 Скоро Новый год!", 1) {
		fmt.Println("fits")
	} else {
		fmt.Println("Not fits")
	}
}

func curTime() time.Time {
	currentTime := time.Now()
	loc, err := time.LoadLocation("Asia/Nicosia")
	_check(err)
	return currentTime.In(loc)
}

func fits(logStr string, days int) bool {
	splitted := strings.Split(logStr, " ")
	dateStr := strings.Join(splitted[:4], " ")
	layout := "02 Jan 2006 15:04"
	t, err := time.Parse(layout, dateStr)
	_check(err)
	diff := time.Now().Sub(t)
	dayDiff := diff.Hours() / 24
	return dayDiff < float64(days)
}

type Logger struct {
	logDir string
}

func (l Logger) log(file string, text string) {
	path := l.logDir + file
	createFile(path)
	lines, err := readLines(path)
	_check(err)
	currentTime := curTime()

	lines = append(lines, currentTime.Format("02 Jan 2006 15:04")+" "+text)
	err = writeLines(lines, path)
	_check(err)
}

func (l Logger) query(file string, days int) ([]string, error) {
	path := l.logDir + file
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	size := 0
	var res []string
	for i := len(lines) - 1; i >= 0; i-- {
		str := lines[i]
		size += len([]rune(str)) + 4
		if fits(str, days) && size < 4096 {
			res = append(res, str)
		}
	}
	return res, nil
}
