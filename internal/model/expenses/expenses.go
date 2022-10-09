package expenses

import (
	"time"
)

type Expense struct {
	ID         string
	Amount     int // копейки
	Category   string
	CategoryID string
	Datetime   time.Time
}

type ExpenseCategory struct {
	ID   string
	Name string
}

type ExpensePeriod int64

const (
	Week ExpensePeriod = iota
	Month
	Year
)

func (p *ExpensePeriod) String() string {
	switch *p {
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

func (p *ExpensePeriod) GetStart(time time.Time) time.Time {
	switch *p {
	default:
		return time.AddDate(0, 0, -7)
	case Week:
		return time.AddDate(0, 0, -7)
	case Month:
		return time.AddDate(0, -1, 0)
	case Year:
		return time.AddDate(-1, 0, 0)
	}
}
