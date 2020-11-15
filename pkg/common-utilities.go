package main

import (
	"errors"
	"strings"
	"time"
)

const releaseHistoryDateFormat = "2006-01-02T15:04:05"
const dateFormat = "2006-01-02T15:04:05.000-07:00"

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func parseTime(timeString string) time.Time {
	parsedTime, err := time.Parse(releaseHistoryDateFormat, timeString)
	if err == nil {
		return parsedTime
	}
	return time.Time{}
}

func arrayAverage(items []uint32) float32 {
	if len(items) == 0 {
		return 0
	}

	total := uint32(0)
	for i := 0; i < len(items); i++ {
		total += items[i]
	}
	return float32(total) / float32(len(items))
}

func arrayAverageDurationIgnoreZero(items []uint32) uint32 {
	total := uint32(0)
	count := uint32(0)
	for i := 0; i < len(items); i++ {
		if items[i] != 0 {
			total += items[i]
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / count
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func dateDiff(date1 string, date2 string) (time.Duration, error) {
	date1Parsed, err1 := time.Parse(releaseHistoryDateFormat, date1)
	date2Parsed, err2 := time.Parse(releaseHistoryDateFormat, date2)

	if err1 == nil && err2 == nil {
		return date1Parsed.Sub(date2Parsed), nil
	}

	return time.Duration(0), errors.New("failed to parse one or both dates")
}

func boolToInt(input bool) uint32 {
	bitSetVar := uint32(0)
	if input {
		bitSetVar = 1
	}
	return bitSetVar
}
