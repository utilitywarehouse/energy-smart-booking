package domain_test

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
