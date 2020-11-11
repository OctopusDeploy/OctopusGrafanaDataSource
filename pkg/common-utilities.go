package main

import (
	"strings"
	"time"
)

const dateFormat = "2006-01-02T15:04:05"

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
	parsedTime, err := time.Parse(dateFormat, timeString)
	if err != nil {
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

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
