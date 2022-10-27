# Budgetmeter Telegram Bot
Команды бота:
- `addExpenseCommand` - добавить трату. Пример: `/addExpense 10;Дом;2022-10-04 10:00:00`
- `getExpensesCommand` - получить список трат за неделю, месяц и год. Пример: `/getExpenses week`"
- `requestCurrencyChangeCommand` - вызвать меню смены валюты"
- `setCurrencyCommand` - установить валюту ввода и отображения отчетов. Пример: `/setCurrency EUR`
- `setLimitCommand` - установить лимит трат на категорию. Пример: `/setLimit Ремонт 1200.50`

## Logs
- STDOUT
- папка logs
- Graylog: http://127.0.0.1:7555/ (admin/admin)

## Metrics
Prometheus: http://127.0.0.1:9090/
Grafana: http://127.0.0.1:3000/ (admin/admin)

## Tracing
Jaeger: http://127.0.0.1:16686/

# Development
`make up-dev`/`make down-dev` поднимает/выключить локальное окружение для разработки и отладки
`make run` запускает бота
`make run-seeder` запускает сидер для базы данных

## Pre commit
`make migrate` запуск миграций
`make lint` запуск линтера
`make format` запуск форматирования импортов
`make tests` запуск юнит и функциональных тестов

`make` запускает все вышеперечисленное

# Testing
`make tests` запуск юнит и функциональных тестов
`make integration-tests` запуск интеграционных тестов