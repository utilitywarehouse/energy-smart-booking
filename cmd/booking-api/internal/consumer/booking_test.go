package consumer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/googleapis/type/date"
)

func Test_dateIntoTime(t *testing.T) {

	expected := time.Date(2020, time.August, 5, 0, 0, 0, 0, time.UTC)

	actual, err := dateIntoTime(&date.Date{
		Year:  2020,
		Month: 8,
		Day:   5,
	})

	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	if !a.Equal(&expected, actual) {
		t.Fail()
	}
}
