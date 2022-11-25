package utils

import (
	"time"
)

const TimeIso8601 = "2006-01-02T15:04:05Z"

var CST, _ = time.LoadLocation("Asia/Shanghai")

func init() {
	if CST == nil {
		CST = time.FixedZone("CST", 8*3600)
	}
}

func BeginningOfUtcDay(t time.Time) time.Time {
	t = t.UTC()
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func BeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func DaysBetweenTwoDates(date1, date2 time.Time) int64 {
	if date2.Before(date1) {
		date1, date2 = date2, date1
	}

	return int64(date2.Sub(date1).Hours() / 24)
}
