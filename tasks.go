package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	descr      string
	id         int
	createDate time.Time
	recurring  bool
	score      int
}

func serialize(task Task) string {
	dateStr := task.createDate.Format("02 Jan 2006")
	return fmt.Sprintf("%s$$%d$$%s$$%t$$%d", task.descr, task.id, dateStr, task.recurring, task.score)
}

type ParseError struct{}

func (m *ParseError) Error() string {
	return "Parse error"
}

func parseInt(str string) int {
	i, err := strconv.Atoi(str)
	_check(err)
	return i
}

func deserialize(str string) (Task, error) {
	splitted := strings.Split(str, "$$")
	if len(splitted) != 5 {
		return Task{}, &ParseError{}
	}
	layout := "02 Jan 2006"
	t, err := time.Parse(layout, splitted[2])
	_check(err)
	id := parseInt(splitted[1])
	score := parseInt(splitted[4])
	recurring, err := strconv.ParseBool(splitted[3])
	_check(err)
	task := Task{descr: splitted[0], id: id, createDate: t, recurring: recurring, score: score}
	return task, nil
}

type TaskManager struct {
	taskDir    string
	daily      map[int]Task
	dailyId    int
	dailyRec   map[int]Task
	dailyRecId int
	weekly     map[int]Task
	weeklyId   int
}

func (tm TaskManager) summ() string {
	res := ""
	for _, t := range tm.daily {
		res += fmt.Sprint(t.id, ": ", t.descr, " (", t.score, ")") + "\n"
	}
	return res
}

func (tm *TaskManager) addDaily(descr string) {
	tm.daily[tm.dailyId] = Task{id: tm.dailyId, descr: descr, createDate: curTime(), recurring: false, score: 0}
	tm.dailyId += 1
}

func (tm *TaskManager) completeDaily(id int) bool {
	_, ok := tm.daily[id]
	if ok {
		delete(tm.daily, id)
	}
	return ok
}

func (tm TaskManager) dump() {
	path := tm.taskDir + "daily"
	createFile(path)
	var lines []string
	for _, t := range tm.daily {
		lines = append(lines, serialize(t))
	}
	writeLines(lines, path)
	var id [1]string
	id[0] = fmt.Sprint(tm.dailyId)
	id_file := tm.taskDir + "ids"
	writeLines(id[:], id_file)
}

func (tm *TaskManager) load() {
	path := tm.taskDir + "daily"
	lines, e := readLines(path)
	if e != nil {
		return
	}
	for _, line := range lines {
		task, e := deserialize(line)
		_check(e)
		tm.daily[task.id] = task
	}
	id_file := tm.taskDir + "ids"
	indices, e := readLines(id_file)
	_check(e)
	tm.dailyId = parseInt(indices[0])
}
