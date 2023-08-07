package utilities

import (
	"fmt"
	"time"

	"google.golang.org/genproto/googleapis/type/date"
)

func DateIntoTime(d *date.Date) (*time.Time, error) {
	t, err := time.ParseInLocation(
		time.DateOnly,
		fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day),
		time.UTC,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
