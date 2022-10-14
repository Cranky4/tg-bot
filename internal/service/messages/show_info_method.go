package servicemessages

import "strings"

func (m *Model) showInfo() string {
	return strings.Join([]string{
		"Привет, я буду считать твои деньги. Вот что я умею:\n",
		addExpenseCommand,
		"- добавить трату\nПример: /addExpense 10;Дом;2022-10-04 10:00:00\n",
		getExpensesCommand,
		" - получить список трат за неделю, месяц и год\nПример: /getExpenses week\n",
		requestCurrencyChangeCommand,
		" - вызвать менюсмены валюты\n",
		setCurrencyCommand,
		" - установить валюту ввода и отображения отчетов.\nПример: /setCurrency EUR\n",
		setLimitCommand,
		" - установить лимит трат на категорию.\nПример: /setLimit Ремонт 1200.50\n",
	}, "")
}
