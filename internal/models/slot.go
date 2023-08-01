package models

import "google.golang.org/genproto/googleapis/type/date"

type Slot struct {
	Date      date.Date
	StartTime int32
	EndTime   int32
}
