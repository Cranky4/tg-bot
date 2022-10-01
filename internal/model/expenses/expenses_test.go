package expenses

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ExpensePeriodWeek_ShouldReturnCorrectStartDate(t *testing.T) {
	p := Week

	exp, err := time.Parse("2006-01-02 15:04:05", "2022-09-24 00:00:00")
	assert.NoError(t, err)

	now, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 00:00:00")
	assert.NoError(t, err)

	assert.Equal(t, exp, p.GetStart(now))
}

func Test_ExpensePeriodMonth_ShouldReturnCorrectStartDate(t *testing.T) {
	p := Month

	exp, err := time.Parse("2006-01-02 15:04:05", "2022-09-01 00:00:00")
	assert.NoError(t, err)

	now, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 00:00:00")
	assert.NoError(t, err)

	start := p.GetStart(now)

	assert.Equal(t, exp, start)
}

func Test_ExpensePeriodYear_ShouldReturnCorrectStartDate(t *testing.T) {
	p := Year

	exp, err := time.Parse("2006-01-02 15:04:05", "2021-10-01 00:00:00")
	assert.NoError(t, err)

	now, err := time.Parse("2006-01-02 15:04:05", "2022-10-01 00:00:00")
	assert.NoError(t, err)

	start := p.GetStart(now)

	assert.Equal(t, exp, start)
}
