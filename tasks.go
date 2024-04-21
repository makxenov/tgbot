package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	Descr      string
	Id         int
	CreateDate time.Time
	recurring  bool
	Score      int
}

func serialize(task Task) string {
	bytes, err := json.Marshal(task)
	_check(err)
	return string(bytes)
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
	var task Task
	err := json.Unmarshal([]byte(str), &task)
	return task, err
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
		res = append(res, fmt.Sprint(t.Id, ": ", t.Descr))
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
	tm.backlog[tm.backlogId] = Task{Id: tm.backlogId, Descr: descr, CreateDate: curTime(), recurring: false, Score: 0}
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
	task.Id = tm.dailyId
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
		list[task.Id] = task
		if task.Id > maxid {
			maxid = task.Id
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
