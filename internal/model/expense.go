package model

import "time"

type Expense struct {
	ID         string
	Amount     int64 // копейки
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

const (
	weekly  = "Недельный"
	monthly = "Месячный"
	annual  = "Годовой"
)

func (p *ExpensePeriod) String() string {
	switch *p {
	default:
		return weekly
	case Week:
		return weekly
	case Month:
		return monthly
	case Year:
		return annual
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
