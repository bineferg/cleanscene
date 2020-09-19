package event

import (
	"strings"
	"time"
)

type Event struct {
	Title    string
	Location string
	City     string
	Country  string
	AirCode  string
}

func ParseDate(eventStr string, startIdx int) (time.Time, int) {
	eventDateStr := eventStr[startIdx:16]
	dateTindex := strings.Index(eventDateStr, "T")
	dateToParse := eventDateStr[:dateTindex]
	d, _ := time.Parse("2006-01-02", dateToParse)
	return d, 17
}

func ParseTitle(eventStr string, startIdx int) (string, int) {
	titleStart := eventStr[startIdx:]
	idx := strings.Index(titleStart, "/")
	eventRest := titleStart[idx+1:]
	// Separate title from Venue
	titleEnd := strings.Index(eventRest, "/")
	title := eventRest[:titleEnd]
	return title, titleEnd
}
