package expenses

import (
	"time"
)

type Expense struct {
	Amount   int // копейки
	Category string
	Datetime time.Time
}

type ExpensePeriod int64

const (
	Week ExpensePeriod = iota
	Month
	Year
)

func (p ExpensePeriod) String() string {
	switch p {
	default:
		return "Недельный"
	case Week:
		return "Недельный"
	case Month:
		return "Месячный"
	case Year:
		return "Годовой"
	}
}

func (p ExpensePeriod) GetStart(now time.Time) time.Time {
	switch p {
	default:
		return now.AddDate(0, 0, -7)
	case Week:
		return now.AddDate(0, 0, -7)
	case Month:
		return now.AddDate(0, -1, 0)
	case Year:
		return now.AddDate(-1, 0, 0)
	}
}
