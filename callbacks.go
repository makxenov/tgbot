package main

import (
	"time"
)

type doer func()

func notify(since time.Time, period time.Duration, handler doer) {
	delta := since.Sub(curTime())
	time.Sleep(delta)
	for {
		handler()
		time.Sleep(period)
	}
}

func dailyNotification() {
	tm.load()
	bot.send(tm.summ())
}

func nearestHour(hour int) time.Time {
	now := curTime()
	loc, err := time.LoadLocation("Asia/Nicosia")
	_check(err)
	result := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, loc)
	if result.Before(now) {
		result = result.Add(24 * time.Hour)
	}
	return result
}

func nearestWeakday(weekday time.Weekday, hour int) time.Time {
	hourPrecision := nearestHour(hour)
	today := hourPrecision.Weekday()
	diff := int(weekday-today+7) % 7
	return hourPrecision.Add(time.Duration(diff*24) * time.Hour)
}

func weeklyPlanning() {
	tm.load()
	bot.send("Plan")
}

func initCallbacks() {
	// daily task reminder
	go notify(nearestHour(20), 24*time.Hour, dailyNotification)
	// weekly reminder
	go notify(nearestWeakday(time.Sunday, 21), 24*7*time.Hour, weeklyPlanning)
}
