package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	// init pgsql.
	_ "github.com/jackc/pgx/stdlib"
	"gitlab.ozon.dev/cranky4/tg-bot/internal/utils/seeder"
)

var dsn string

func init() {
	flag.StringVar(&dsn, "dsn", "postgres://tg_bot_user:secret@localhost:5432/tg_bot", "Database DSN")
}

func main() {
	seeder := seeder.NewSeeder(dsn)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	if err := seeder.SeedExpenses(ctx, 100, 10); err != nil {
		cancel()
		log.Fatal(err)
	}

	log.Println("done")
}
