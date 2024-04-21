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
	taskDir   string
	daily     map[int]Task
	dailyId   int
	backlog   map[int]Task
	backlogId int
}

func taskListDescr(list map[int]Task) string {
	res := []string{}
	for _, t := range list {
		res = append(res, fmt.Sprint(t.id, ": ", t.descr))
	}
	return strings.Join(res[:], "\n")
}

func (tm TaskManager) dailySumm() string {
	return taskListDescr(tm.daily)
}

func (tm TaskManager) backlogSumm() string {
	return taskListDescr(tm.backlog)
}

func (tm *TaskManager) addBacklog(descr string) {
	tm.backlog[tm.backlogId] = Task{id: tm.backlogId, descr: descr, createDate: curTime(), recurring: false, score: 0}
	tm.backlogId += 1
}

func (tm *TaskManager) completeDaily(id int) bool {
	_, ok := tm.daily[id]
	if ok {
		delete(tm.daily, id)
	}
	return ok
}

func (tm *TaskManager) takeToDaily(id int, preserveBacklog bool) bool {
	task, ok := tm.backlog[id]
	if !ok {
		return false
	}
	task.id = tm.dailyId
	tm.daily[tm.dailyId] = task
	tm.dailyId += 1
	if !preserveBacklog {
		delete(tm.backlog, id)
	}
	return true
}

func dumpTaskMap(list map[int]Task, path string) {
	createFile(path)
	var lines []string
	for _, t := range list {
		lines = append(lines, serialize(t))
	}
	writeLines(lines, path)
}

func loadTaskMap(list map[int]Task, path string) int {
	lines, e := readLines(path)
	if e != nil {
		return 0
	}
	maxid := 0
	for _, line := range lines {
		task, e := deserialize(line)
		_check(e)
		list[task.id] = task
		if task.id > maxid {
			maxid = task.id
		}
	}
	return maxid + 1
}

func (tm TaskManager) dump() {
	dumpTaskMap(tm.daily, tm.taskDir+"daily")
	dumpTaskMap(tm.backlog, tm.taskDir+"backlog")
}

func (tm *TaskManager) load() {
	tm.dailyId = loadTaskMap(tm.daily, tm.taskDir+"daily")
	tm.backlogId = loadTaskMap(tm.backlog, tm.taskDir+"backlog")
}
